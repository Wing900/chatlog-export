package chatlog

import (
	"fmt"
	"runtime"

	"github.com/sjzar/chatlog/internal/chatlog"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(decryptCmd)
	decryptCmd.Flags().StringVarP(&dataDir, "data-dir", "d", "", "data dir")
	decryptCmd.Flags().StringVarP(&workDir, "work-dir", "w", "", "work dir")
	decryptCmd.Flags().StringVarP(&key, "key", "k", "", "key")
	decryptCmd.Flags().StringVarP(&decryptPlatform, "platform", "p", runtime.GOOS, "platform")
	decryptCmd.Flags().IntVarP(&decryptVer, "version", "v", 3, "version")
}

var (
	dataDir         string
	workDir         string
	key             string
	decryptPlatform string
	decryptVer      int
)

var decryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "decrypt",
	Run: func(cmd *cobra.Command, args []string) {
		m, err := chatlog.New("")
		if err != nil {
			log.Err(err).Msg("failed to create chatlog instance")
			return
		}
		if err := m.CommandDecrypt(dataDir, workDir, key, decryptPlatform, decryptVer); err != nil {
			log.Err(err).Msg("failed to decrypt")
			return
		}
		fmt.Println("decrypt success")
	},
}
