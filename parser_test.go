package gedis

import (
	"testing"
	"strings"
	"runtime"
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
