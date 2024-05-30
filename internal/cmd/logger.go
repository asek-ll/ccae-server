package cmd

import (
	"fmt"
	"log"

	"github.com/asek-ll/aecc-server/pkg/logger"

	"github.com/fatih/color"
)

var logOptions = []logger.Option{
	logger.Format(fmt.Sprintf("%s [{{.Level}}] {{.Message}}",
		color.GreenString(`{{.Time.Format "2006-01-02T15:04:05"}}`),
	)),
	logger.LevelFormat(func(l logger.Level) string {
		if l.Value >= logger.ERROR.Value {
			return color.RedString(l.Padded())
		}
		if l.Value >= logger.WARN.Value {
			return color.HiRedString(l.Padded())
		}
		if l.Value >= logger.INFO.Value {
			return l.Padded()
		}
		return color.GreenString(l.Padded())
	}),
	logger.WithLevel(logger.INFO),
}

func setupLogger() *log.Logger {
	logger.SetupStd(logOptions...)
	l := logger.New(logOptions...)
	return l
}
