package gedis

import (
	"bytes"
	"errors"
	"testing"
)

func Test_writeBulk(t *testing.T) {
	expected := []byte("$4\r\nPING\r\n")
	parsed := WriteBulk("PING")

	if !bytes.Equal(expected, parsed) {
		t.Errorf("writeBulk(%#v)\nG: %v\nE: %v", "PING", parsed, expected)
	}
}

func Benchmark_writeBulk(b *testing.B) {
	for i := 0; i < b.N; i++ {
		WriteBulk("PING")
	}
}

func Test_writeMultiBulk(t *testing.T) {
	cmd := "*1\r\n$4\r\nPING\r\n"
	expected := []byte(cmd)

	if parsed := WriteMultiBulk("PING"); !bytes.Equal(expected, parsed) {
		t.Errorf("writeMultiBulk(%#v)\nG: %v\nE: %v", cmd, parsed, expected)
	}

	cmd = "*3\r\n$3\r\nSET\r\n$5\r\nlorem\r\n$5\r\n12345\r\n"
	expected = []byte(cmd)

	if parsed := WriteMultiBulk("SET", "lorem", "12345"); !bytes.Equal(expected, parsed) {
		t.Errorf("writeMultiBulk(%#v)\nG: %v\nE: %v", cmd, parsed, expected)
	}
}

func Benchmark_writeMultiBulk(b *testing.B) {
	for i := 0; i < b.N; i++ {
		WriteMultiBulk("SET", "lorem", "12345")
	}
}

func Test_writeInt(t *testing.T) {
	expected := []byte(":1234\r\n")
	parsed := WriteInt(1234)

	if !bytes.Equal(expected, parsed) {
		t.Errorf("\nexpected %#v\nreturned %#v", expected, parsed)
		t.FailNow()
	}

	expected = []byte(":-1234\r\n")
	parsed = WriteInt(-1234)

	if !bytes.Equal(expected, parsed) {
		t.Errorf("\nexpected %#v\nreturned %#v", expected, parsed)
		t.FailNow()
	}
}

func Test_writeError(t *testing.T) {
	err := errors.New("unknown")
	expected := []byte("-ERR unknown\r\n")
	parsed := WriteError(err)

	if !bytes.Equal(expected, parsed) {
		t.Errorf("\nexpected %q\nreturned %q", expected, parsed)
	}
}

func Test_writeStatus(t *testing.T) {
	expected := []byte("+OK\r\n")
	parsed := WriteStatus("OK")

	if !bytes.Equal(expected, parsed) {
		t.Errorf("\nexpected %q\nreturned %q", expected, parsed)
	}
}

func Benchmark_WriteStatus(b *testing.B) {
	for i := 0; i < b.N; i++ {
		WriteStatus("OK")
	}
}

func TestWrite(t *testing.T) {
	var writer bytes.Buffer

	expected := "*4\r\n$4\r\nPING\r\n:123\r\n$-1\r\n-ERR unknown\r\n"

	Write(&writer, "PING", 123, nil, errors.New("unknown"))

	if res := writer.String(); expected != res {
		t.Errorf("Write()\nexpected %q\nreturned %q", expected, res)
	}
}

func TestWrite_error(t *testing.T) {
	var writer bytes.Buffer

	a := Asserter{t, 1}

	_, err := Write(&writer)
	a.NotNil(err)
	a.StringEq("", writer.String())
}
