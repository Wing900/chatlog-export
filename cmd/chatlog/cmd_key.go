package chatlog

import (
	"fmt"

	"github.com/sjzar/chatlog/internal/chatlog"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(keyCmd)
	keyCmd.Flags().IntVarP(&pid, "pid", "p", 0, "pid")
}

var pid int
var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "key",
	Run: func(cmd *cobra.Command, args []string) {
		m, err := chatlog.New("")
		if err != nil {
			log.Err(err).Msg("failed to create chatlog instance")
			return
		}
		ret, err := m.CommandKey(pid)
		if err != nil {
			log.Err(err).Msg("failed to get key")
			return
		}
		fmt.Println(ret)
	},
}
