package logger

import (
	"html/template"
)

type Option func(l *logWriter)

func Format(format string) Option {
	return func(l *logWriter) {
		tmpl, err := template.New("logFmt").Parse(format)
		if err == nil {
			l.tmpl = tmpl
		}
	}
}

func LevelFormat(format func(l Level) string) Option {
	return func(l *logWriter) {
		l.levelFormat = format
	}
}

func WithLevel(lvl Level) Option {
	return func(log *logWriter) {
		log.lvl = lvl
	}
}
