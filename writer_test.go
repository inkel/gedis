package gedis

import (
	"bytes"
	"errors"
	"testing"
)

func Test_writeBulk(t *testing.T) {
	expected := []byte("$4\r\nPING\r\n")
	parsed := writeBulk("PING")

	if !bytes.Equal(expected, parsed) {
		t.Errorf("writeBulk(%#v)\nG: %v\nE: %v", "PING", parsed, expected)
	}
}

func Benchmark_writeBulk(b *testing.B) {
	for i := 0; i < b.N; i++ {
		writeBulk("PING")
	}
}

func Test_writeMultiBulk(t *testing.T) {
	cmd := "*1\r\n$4\r\nPING\r\n"
	expected := []byte(cmd)

	if parsed := writeMultiBulk("PING"); !bytes.Equal(expected, parsed) {
		t.Errorf("writeMultiBulk(%#v)\nG: %v\nE: %v", cmd, parsed, expected)
	}

	cmd = "*3\r\n$3\r\nSET\r\n$5\r\nlorem\r\n$5\r\n12345\r\n"
	expected = []byte(cmd)

	if parsed := writeMultiBulk("SET", "lorem", "12345"); !bytes.Equal(expected, parsed) {
		t.Errorf("writeMultiBulk(%#v)\nG: %v\nE: %v", cmd, parsed, expected)
	}
}

func Benchmark_writeMultiBulk(b *testing.B) {
	for i := 0; i < b.N; i++ {
		writeMultiBulk("SET", "lorem", "12345")
	}
}

func Test_writeInt(t *testing.T) {
	expected := []byte(":1234\r\n")
	parsed := writeInt(1234)

	if !bytes.Equal(expected, parsed) {
		t.Errorf("\nexpected %#v\nreturned %#v", expected, parsed)
		t.FailNow()
	}

	expected = []byte(":-1234\r\n")
	parsed = writeInt(-1234)

	if !bytes.Equal(expected, parsed) {
		t.Errorf("\nexpected %#v\nreturned %#v", expected, parsed)
		t.FailNow()
	}
}

func Test_writeError(t *testing.T) {
	err := errors.New("ERR unknown")
	expected := []byte("-ERR unknown\r\n")
	parsed := writeError(err)

	if !bytes.Equal(expected, parsed) {
		t.Errorf("\nexpected %q\nreturned %q", expected, parsed)
	}
}

func TestWrite(t *testing.T) {
	var writer bytes.Buffer

	expected := "*1\r\n$4\r\nPING\r\n"

	Write(&writer, "PING")

	if res := writer.String(); expected != res {
		t.Errorf("Write()\nexpected %q\nreturned %q", expected, res)
	}
}
