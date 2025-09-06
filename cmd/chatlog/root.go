package chatlog

import (
	"github.com/sjzar/chatlog/internal/chatlog"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	// windows only
	cobra.MousetrapHelpText = ""

	rootCmd.PersistentFlags().BoolVar(&Debug, "debug", false, "debug")
}

func Execute() {
	// Initialize the logger here to ensure it's always called
	// before any command execution.
	cobra.OnInitialize(InitLog)

	if err := rootCmd.Execute(); err != nil {
		log.Err(err).Msg("command execution failed")
	}
}

var rootCmd = &cobra.Command{
	Use:     "chatlog",
	Short:   "chatlog",
	Long:    `chatlog`,
	Example: `chatlog`,
	Args:    cobra.MinimumNArgs(0),
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
	Run: Root,
}

func Root(cmd *cobra.Command, args []string) {

	m, err := chatlog.New("")
	if err != nil {
		log.Err(err).Msg("failed to create chatlog instance")
		return
	}

	if err := m.Run(); err != nil {
		log.Err(err).Msg("failed to run chatlog instance")
	}
}
