package gedis

import (
	"fmt"
	"path"
	"runtime"
	"testing"
)

func e(t *testing.T, skip int, format string, args ...interface{}) {
	_, file, ln, _ := runtime.Caller(skip + 1)
	t.Errorf("\r\t%s:%d: %s", path.Base(file), ln, fmt.Sprintf(format, args...))
}

func assertStringEq(t *testing.T, skip int, expected string, actual interface{}) {
	if value, ok := actual.(string); ok {
		if expected != value {
			e(t, skip, "assertStringEq()\nExpected %q\rReturned %q", expected, value)
			t.FailNow()
		}
	} else {
		e(t, skip, "assertStringEq(): Cannot convert to string: %#v\n", actual)
		t.FailNow()
	}
}

func assertIntegerEq(t *testing.T, skip int, expected int, actual interface{}) {
	if value, ok := actual.(int); ok {
		if expected != value {
			e(t, skip, "assertIntegerEq()\nExpected %#v\nReturned %#v", expected, value)
			t.FailNow()
		}
	} else {
		e(t, skip, "assertIntegerEq(): Cannot convert to int: %#v", actual)
		t.FailNow()
	}
}

func assertNotError(t *testing.T, skip int, err error) {
	if err != nil {
		e(t, skip, "assertNotError(): unexpected error: %v", err)
		t.FailNow()
	}
}
