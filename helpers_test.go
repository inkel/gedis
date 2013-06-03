package gedis

import (
	"testing"
	"runtime"
)

func assertStringEq(t *testing.T, skip int, expected string, actual interface{}) {
	_, file, ln, _ := runtime.Caller(skip)

	if value, ok := actual.(string); ok {
		if expected != value {
			t.Errorf("\r%s:%d: Expected %q\nReturned %q", file, ln, expected, value)
			t.FailNow()
		}
	} else {
		t.Errorf("\r%s:%d: Cannot convert to string: %#v", file, ln, actual)
		t.FailNow()
	}
}

func assertIntegerEq(t *testing.T, skip int, expected int, actual interface{}) {
	_, file, ln, _ := runtime.Caller(skip)

	if value, ok := actual.(int); ok {
		if expected != value {
			t.Errorf("\r%s:%d: Expected %q\nReturned %q", file, ln, expected, value)
			t.FailNow()
		}
	} else {
		t.Errorf("\r%s:d: Cannot convert to int: %#v", file, ln, actual)
		t.FailNow()
	}
}

func assertNotError(t *testing.T, skip int, err error) {
	if err != nil {
		_, file, ln, _ := runtime.Caller(skip)

		t.Errorf("\r%s:%d: Returned unexpected error: %v", file, ln, err)
	}
}
