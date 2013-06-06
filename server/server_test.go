package server

import (
	"bytes"
	"path"
	"runtime"
	"testing"
)

func fail_Read(t *testing.T, input string) {
	_, file, ln, _ := runtime.Caller(1)
	file = path.Base(file)

	reader := bytes.NewBufferString(input)

	_, err := Read(reader)

	if err == nil {
		t.Fatalf("\r\t%s:%d: FAIL error expected", file, ln)
	} else {
		t.Logf("\r\t%s:%d: PASS (%v)", file, ln, err)
	}
}

func TestRead_errors(t *testing.T) {
	fail_Read(t, "lorem ipsum")
	fail_Read(t, "+OK")
	fail_Read(t, "-ERR lorem ipsum")
	fail_Read(t, ":123")
	fail_Read(t, "*1")
	fail_Read(t, "*1\r\n")
	fail_Read(t, "*1\r\n$5lorem")
	fail_Read(t, "*1\r\n$5\r\nlorem")
	fail_Read(t, "*a\r\n$5\r\nlorem\r\n")
	fail_Read(t, "*1\r\n$b\r\nlorem\r\n")
	// fail_Read(t, "*1\r\n$5\r\nlorem\r\n$-1\r\n")
	fail_Read(t, "*2\r\n$5\r\nlorem\r\n:1234\r\n")
	// fail_Read(t, "*1\r\n$5\r\nlorem\r\n$5\r\nipsum\r\n")
}

func pass_Read(t *testing.T, input string, expected ...[]byte) {
	_, file, ln, _ := runtime.Caller(1)
	file = path.Base(file)

	reader := bytes.NewBufferString(input)

	res, err := Read(reader)

	if err != nil {
		t.Fatalf("\r\t%s:%d: unexpected error: %v", file, ln, err)
	}

	if len(expected) != len(res) {
		t.Fatalf("\r\t%s:%d: expected %d results, got %d", file, ln, len(expected), len(res))
	}

	for i, exp := range expected {
		if !bytes.Equal(exp, res[i]) {
			t.Fatal("\r\t%s:%d: at index %d\nexpected %#v\ngot      %#v", file, ln, i, exp, res[i])
		}
	}
}

func TestRead_success(t *testing.T) {
	pass_Read(t, "*1\r\n$5\r\nlorem\r\n", []byte("lorem"))
	pass_Read(t, "*2\r\n$5\r\nlorem\r\n$5\r\nipsum\r\n", []byte("lorem"), []byte("ipsum"))
	pass_Read(t, "*1\r\n$12\r\nlorem\r\nipsum\r\n", []byte("lorem\r\nipsum"))
}

func TestRead(t *testing.T) {
	reader := bytes.NewBufferString("*1\r\n$5\r\nlorem\r\n$5\r\nipsum\r\n")

	res, err := Read(reader)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(res[0], []byte("lorem")) {
		t.Fatalf("unexpected response: %q", res)
	}

	res, err = Read(reader)

	if err == nil {
		t.Fatal("nil when expecting error")
	}
}

func Benchmark_Read(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		reader := bytes.NewBufferString("*2\r\n$5\r\nlorem\r\n$5\r\nipsum\r\n")
		b.StartTimer()
		Read(reader)
	}
}
