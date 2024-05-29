package logger

import (
	"html/template"
)

type Option func(l *Logger)

func Format(format string) Option {
	return func(l *Logger) {
		tmpl, err := template.New("logFmt").Parse(format)
		if err == nil {
			l.tmpl = tmpl
		}
	}
}

func LevelFormat(format func(l Level) string) Option {
	return func(l *Logger) {
		l.levelFormat = format
	}
}

func WithLevel(lvl Level) Option {
	return func(log *Logger) {
		log.lvl = lvl
	}
}
