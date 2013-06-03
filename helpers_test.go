package gedis

import (
	"fmt"
	"path"
	"runtime"
	"testing"
)

type Asserter struct {
	t    *testing.T
	skip int
}

func (a *Asserter) logf(format string, args ...interface{}) {
	_, file, ln, _ := runtime.Caller(a.skip + 1)
	a.t.Errorf("\r\t%s:%d: %s", path.Base(file), ln, fmt.Sprintf(format, args...))
}

func (a *Asserter) StringEq(expected string, actual interface{}) {
	if value, ok := actual.(string); ok {
		if expected != value {
			a.logf("\nexpected %q\nreturned %q", expected, value)
			a.t.FailNow()
		}
	} else {
		a.logf("cannot convert to string: %#v\n", actual)
		a.t.FailNow()
	}
}

func (a *Asserter) IntegerEq(expected int, actual interface{}) {
	if value, ok := actual.(int); ok {
		if expected != value {
			a.logf("\nexpected %#v\nreturned %#v", expected, value)
			a.t.FailNow()
		}
	} else {
		a.logf("cannot convert to int: %#v", actual)
		a.t.FailNow()
	}
}

func (a *Asserter) Nil(val interface{}) {
	if val != nil {
		a.logf("nil expected, got: %v", val)
		a.t.FailNow()
	}
}

func (a *Asserter) NotNil(val interface{}) {
	if val == nil {
		a.logf("got nil when non-nil was expected")
		a.t.FailNow()
	}
}
