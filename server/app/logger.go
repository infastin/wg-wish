package app

import (
	"io"
	"os"

	"github.com/rs/zerolog"
)

type LoggerParams struct {
	Level             zerolog.Level
	AdditionalWriters []io.Writer
}

func NewLogger(params *LoggerParams) (zerolog.Logger, error) {
	writers := make([]io.Writer, 0, 1+len(params.AdditionalWriters))
	writers = append(writers, zerolog.ConsoleWriter{Out: os.Stdout}) //nolint:exhaustruct
	writers = append(writers, params.AdditionalWriters...)

	mw := zerolog.MultiLevelWriter(writers...)
	lg := zerolog.New(mw).Level(params.Level)
	lg = lg.With().Timestamp().Logger()

	return lg, nil
}
