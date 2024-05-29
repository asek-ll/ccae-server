package logger

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"
	"time"
)

type Level struct {
	value int16
	name  [5]byte
}

func (l Level) Trim() string {
	if l.name[4] == ' ' {
		return string(l.name[:4])
	}
	return string(l.name[:])
}

func (l Level) Padded() string {
	return string(l.name[:])
}

var (
	TRACE Level = Level{00, [5]byte{'T', 'R', 'A', 'C', 'E'}}
	DEBUG Level = Level{10, [5]byte{'D', 'E', 'B', 'U', 'G'}}
	INFO  Level = Level{20, [5]byte{'I', 'N', 'F', 'O', ' '}}
	WARN  Level = Level{30, [5]byte{'W', 'A', 'R', 'N', ' '}}
	ERROR Level = Level{40, [5]byte{'E', 'R', 'R', 'O', 'R'}}
)

const (
	DefaultFormat = `{{.Time.Format "2006-01-02 15:04:05"}} {{.Level.Padded}} {{.Message}}`
)

type Logger struct {
	lvl    Level
	stdout io.Writer
	templ  *template.Template
}
type templateParams struct {
	Time    time.Time
	Level   Level
	Message string
}

func New() *Logger {
	log := Logger{
		lvl:    INFO,
		stdout: os.Stdout,
		templ:  template.Must(template.New("log").Parse(DefaultFormat)),
	}

	return &log
}

func (log *Logger) Logf(lvl Level, format string, args ...any) {
	if lvl.value < log.lvl.value {
		fmt.Println(lvl)
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
		Level:   lvl,
		Message: strings.TrimSuffix(msg, "\n"),
	}

	buf := bytes.Buffer{}
	err := log.templ.Execute(&buf, params)
	if err != nil {
		fmt.Printf("failed to execute template, %v\n", err) // should never happen
	}
	buf.WriteByte('\n')
	data := buf.Bytes()
	_, _ = log.stdout.Write(data)
}
