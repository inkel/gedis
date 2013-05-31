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

	if !bytes.Equal(expected, data) {
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
