package logging

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Logger interface {
	AddKV(k, v string)
	WriteLog() error
}

type consoleLogger struct {
	writer         io.Writer
	properties     map[string]string
	requestStarted time.Time
}

func NewLogger(w io.Writer, r *http.Request) Logger {
	properties := make(map[string]string)

	properties["date/time"] = time.Now().Format(time.Stamp)
	properties["method"] = r.Method
	properties["path"] = r.URL.Path

	return &consoleLogger{
		writer:         w,
		properties:     properties,
		requestStarted: time.Now(),
	}
}

func (l *consoleLogger) AddKV(key, value string) {
	l.properties[key] = value
}

func (l *consoleLogger) WriteLog() error {
	l.properties["duration"] = time.Since(l.requestStarted).Abs().String()

	sb := strings.Builder{}

	_, err := sb.WriteString("{\n")
	if err != nil {
		return err
	}
	for k, v := range l.properties {
		_, err = sb.WriteString(fmt.Sprintf("\t%v: %v,\n", k, v))
		if err != nil {
			return err
		}
	}

	_, err = sb.WriteString("}\n")
	if err != nil {
		return err
	}

	_, err = l.writer.Write([]byte(sb.String()))
	return err
}
