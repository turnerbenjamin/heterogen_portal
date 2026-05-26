// Package logging provides a small structured logger used by the application
// for emitting simple key/value JSON logs for HTTP requests and related data.
//
// The logger is lightweight and intended for diagnostic output; callers supply
// an io.Writer where JSON objects are written.
package logging

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Logger is the public interface used by callers to add key/value pairs and
// write the resulting JSON log to the configured writer.
type Logger interface {
	// AddKV adds or overwrites a property key with the provided value.
	AddKV(k, v string)
	// WriteLog encodes the accumulated properties as JSON and writes them to
	// the configured writer. It returns any error encountered while writing.
	WriteLog() error
}

// logger is the concrete implementation of Logger. It holds a writer and a
// map of properties that will be marshalled to JSON when WriteLog is called.
// The type is unexported; callers interact with it via the `Logger` interface.
type logger struct {
	writer         io.Writer
	properties     map[string]string
	requestStarted time.Time

	// propertiesMutex protects access to `properties`.
	propertiesMutex sync.RWMutex
	// writerMutex serializes writes to the underlying writer from this
	// logger's WriteLog calls.
	writerMutex sync.Mutex
}

// Constants for property keys used in emitted logs.
const (
	LOGGER_KEY_LOGGER_ERROR_NIL_REQUEST  = "_LOGGER_ERROR_NIL_REQUEST"
	LOGGER_KEY_LOGGER_ERROR_RESERVED_KEY = "_LOGGER_ERROR_RESERVED_KEY"
	LOGGER_KEY_REQUEST_METHOD            = "_METHOD"
	LOGGER_KEY_REQUEST_PATH              = "_PATH"
	LOGGER_KEY_REQUEST_TIME              = "_TIME"
	LOGGER_KEY_REQUEST_DURATION          = "_DURATION"
)

// Error message constants used by the logger when producing error keys or
// returning errors to callers.
const (
	Err_RequestIsNil = "request is nil, data cannot be printed"
	Err_WriterIsNil  = "writer is nil, unable to write logs"
	Err_InvalidKey   = "keys must not be prefixed with an _"
)

// NewLogger creates a new Logger that writes JSON objects to `w` and
// populates initial properties from `r` when non-nil (method and path).
func NewLogger(w io.Writer, r *http.Request) Logger {
	properties := make(map[string]string)

	if r != nil {
		properties[LOGGER_KEY_REQUEST_METHOD] = r.Method
		properties[LOGGER_KEY_REQUEST_PATH] = r.URL.Path
	} else {
		properties[LOGGER_KEY_LOGGER_ERROR_NIL_REQUEST] = Err_RequestIsNil
	}

	properties[LOGGER_KEY_REQUEST_TIME] = time.Now().Format(time.RFC3339)

	return &logger{
		writer:         w,
		properties:     properties,
		requestStarted: time.Now(),
	}
}

// AddKV adds or updates a property on the logger. Keys prefixed with an
// underscore are reserved
func (l *logger) AddKV(key, value string) {
	if strings.HasPrefix(key, "_") {
		key = LOGGER_KEY_LOGGER_ERROR_RESERVED_KEY
		value = Err_InvalidKey
	}
	l.propertiesMutex.Lock()
	l.properties[key] = value
	l.propertiesMutex.Unlock()
}

// WriteLog marshals the current properties to JSON and writes them to the
// configured writer. Returns errors when writer is nil or writer.Write returns
// an error
func (l *logger) WriteLog() error {
	if l.writer == nil {
		return errors.New(Err_WriterIsNil)
	}

	duration := time.Since(l.requestStarted).String()

	l.propertiesMutex.Lock()
	l.writerMutex.Lock()
	defer l.propertiesMutex.Unlock()
	defer l.writerMutex.Unlock()

	l.properties[LOGGER_KEY_REQUEST_DURATION] = duration

	enc := json.NewEncoder(l.writer)
	enc.SetIndent("", "  ")

	return enc.Encode(l.properties)
}
