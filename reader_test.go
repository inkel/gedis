package gedis

import (
	"testing"
	"strings"
	"runtime"
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
