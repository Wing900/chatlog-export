package darwin

import (
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v4/process"

	"github.com/sjzar/chatlog/internal/errors"
	"github.com/sjzar/chatlog/internal/wechat/model"
	"github.com/sjzar/chatlog/pkg/appver"
)

const (
	ProcessNameOfficial = "WeChat"
	ProcessNameBeta     = "Weixin"
	V3DBFile            = "Message/msg_0.db"
	V4DBFile            = "db_storage/session/session.db"
)

// Detector 实现 macOS 平台的进程检测器
type Detector struct{}

// NewDetector 创建一个新的 macOS 检测器
func NewDetector() *Detector {
	return &Detector{}
}

// FindProcesses 查找所有微信进程并返回它们的信息
func (d *Detector) FindProcesses() ([]*model.Process, error) {
	processes, err := process.Processes()
	if err != nil {
		log.Err(err).Msg("获取进程列表失败")
		return nil, err
	}

	var result []*model.Process
	for _, p := range processes {
		name, err := p.Name()
		if err != nil || (name != ProcessNameOfficial && name != ProcessNameBeta) {
			continue
		}

		// 获取进程信息
		procInfo, err := d.getProcessInfo(p)
		if err != nil {
			log.Err(err).Msgf("获取进程 %d 的信息失败", p.Pid)
			continue
		}

		result = append(result, procInfo)
	}

	return result, nil
}

// getProcessInfo 获取微信进程的详细信息
func (d *Detector) getProcessInfo(p *process.Process) (*model.Process, error) {
	procInfo := &model.Process{
		PID:      uint32(p.Pid),
		Status:   model.StatusOffline,
		Platform: model.PlatformMacOS,
	}

	// 获取可执行文件路径
	exePath, err := p.Exe()
	if err != nil {
		log.Err(err).Msg("获取可执行文件路径失败")
		return nil, err
	}
	procInfo.ExePath = exePath

	// 获取版本信息
	// 注意：macOS 的版本获取方式可能与 Windows 不同
	versionInfo, err := appver.New(exePath)
	if err != nil {
		log.Err(err).Msg("获取版本信息失败")
		procInfo.Version = 3
		procInfo.FullVersion = "3.0.0"
	} else {
		procInfo.Version = versionInfo.Version
		procInfo.FullVersion = versionInfo.FullVersion
	}

	// 初始化附加信息（数据目录、账户名）
	if err := d.initializeProcessInfo(p, procInfo); err != nil {
		log.Err(err).Msg("初始化进程信息失败")
		// 即使初始化失败也返回部分信息
	}

	return procInfo, nil
}

// initializeProcessInfo 获取进程的数据目录和账户名
func (d *Detector) initializeProcessInfo(p *process.Process, info *model.Process) error {
	// 使用 lsof 命令获取进程打开的文件
	files, err := d.getOpenFiles(int(p.Pid))
	if err != nil {
		log.Err(err).Msg("获取打开的文件失败")
		return err
	}

	dbPath := V3DBFile
	if info.Version == 4 {
		dbPath = V4DBFile
	}

	for _, filePath := range files {
		if strings.Contains(filePath, dbPath) {
			parts := strings.Split(filePath, string(filepath.Separator))
			if len(parts) < 4 {
				log.Debug().Msg("无效的文件路径格式: " + filePath)
				continue
			}

			// v3:
			// /Users/sarv/Library/Containers/com.tencent.xinWeChat/Data/Library/Application Support/com.tencent.xinWeChat/2.0b4.0.9/<id>/Message/msg_0.db
			// v4:
			// /Users/sarv/Library/Containers/com.tencent.xWeChat/Data/Documents/xwechat_files/<id>/db_storage/message/message_0.db

			info.Status = model.StatusOnline
			if info.Version == 4 {
				info.DataDir = strings.Join(parts[:len(parts)-3], string(filepath.Separator))
				info.AccountName = parts[len(parts)-4]
			} else {
				info.DataDir = strings.Join(parts[:len(parts)-2], string(filepath.Separator))
				info.AccountName = parts[len(parts)-3]
			}
			return nil
		}
	}

	return nil
}

// getOpenFiles 使用 lsof 命令获取进程打开的文件列表
func (d *Detector) getOpenFiles(pid int) ([]string, error) {
	// 执行 lsof -p <pid> 命令，使用 -F n 选项只输出文件名
	cmd := exec.Command("lsof", "-p", strconv.Itoa(pid), "-F", "n")
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.RunCmdFailed(err)
	}

	// 解析 lsof -F n 输出
	// 格式为: n/path/to/file
	lines := strings.Split(string(output), "\n")
	var files []string

	for _, line := range lines {
		if strings.HasPrefix(line, "n") {
			// 移除前缀 'n' 获取文件路径
			filePath := line[1:]
			if filePath != "" {
				files = append(files, filePath)
			}
		}
	}

	return files, nil
}
