package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusSpyWriter_Write_CallsWriteOnUnderlyingWriter(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	ssw := statusSpyWriter{ResponseWriter: w}

	wantStatusCode := 418
	wantBody := []byte("test_content")
	wantWritten := len(wantBody)

	ssw.WriteHeader(wantStatusCode)
	gotWritten, err := ssw.Write(wantBody)
	require.Nil(t, err)

	res := w.Result()

	gotStatusCode := res.StatusCode

	assert.Nil(t, err)
	assert.Equal(t, wantStatusCode, gotStatusCode)
	assert.Equal(t, string(wantBody), w.Body.String())
	assert.Equal(t, wantWritten, gotWritten)
}

func TestStatusSpyWriter_Write_DefaultsStatusCodeTo200(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	ssw := statusSpyWriter{ResponseWriter: w}

	_, err := ssw.Write([]byte("body"))
	require.Nil(t, err)

	wantStatusCode := 200
	gotStatusCode := w.Result().StatusCode

	assert.Equal(t, wantStatusCode, gotStatusCode)
}

func TestStatusSpyWriter_WriteHeader_TracksFinalStatusCode(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	ssw := statusSpyWriter{ResponseWriter: w}

	ssw.WriteHeader(201)
	ssw.WriteHeader(500)
	ssw.WriteHeader(404)

	want := 401
	ssw.WriteHeader(want)

	_, err := ssw.Write([]byte("body"))
	require.Nil(t, err)

	gotTracked := ssw.statusCode
	gotWritten := w.Result().StatusCode

	assert.Equal(t, want, gotTracked)
	assert.Equal(t, want, gotWritten)
}
