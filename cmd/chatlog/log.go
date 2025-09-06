package chatlog

import (
	"github.com/rs/zerolog"
)

var Debug bool

// InitLog initializes the global logger.
// It must be called after flag parsing.
func InitLog() {
	zerolog.SetGlobalLevel(zerolog.Disabled)
}
