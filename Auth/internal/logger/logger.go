package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/rs/zerolog"
)

func New() zerolog.Logger {
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
		FormatLevel: func(i interface{}) string {
			if i == nil {
				return ""
			}

			level := strings.ToUpper(fmt.Sprintf("%s", i))
			switch level {
			case "INFO":
				return color.New(color.FgGreen).Sprint("INF")
			case "WARN":
				return color.New(color.FgYellow).Sprint("WRN")
			case "ERROR":
				return color.New(color.FgRed).Sprint("ERR")
			case "DEBUG":
				return color.New(color.FgBlue).Sprint("DBG")
			default:
				return level
			}
		},
	}

	log := zerolog.New(writer).With().Timestamp().Logger()

	log.Info().Msg("Logger loaded")
	return log
}
