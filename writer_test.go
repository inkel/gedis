package gedis

import (
	"bytes"
	"testing"
)

func Test_WriteBulk(t *testing.T) {
	expected := []byte("$4\r\nPING\r\n")
	parsed := WriteBulk("PING")

	if !bytes.Equal(expected, parsed) {
		t.Errorf("WriteBulk(%#v)\nG: %v\nE: %v", "PING", parsed, expected)
	}
}

func Benchmark_WriteBulk(b *testing.B) {
	for i := 0; i < b.N; i++ {
		WriteBulk("PING")
	}
}

func Test_WriteMultiBulk(t *testing.T) {
	cmd := "*1\r\n$4\r\nPING\r\n"
	expected := []byte(cmd)

	if parsed := WriteMultiBulk("PING"); !bytes.Equal(expected, parsed) {
		t.Errorf("WriteMultiBulk(%#v)\nG: %v\nE: %v", cmd, parsed, expected)
	}

	cmd = "*3\r\n$3\r\nSET\r\n$5\r\nlorem\r\n$5\r\n12345\r\n"
	expected = []byte(cmd)

	if parsed := WriteMultiBulk("SET", "lorem", "12345"); !bytes.Equal(expected, parsed) {
		t.Errorf("WriteMultiBulk(%#v)\nG: %v\nE: %v", cmd, parsed, expected)
	}
}

func Benchmark_WriteMultiBulk(b *testing.B) {
	for i := 0; i < b.N; i++ {
		WriteMultiBulk("SET", "lorem", "12345")
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
