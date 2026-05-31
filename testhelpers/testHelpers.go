package testhelpers

import (
	"bytes"
	"fmt"
	"regexp"
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

func AssertErrorEqual(t testing.TB, got, want error) {
	t.Helper()

	AssertErrorNotNil(t, got, want)
	if got.Error() != want.Error() {
		t.Fatalf("got %s, but want %s", got.Error(), want.Error())
	}
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
