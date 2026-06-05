package handlers

import "net/http"

type statusSpyWriter struct {
	http.ResponseWriter
	statusCode    int
	headerWritten bool
}

func (w *statusSpyWriter) WriteHeader(code int) {
	w.statusCode = code
}

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
