package glance

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sjzar/chatlog/internal/errors"
)

// FIXME 按照 region 读取效率较低，512MB 内存读取耗时约 18s

type Glance struct {
	PID        uint32
	MemRegions []MemRegion
	pipePath   string
	data       []byte
}

func NewGlance(pid uint32) *Glance {
	return &Glance{
		PID:      pid,
		pipePath: filepath.Join(os.TempDir(), fmt.Sprintf("chatlog_pipe_%d", time.Now().UnixNano())),
	}
}

func (g *Glance) Read() ([]byte, error) {
	if g.data != nil {
		return g.data, nil
	}

	regions, err := GetVmmap(g.PID)
	if err != nil {
		return nil, err
	}
	g.MemRegions = MemRegionsFilter(regions)

	if len(g.MemRegions) == 0 {
		return nil, errors.ErrNoMemoryRegionsFound
	}

	region := g.MemRegions[0]

	// 1. Create pipe file
	if err := exec.Command("mkfifo", g.pipePath).Run(); err != nil {
		return nil, errors.CreatePipeFileFailed(err)
	}
	defer os.Remove(g.pipePath)

	// Start a goroutine to read from the pipe
	dataCh := make(chan []byte, 1)
	errCh := make(chan error, 1)
	go func() {
		// Open pipe for reading
		file, err := os.OpenFile(g.pipePath, os.O_RDONLY, 0600)
		if err != nil {
			errCh <- errors.OpenPipeFileFailed(err)
			return
		}
		defer file.Close()

		// Read all data from pipe
		data, err := io.ReadAll(file)
		if err != nil {
			errCh <- errors.ReadPipeFileFailed(err)
			return
		}
		dataCh <- data
	}()

	// 2 & 3. Execute lldb command to read memory directly with all parameters
	size := region.End - region.Start
	lldbCmd := fmt.Sprintf("lldb -p %d -o \"memory read --binary --force --outfile %s --count %d 0x%x\" -o \"quit\"",
		g.PID, g.pipePath, size, region.Start)

	cmd := exec.Command("bash", "-c", lldbCmd)

	// Set up stdout pipe for monitoring (optional)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, errors.RunCmdFailed(err)
	}

	// Monitor lldb output (optional)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			// Uncomment for debugging:
			// fmt.Println(scanner.Text())
		}
	}()

	// Wait for data with timeout
	select {
	case data := <-dataCh:
		g.data = data
	case err := <-errCh:
		return nil, errors.ReadMemoryFailed(err)
	case <-time.After(30 * time.Second):
		cmd.Process.Kill()
		return nil, errors.ErrReadMemoryTimeout
	}

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		// We already have the data, so just log the error
		log.Err(err).Msg("lldb process exited with error")
	}

	return g.data, nil
}
