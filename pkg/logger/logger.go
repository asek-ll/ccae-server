package logger

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

type Level struct {
	Value int16
	Name  [5]byte
}

func (l Level) Trim() string {
	if l.Name[4] == ' ' {
		return string(l.Name[:4])
	}
	return string(l.Name[:])
}

func (l Level) Padded() string {
	return string(l.Name[:])
}

func (l Level) Braced() string {
	var buf [7]byte
	buf[0] = '['
	copy(buf[1:], l.Name[:])
	if buf[5] == ' ' {
		buf[5] = ']'
		buf[6] = ' '
	} else {
		buf[6] = ']'
	}
	return string(buf[:])
}

var (
	TRACE Level = Level{00, [5]byte{'T', 'R', 'A', 'C', 'E'}}
	DEBUG Level = Level{10, [5]byte{'D', 'E', 'B', 'U', 'G'}}
	INFO  Level = Level{20, [5]byte{'I', 'N', 'F', 'O', ' '}}
	WARN  Level = Level{30, [5]byte{'W', 'A', 'R', 'N', ' '}}
	ERROR Level = Level{40, [5]byte{'E', 'R', 'R', 'O', 'R'}}
)

const (
	DefaultFormat = `{{.Time.Format "2006-01-02 15:04:05"}} {{.Level}} {{.Message}}`
)

type Log interface {
	Logf(lvl Level, format string, args ...any)
}

type writer struct {
	Log
}

func (w *writer) Write(p []byte) (n int, err error) {
	w.Logf(INFO, string(p))
	return len(p), nil
}

func ToWriter(l Log) io.Writer {
	return &writer{l}
}

func SetupStd(l Log) {
	log.SetOutput(ToWriter(l))
	log.SetPrefix("")
	log.SetFlags(0)
}

type levelFormatFunc func(l Level) string

var nop levelFormatFunc = func(l Level) string { return l.Padded() }

type Logger struct {
	lvl         Level
	stdout      io.Writer
	tmpl        *template.Template
	levelFormat levelFormatFunc
}
type templateParams struct {
	Time    time.Time
	Level   string
	Message string
}

func New(opts ...Option) *Logger {
	log := &Logger{
		lvl:    INFO,
		stdout: os.Stdout,
	}
	for _, opt := range opts {
		opt(log)
	}
	if log.tmpl == nil {
		log.tmpl = template.Must(template.New("log").Parse(DefaultFormat))
	}

	if log.levelFormat == nil {
		log.levelFormat = nop
	}

	return log
}

func (log *Logger) Logf(lvl Level, format string, args ...any) {
	if lvl.Value < log.lvl.Value {
		return
	}

	var msg string
	if len(args) == 0 {
		msg = format
	} else {
		msg = fmt.Sprintf(format, args...)
	}

	params := templateParams{
		Time:    time.Now(),
		Level:   log.levelFormat(lvl),
		Message: strings.TrimSuffix(msg, "\n"),
	}

	buf := bytes.Buffer{}
	err := log.tmpl.Execute(&buf, params)
	if err != nil {
		fmt.Printf("failed to execute template, %v\n", err) // should never happen
	}
	buf.WriteByte('\n')
	data := buf.Bytes()
	_, _ = log.stdout.Write(data)
}
