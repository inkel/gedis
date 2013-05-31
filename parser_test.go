package gedis

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func assertString(t *testing.T, input, expected string) {
	_, file, ln, _ := runtime.Caller(1)

	reader := strings.NewReader(input)

	data, err := Parse(reader)

	if err != nil {
		t.Errorf("%s:%d: returned an error: %v", file, ln, err)
		t.FailNow()
	}

	got, ok := data.(string)

	if !ok {
		t.Errorf("%s:%d: Cannot convert to string: %#v", file, ln, data)
		t.FailNow()
	}

	if got != expected {
		t.Errorf("%s:%d:\nreturned %q\nexpected %q", file, ln, data, expected)
	}
}

func Test_Parse_Status(t *testing.T) {
	assertString(t, "+OK\r\n", "OK")
	assertString(t, "+PING\r\n", "PING")
}

func Benchmark_Parse_Status(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		reader := strings.NewReader("+OK\r\n")
		b.StartTimer()
		Parse(reader)
	}
}

func Test_Parse_Error(t *testing.T) {
	_, err := Parse(strings.NewReader("-ERR unknown\r\n"))

	if err == nil {
		t.Errorf("Parsing an error reply didn't return an error")
		t.FailNow()
	}

	if err.Error() != "ERR unknown" {
		t.Errorf("Unexpected error: %v", err)
		t.FailNow()
	}
}

func assertInteger(t *testing.T, input string, expected int) {
	_, file, ln, _ := runtime.Caller(1)

	reader := strings.NewReader(input)

	ret, err := Parse(reader)

	if err != nil {
		t.Errorf("%s:%d: Parse(%#v) returned an error: %v", file, ln, []byte(input), err)
		t.FailNow()
	}

	got, ok := ret.(int)

	if !ok {
		t.Errorf("%s:%d: Parse(%q): Couldn't convert to int: %#v", file, ln, input, ret)
		t.FailNow()
	}

	if got != expected {
		t.Errorf("%s:%d: Parse(%#v)\nreturned %#v\nexpected %#v", file, ln, []byte(input), got, expected)
		t.FailNow()
	}
}

func Test_Parse_Integer(t *testing.T) {
	assertInteger(t, ":1234\r\n", 1234)
	assertInteger(t, ":-1234\r\n", -1234)
}

func Test_Parse_Bulk(t *testing.T) {
	assertString(t, "$5\r\nlorem\r\n", "lorem")
	assertString(t, "$12\r\nlorem\r\nipsum\r\n", "lorem\r\nipsum")
}

func assertMultiBulk(t *testing.T, input string, expect ...interface{}) {
	_, file, ln, _ := runtime.Caller(1)

	reader := strings.NewReader(input)

	ret, err := Parse(reader)

	if err != nil {
		t.Errorf("%s:%d: Parse(%#v) returned an error: %v", file, ln, []byte(input), err)
		t.FailNow()
	}

	if expect == nil && ret != nil {
		t.Errorf("%s:%d: nil expected, got %#v", file, ln, ret)
		t.FailNow()
	}

	if expect == nil {
		return
	}

	data, ok := ret.([]interface{})

	if !ok {
		t.Errorf("%s:%d: can't convert results to []interface{}: %#v", file, ln, ret)
		t.FailNow()
	}

	if len(data) != len(expect) {
		t.Errorf("%s:%d: mismatched number of results\nreturned %d\nexpected %d", file, ln, len(data), len(expect))
		t.FailNow()
	}

	compare(t, file, ln, data, expect...)
}

func compare(t *testing.T, file string, ln int, data []interface{}, expect ...interface{}) {
	ok := true

	for i, exp := range expect {
		switch exp.(type) {
		case string:
			if got, ok := data[i].(string); ok {
				if got != exp.(string) {
					t.Errorf("%s:%d:\nreturned %q\nexpected %q", file, ln, got, exp)
				}
			} else {
				t.Errorf("%s:%d: can't convert to string: %#v", file, ln, data[i])
			}
		case int:
			if got, ok := data[i].(int); ok {
				if got != exp {
					t.Errorf("%s:%d:\nreturned %q\nexpected %q", file, ln, got, exp)
				}
			} else {
				t.Errorf("%s:%d: can't convert to int: %#v", file, ln, data[i])
			}
		case error:
			if got, ok := data[i].(error); ok {
				if got.Error() != exp.(error).Error() {
					t.Errorf("%s:%d:\nreturned %q\nexpected %q", file, ln, got, exp)
				}
			} else {
				t.Errorf("%s:%d: can't convert to error: %#v", file, ln, data[i])
			}
		case nil:
			if data[i] != nil {
				t.Errorf("%s:%d: expecting nil, got: %#v", file, ln, data[i])
			}
		case []string:
			in := data[i].([]interface{})
			for j, e := range exp.([]string) {
				if in[j] != e {
					t.Errorf("%s:%d: at index %d\nreturned %q\nexpected %q", file, ln, i, in[j], e)
				}
			}
		default:
			t.Errorf("%s:%d: something happened %#v", file, ln, data[i])
		}

		if !ok {
			t.FailNow()
		}
	}
}

func Test_Parse_MultiBulk(t *testing.T) {
	assertMultiBulk(t, "*1\r\n+OK\r\n", "OK")
	assertMultiBulk(t, "*1\r\n:1234\r\n", 1234)
	assertMultiBulk(t, "*2\r\n+OK\r\n-ERR unknown\r\n", "OK", fmt.Errorf("ERR unknown"))
	assertMultiBulk(t, "*3\r\n:1234\r\n$5\r\nlorem\r\n$-1\r\n", 1234, "lorem", nil)
	assertMultiBulk(t, "*1\r\n*1\r\n$5\r\nlorem\r\n", []string{"lorem"})
	assertMultiBulk(t, "*-1")
}
