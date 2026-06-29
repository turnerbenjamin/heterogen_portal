package testhelpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"
)

func AssertErrorNil(t testing.TB, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("got %s, but want nil", err.Error())
	}
}

func AssertErrorNotNil(t testing.TB, got, want error) {
	t.Helper()

	if got == nil {
		t.Fatalf("got nil, but want %s", want.Error())
	}
}

func AssertNotNil(t testing.TB, got, want any) {
	t.Helper()

	if got == nil {
		t.Fatalf("got nil, but want %v", want)
	}
}

func AssertEqual(t testing.TB, got, want any) {
	t.Helper()

	AssertNotNil(t, got, want)
	if got != want {
		t.Fatalf("got %v, but want %v", got, want)
	}
}

func AssertErrorEqual(t testing.TB, got, want error) {
	t.Helper()

	AssertErrorNotNil(t, got, want)
	if got.Error() != want.Error() {
		t.Fatalf("got %s, but want %s", got.Error(), want.Error())
	}
}

func AssertStringSliceEqual(t testing.TB, got, want []string) {
	if got != nil && want == nil {
		t.Fatalf("got %s, but want nil", strings.Join(got, ", "))
	}

	if got == nil && want != nil {
		t.Fatalf("got nil, but want %s", strings.Join(want, ", "))
	}

	if got == nil && want == nil {
		return
	}

	gotString := strings.Join(got, ", ")
	wantString := strings.Join(want, ", ")
	AssertStringEqual(t, gotString, wantString)
}

func AssertIntEqual(t testing.TB, got, want int) {
	t.Helper()

	if got != want {
		t.Fatalf("got %d, but want %d", got, want)
	}
}

func AssertStringEqual(t testing.TB, got, want string) {
	t.Helper()

	if got != want {
		t.Fatalf("got %s, but want %s", got, want)
	}
}

func AssertStringMatches(t testing.TB, got, wantPattern string) {
	t.Helper()
	want := fmt.Sprintf("string message matching pattern (%s)", wantPattern)

	doMatch, err := regexp.Match(wantPattern, []byte(got))
	if err != nil {
		t.Fatalf("unable to match string: %s", err.Error())
	}

	if !doMatch {
		t.Fatalf("got %s, but want %s", got, want)
	}

}

func AssertCookieEqual(t testing.TB, got, want *http.Cookie) {
	t.Helper()

	AssertNotNil(t, got, want)
	AssertStringEqual(t, got.Value, want.Value)
	AssertStringEqual(t, got.Path, want.Path)
	AssertEqual(t, got.SameSite, want.SameSite)
	AssertIntEqual(t, got.MaxAge, want.MaxAge)
	AssertEqual(t, got.Expires, want.Expires)
	AssertEqual(t, got.Secure, want.Secure)
	AssertEqual(t, got.Partitioned, want.Partitioned)
	AssertEqual(t, got.HttpOnly, want.HttpOnly)
}

func AssertBytesEqual(t testing.TB, got, want []byte) {
	t.Helper()
	if !bytes.Equal(got, want) {
		t.Fatalf("got %s, but want %s", got, want)
	}
}

func AssertErrorMessageMatches(t testing.TB, got error, wantPattern string) {
	t.Helper()

	want := fmt.Errorf("error message matching pattern (%s)", wantPattern)
	AssertErrorNotNil(t, got, want)

	doMatch, err := regexp.Match(wantPattern, []byte(got.Error()))
	if err != nil {
		t.Fatalf("unable to match error message: %s", err.Error())
	}

	if !doMatch {
		t.Fatalf("got %s, but want %s", got.Error(), want.Error())
	}
}

func AssertSlogsContain(t testing.TB, logs []byte, attributes map[string]any) {
	t.Helper()

	logAttributes := map[string]any{}
	err := json.Unmarshal(logs, &logAttributes)

	if err != nil {
		t.Fatalf("unable to parse logs as map: %s", err.Error())
	}

	for k, want := range attributes {
		got, ok := logAttributes[k]
		if !ok {
			t.Fatalf("got none but want log attributes to contain key %s", k)
		}

		if !compareSlogValues(got, want) {
			t.Fatalf("for key %s, got %v but want %v", k, got, want)
		}
	}
}

type SlogValue string

// AnySlogValue may be used in AssertSlogsContain to skip value comparison for a
// given key
const AnySlogValue SlogValue = "**_any_slog_value_**"

func compareSlogValues(got, want any) bool {
	if want == AnySlogValue {
		return true
	}

	if got == want {
		return true
	}

	toFloat64 := func(v any) (float64, bool) {
		switch n := v.(type) {
		case int:
			return float64(n), true
		case int8:
			return float64(n), true
		case int16:
			return float64(n), true
		case int32:
			return float64(n), true
		case int64:
			return float64(n), true
		case uint:
			return float64(n), true
		case uint8:
			return float64(n), true
		case uint16:
			return float64(n), true
		case uint32:
			return float64(n), true
		case uint64:
			return float64(n), true
		case float32:
			return float64(n), true
		case float64:
			return n, true
		default:
			return 0, false
		}
	}

	gotF, gotIsNumber := toFloat64(got)
	wantF, wantIsNumber := toFloat64(want)

	if gotIsNumber && wantIsNumber {
		return gotF == wantF
	}

	return false
}
