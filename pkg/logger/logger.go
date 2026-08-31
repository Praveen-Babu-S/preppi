package logger

import (
	"os"

	"github.com/rs/zerolog"
)

func New(level string) zerolog.Logger {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	switch level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}

	output := zerolog.ConsoleWriter{Out: os.Stdout}
	return zerolog.New(output).With().Timestamp().Logger()
}
