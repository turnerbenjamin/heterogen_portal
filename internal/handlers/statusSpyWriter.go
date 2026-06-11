// Package handlers contains HTTP handlers for the application.
//
// This file provides a response-writer helper used by the pipeline to capture
// status codes without immediately committing headers.
package handlers

import "net/http"

// statusSpyWriter captures the status code written by a handler and defers
// writing headers to the underlying ResponseWriter until the body is written.
type statusSpyWriter struct {
	http.ResponseWriter
	statusCode    int
	headerWritten bool
}

// WriteHeader records the status code but does not write headers to the
// underlying writer. Actual header write is performed on the first call to
// Write.
func (w *statusSpyWriter) WriteHeader(code int) {
	w.statusCode = code
}

// Write ensures a default 200 status is set, writes headers once, and then
// writes the response body to the underlying writer.
func (w *statusSpyWriter) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	if !w.headerWritten {
		w.ResponseWriter.WriteHeader(w.statusCode)
		w.headerWritten = true
	}
	return w.ResponseWriter.Write(b)
}
