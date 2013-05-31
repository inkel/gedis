package gedis

import (
	"testing"
	"strings"
	"runtime"
	"bytes"
)

func pass_readNumber(t *testing.T, line string, expected int) {
	_, file, ln, _ := runtime.Caller(1)

	n, err := readNumber(strings.NewReader(line))

	if err != nil {
		t.Logf("%s:%d: readNumber(%q) returned an error: %#v", file, ln, line, err)
		t.FailNow()
	}

	if n != expected {
		t.Logf("%s:%d: readNumber(%q) returned %d, expected %d", file, ln, line, n, expected)
		t.FailNow()
	}
}

func fail_readNumber(t *testing.T, line string) {
	_, file, ln, _ := runtime.Caller(1)
	_, err := readNumber(strings.NewReader(line))

	if err == nil {
		t.Errorf("%s:%s: readNumber(%q) didn't return an error", file, ln, line)
	}
}

func Test_readNumber(t *testing.T) {
	pass_readNumber(t, "1234\r\n", 1234)
	pass_readNumber(t, "-1234\r\n", -1234)
	fail_readNumber(t, "abc\r\n")
	fail_readNumber(t, "12ab34\r\n")
}

func Benchmark_readNumber(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		line := strings.NewReader("1234\r\n")
		b.StartTimer()
		readNumber(line)
	}
}

func pass_readLine(t *testing.T, line string) {
	_, file, ln, _ := runtime.Caller(1)

	expected := []byte(line)
	input := []byte(line + "\r\n")
	reader := strings.NewReader(string(input))

	res, err := readLine(reader)

	if err != nil {
		t.Errorf("%s:%d: readLine() returned an error: %v", file, ln, err)
		t.FailNow()
	}

	if !bytes.Equal(expected, res) {
		t.Errorf("%s:%d: readLine()\nreturned %#v\nexpected %#v", file, ln, res, expected)
		t.FailNow()
	}
}

func Test_readLine(t *testing.T) {
	pass_readLine(t, "Lorem ipsum dolor sit amet")
	pass_readLine(t, "Lorem\ripsum")
}

func Benchmark_readLine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		line := strings.NewReader("Lorem ipsum\rdolor sit amet\r\n")
		b.StartTimer()
		readLine(line)
	}
}
