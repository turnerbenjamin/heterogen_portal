package logging

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/turnerbenjamin/heterogen_portal/testhelpers"
)

type testWriter struct {
	t testing.TB
	e error
	w bytes.Buffer
}

func (tw *testWriter) Write(p []byte) (n int, err error) {
	tw.t.Helper()

	n, err = tw.w.Write(p)
	if err != nil {
		tw.t.Fatalf("writer failed with error %s", err.Error())
	}
	return n, tw.e
}

func TestNewLogger_HandlesNilWriter(t *testing.T) {
	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	l := NewLogger(nil, r)
	l.AddKV("k", "v")

	err := l.WriteLog()
	testhelpers.AssertErrorEqual(t, err, errors.New(Err_WriterIsNil))
}

func TestNewLogger_ReturnsNilResponseWithoutErrorOrPanic(t *testing.T) {
	w := &bytes.Buffer{}
	l := NewLogger(w, nil)

	testData := map[string]string{"testKey": "testValue"}
	for k, v := range testData {
		l.AddKV(k, v)
	}

	err := l.WriteLog()

	want := maps.Clone(testData)
	want[LOGGER_KEY_LOGGER_ERROR_NIL_REQUEST] = Err_RequestIsNil

	testhelpers.AssertErrorNil(t, err)
	assertLogHasProperties(t, w, want)
}

func TestNewLogger_ReturnsALoggerWithStandardProperties(t *testing.T) {
	testMethod := "PATCH"
	testPath := "/test-method-path"

	w := &bytes.Buffer{}
	r := httptest.NewRequest(testMethod, testPath, strings.NewReader(""))

	l := NewLogger(w, r)

	err := l.WriteLog()

	testhelpers.AssertErrorNil(t, err)
	assertLogHasProperties(t, w, map[string]string{
		LOGGER_KEY_REQUEST_PATH:     regexp.QuoteMeta(testPath),
		LOGGER_KEY_REQUEST_METHOD:   regexp.QuoteMeta(testMethod),
		LOGGER_KEY_REQUEST_TIME:     ".*",
		LOGGER_KEY_REQUEST_DURATION: ".*",
	})
}

func TestAddKV_AddsPropertiesToTheLog(t *testing.T) {
	testMethod := "PATCH"
	testPath := "/test-method-path"

	testLogs := map[string]string{
		"test_log_1": "a test value",
		"test_log_2": "another test value",
	}

	w := &bytes.Buffer{}
	r := httptest.NewRequest(testMethod, testPath, strings.NewReader(""))

	l := NewLogger(w, r)

	for k, v := range testLogs {
		l.AddKV(k, v)
	}
	err := l.WriteLog()

	testhelpers.AssertErrorNil(t, err)
	assertLogHasProperties(t, w, testLogs)
}

func TestAddKV_ShowsLogErrorForInvalidKeys(t *testing.T) {
	w := &bytes.Buffer{}
	r := httptest.NewRequest("POST", "/test", strings.NewReader(""))

	l := NewLogger(w, r)

	l.AddKV("_invalidKey", "value")
	err := l.WriteLog()

	testhelpers.AssertErrorNil(t, err)
	assertLogHasProperties(t, w, map[string]string{
		LOGGER_KEY_LOGGER_ERROR_RESERVED_KEY: Err_InvalidKey,
	})
}

func TestAddKV_HandlesEmptyInputs(t *testing.T) {
	testMethod := "PATCH"
	testPath := "/test-method-path"

	testLogs := map[string]string{
		"empty_value": "",
		"":            "empty_key",
	}

	w := &bytes.Buffer{}
	r := httptest.NewRequest(testMethod, testPath, strings.NewReader(""))

	l := NewLogger(w, r)

	for k, v := range testLogs {
		l.AddKV(k, v)
		testLogs[k] = regexp.QuoteMeta(v)
	}
	err := l.WriteLog()

	testhelpers.AssertErrorNil(t, err)
	assertLogHasProperties(t, w, testLogs)
}

func TestAddKV_OverridesValuesWhenConflictingKeys(t *testing.T) {
	testMethod := "PATCH"
	testPath := "/test-method-path"

	key := "test_key"
	values := []string{
		"value1",
		"value2",
		"value3",
	}

	w := &bytes.Buffer{}
	r := httptest.NewRequest(testMethod, testPath, strings.NewReader(""))

	l := NewLogger(w, r)

	lastValue := ""
	for _, v := range values {
		l.AddKV(key, v)
		lastValue = v
	}
	err := l.WriteLog()

	testhelpers.AssertErrorNil(t, err)
	assertLogHasProperties(t, w, map[string]string{
		key: regexp.QuoteMeta(lastValue),
	})
}

func TestWriteLog_ReturnsErrorsFromWriter(t *testing.T) {
	testMethod := "PATCH"
	testPath := "/test-method-path"

	want := errors.New("returns_errors_from_writer_err")
	w := testWriter{
		t: t,
		e: want,
		w: bytes.Buffer{},
	}
	r := httptest.NewRequest(testMethod, testPath, strings.NewReader(""))
	l := NewLogger(&w, r)

	err := l.WriteLog()
	testhelpers.AssertErrorEqual(t, err, want)
}

func TestConcurrency_AddKVAndWriteLog_NoRace(t *testing.T) {
	w := &bytes.Buffer{}
	r := httptest.NewRequest("GET", "/test", nil)
	l := NewLogger(w, r)

	const goroutines = 20
	const perGoroutine = 200

	// Add properties concurrently
	var wg sync.WaitGroup
	for i := range goroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := range perGoroutine {
				l.AddKV(fmt.Sprintf("k-%d-%d", id, j), "v")
			}
		}(i)
	}

	// Add writes while adding properties
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = l.WriteLog()
		}()
	}

	// Wait for all operations to finish
	wg.Wait()

	// Clear buffer and make a final write
	w.Reset()
	err := l.WriteLog()
	testhelpers.AssertErrorNil(t, err)

	// Decode the buffer
	var got map[string]string
	if err := json.Unmarshal(w.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// validate key count
	if len(got) < goroutines*perGoroutine {
		t.Fatalf(
			"got %d properties, but expected at least %d properties",
			goroutines*perGoroutine,
			len(got),
		)
	}

	// Validate all keys in the log
	for i := range goroutines {
		for j := range perGoroutine {
			k := fmt.Sprintf("k-%d-%d", i, j)
			v, ok := got[k]
			if !ok {
				t.Fatalf("missing key %s", k)
			}
			if v != "v" {
				t.Fatalf("expected v, but got %s", v)
			}
		}
	}
}

func assertLogHasProperties(t testing.TB, w *bytes.Buffer, wantProperties map[string]string) {
	t.Helper()

	gotProperties := map[string]string{}
	err := json.Unmarshal(w.Bytes(), &gotProperties)
	if err != nil {
		t.Fatalf("unmarshal failed with error %s", err.Error())
	}

	for k, want := range wantProperties {
		got, ok := gotProperties[k]
		if !ok {
			t.Fatalf("output does not include key %s", k)
		}
		match, err := regexp.Match(want, []byte(got))
		if err != nil {
			t.Fatalf("match failed with error %s", err.Error())
		}
		if !match {
			t.Fatalf("for key %s: got %s, but want %s", k, got, want)
		}
	}
}
