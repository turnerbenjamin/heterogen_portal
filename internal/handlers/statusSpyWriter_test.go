package handlers

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/turnerbenjamin/heterogen_portal/testhelpers"
)

func TestStatusSpyWriter_Write_CallsWriteOnUnderlyingWriter(t *testing.T) {
	w := httptest.NewRecorder()
	ssw := statusSpyWriter{ResponseWriter: w}

	wantStatusCode := 418
	wantBody := []byte("test_content")
	wantWritten := len(wantBody)

	ssw.WriteHeader(wantStatusCode)
	gotWritten, err := ssw.Write(wantBody)
	testhelpers.AssertErrorNil(t, err)

	res := w.Result()

	gotStatusCode := res.StatusCode
	gotBody, err := io.ReadAll(res.Body)
	defer func() {
		err := res.Body.Close()
		if err != nil {
			t.Fatal(err.Error())
		}
	}()

	testhelpers.AssertErrorNil(t, err)
	testhelpers.AssertIntEqual(t, gotStatusCode, wantStatusCode)
	testhelpers.AssertBytesEqual(t, gotBody, wantBody)
	testhelpers.AssertIntEqual(t, gotWritten, wantWritten)
}

func TestStatusSpyWriter_Write_DefaultsStatusCodeTo200(t *testing.T) {
	w := httptest.NewRecorder()
	ssw := statusSpyWriter{ResponseWriter: w}

	_, err := ssw.Write([]byte("body"))
	testhelpers.AssertErrorNil(t, err)

	wantStatusCode := 200
	gotStatusCode := w.Result().StatusCode

	testhelpers.AssertIntEqual(t, gotStatusCode, wantStatusCode)
}

func TestStatusSpyWriter_WriteHeader_TracksFinalStatusCode(t *testing.T) {
	w := httptest.NewRecorder()
	ssw := statusSpyWriter{ResponseWriter: w}

	ssw.WriteHeader(201)
	ssw.WriteHeader(500)
	ssw.WriteHeader(404)

	want := 401
	ssw.WriteHeader(want)

	_, err := ssw.Write([]byte("body"))
	testhelpers.AssertErrorNil(t, err)

	gotTracked := ssw.statusCode
	gotWritten := w.Result().StatusCode

	testhelpers.AssertIntEqual(t, gotTracked, want)
	testhelpers.AssertIntEqual(t, gotWritten, want)
}
