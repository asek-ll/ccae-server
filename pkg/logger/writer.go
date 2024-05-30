package logger

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	DefaultFormat = `{{.Time.Format "2006-01-02 15:04:05"}} {{.Level}} {{.Message}}`
)

type levelFormatFunc func(l Level) string

var nop levelFormatFunc = func(l Level) string { return l.Padded() }
var reTrace = regexp.MustCompile(`.*/log/log\.go.*\n`)

type templateParams struct {
	Time    time.Time
	Level   string
	Message string
}

func extractLevel(p []byte) (Level, string) {
	i := 0
	hasBrace := true
	if i < len(p) && p[i] == '[' {
		i += 1
	}

	level := INFO

	for _, l := range levels {
		name := l.Name()
		if i+len(name) < len(p) && bytes.Equal(name, p[i:i+len(name)]) {
			level = l
			i += len(name)
			break
		}
	}
	if hasBrace && i < len(p) && p[i] == ']' {
		i += 1
	}
	for i < len(p) && p[i] == ' ' {
		i += 1
	}

	return level, string(p[i:])
}

type logWriter struct {
	lvl         Level
	stdout      io.Writer
	tmpl        *template.Template
	levelFormat levelFormatFunc
	lock        sync.Mutex
	errorDump   bool
}

func SetupStd(opts ...Option) {
	log.SetOutput(newWriter(opts...))
	log.SetPrefix("")
	log.SetFlags(0)
}

func New(opts ...Option) *log.Logger {
	return log.New(newWriter(opts...), "", 0)
}

func newWriter(opts ...Option) io.Writer {
	log := &logWriter{
		lvl:       INFO,
		stdout:    os.Stdout,
		errorDump: true,
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

func (w *logWriter) Write(p []byte) (int, error) {

	lvl, msg := extractLevel(p)

	if lvl.Value < w.lvl.Value {
		return 0, nil
	}

	params := templateParams{
		Time:    time.Now(),
		Level:   w.levelFormat(lvl),
		Message: strings.TrimSuffix(msg, "\n"),
	}

	buf := bytes.Buffer{}
	err := w.tmpl.Execute(&buf, params)
	if err != nil {
		fmt.Printf("failed to execute template, %v\n", err) // should never happen
	}
	buf.WriteByte('\n')
	data := buf.Bytes()

	if lvl.Value < ERROR.Value {
		return w.stdout.Write(data)
	}

	w.lock.Lock()
	wn, we := w.stdout.Write(data)

	if w.errorDump {
		stackInfo := make([]byte, 1024*1024)
		if stackSize := runtime.Stack(stackInfo, false); stackSize > 0 {
			traceLines := reTrace.Split(string(stackInfo[:stackSize]), -1)
			if len(traceLines) > 0 {
				_, _ = w.stdout.Write([]byte(">>> stack trace:\n" + traceLines[len(traceLines)-1]))
			}
		}
	}
	w.lock.Unlock()

	return wn, we
}
