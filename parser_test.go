package gedis

import (
	"testing"
	"bytes"
	"strings"
	"runtime"
)

func assertParse(t *testing.T, input, output string) {
	_, file, ln, _ := runtime.Caller(1)

	reader := strings.NewReader(input)
	expected := []byte(output)

	data, err := Parse(reader)

	if err != nil {
		t.Errorf("%s:%d: Parse(%#v) returned an error: %v", file, ln, []byte(input), err)
		t.FailNow()
	}

	got, ok := data.([]byte)

	if !ok {
		t.Errorf("%s:%d: Cannot convert to []byte: %#v", data)
		t.FailNow()
	}

	if !bytes.Equal(expected, got) {
		t.Errorf("%s:%d: Parse(%#v)\nreturned %#v\nexpected %#v", file, ln, []byte(input), data, expected)
	}
}

func Test_Parse_Status(t *testing.T) {
	assertParse(t, "+OK\r\n", "OK")
	assertParse(t, "+PING\r\n", "PING")
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
