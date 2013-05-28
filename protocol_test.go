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

func Test_NewResponse_Status(t *testing.T) {
	response, err := NewResponse([]byte("+PONG\r\n"))

	if !response.IsStatus() {
		t.Errorf("Response is not a status response")
	}

	if err != nil {
		t.Errorf("Response is an error: %v", err)
	}

	if response.Status() != "PONG" {
		t.Errorf("Wrong value\nG: %#v\nE: %#v", response.Status(), "PONG")
	}
}

func Benchmark_NewResponse_Status(b *testing.B) {
	res := []byte("+PONG\r\n")
	for i := 0; i < b.N; i++ {
		NewResponse(res)
	}
}

func Test_NewResponse_Error(t *testing.T) {
	res, err := NewResponse([]byte("-ERR unknown command 'lorem'\r\n"))

	if err == nil {
		t.Errorf("Response is not an error: %v", res)
	}

	if err.Error() != "ERR unknown command 'lorem'" {
		t.Errorf("Error message is not properly set\nG: %v\nE: %v", err.Error(), "Err unknown command 'lorem'")
	}
}

func Benchmark_NewResponse_Error(b *testing.B) {
	res := []byte("-ERR unknown command 'lorem'\r\n")
	for i := 0; i < b.N; i++ {
		NewResponse(res)
	}
}

func Test_NewResponse_Integer(t *testing.T) {
	test_NewResponse_Integer(t, ":1234567890\r\n", 1234567890)
	test_NewResponse_Integer(t, ":-1234567890\r\n", -1234567890)
}

func test_NewResponse_Integer(t *testing.T, response string, expected int64) {
	res, err := NewResponse([]byte(response))

	if err != nil {
		t.Errorf("Response is an error: %v", err)
	}

	if !res.IsInteger() {
		t.Errorf("Response is not an integer")
	}

	n, err := res.Integer()

	if err != nil {
		t.Errorf("Cannot parse integer")
	}

	if n != expected {
		t.Errorf("Wrong integer value\nG: %v\nE: %v", n, expected)
	}
}

func Benchmark_NewResponse_Integer(b *testing.B) {
	res := []byte(":1234567890\r\n")
	for i := 0; i < b.N; i++ {
		NewResponse(res)
	}
}

func Test_NewResponse_Bulk(t *testing.T) {
	res, err := NewResponse([]byte("$6\r\nlipsum\r\n"))

	if err != nil {
		t.Errorf("Response is an error: %v", err)
	}

	if !res.IsBulk() {
		t.Errorf("Response is not a bulk")
	}

	if res.IsNull() {
		t.Errorf("Response is null")
	}

	if string(res.Value()) != "lipsum" {
		t.Errorf("Wrong bulk value\nG: %#v\nE: %#v", res.Value(), []byte("lipsum"))
	}
}

func Benchmark_NewResponse_Bulk(b *testing.B) {
	res := []byte("$6\r\nlipsum\r\n")
	for i := 0; i < b.N; i++ {
		NewResponse(res)
	}
}

func Test_NewResponse_Null(t *testing.T) {
	res, err := NewResponse([]byte("$-1\r\n"))

	if err != nil {
		t.Errorf("Response is an error: %v", err)
	}

	if !res.IsNull() {
		t.Errorf("Response is not null")
	}
}

func Benchmark_NewResponse_Null(b *testing.B) {
	res := []byte("$-1")
	for i := 0; i < b.N; i++ {
		NewResponse(res)
	}
}

func Test_NewResponse_MultiBulk(t *testing.T) {
	res, err := NewResponse([]byte("*2\r\n$5\r\nlorem\r\n$5\r\nipsum\r\n"))

	if err != nil {
		t.Errorf("Response is an error: %v", err)
	}

	if !res.IsMultiBulk() {
		t.Errorf("Response is not multi-bulk")
	}

	values := res.Values()

	if len(values) != 2 {
		t.Errorf("Wrong number of values returned\nG: %v\nE: 2", len(values))
	} else {
		for i, expected := range []string{"lorem", "ipsum"} {
			if expected != string(values[i]) {
				t.Errorf("Wrong value at %d\nG: %#v\nE: %#v", i, values[i], []byte(expected))
			}
		}
	}
}

func Benchmark_NewResponse_MultiBulk(b *testing.B) {
	res := []byte("*2\r\n$5\r\nlorem\r\n$5\r\nipsum\r\n")
	for i := 0; i < b.N; i++ {
		NewResponse(res)
	}
}

func Test_NewResponse_MultiBulk_Empty(t *testing.T) {
	res, err := NewResponse([]byte("*0\r\n"))

	if err != nil {
		t.Errorf("Response is an error: %v", err)
	}

	if !res.IsMultiBulk() {
		t.Errorf("Response is not multi-bulk")
	}

	values := res.Values()

	if len(values) != 0 {
		t.Errorf("Wrong number of values returned\nG: %v\nE: 0", len(values))
	}
}

func Benchmark_NewResponse_MultiBulk_Empty(b *testing.B) {
	res := []byte("*0\r\n")
	for i := 0; i < b.N; i++ {
		NewResponse(res)
	}
}

func Test_NewResponse_MultiBulk_Null(t *testing.T) {
	res, err := NewResponse([]byte("*-1\r\n"))

	if err != nil {
		t.Errorf("Response is an error: %v", err)
	}

	if !res.IsMultiBulk() {
		t.Errorf("Response is not multi-bulk")
	}

	if !res.IsNull() {
		t.Errorf("Response is not null")
	}

	values := res.Values()

	if len(values) != 0 {
		t.Errorf("Wrong number of values returned\nG: %v\nE: 0", len(values))
	}
}

func Benchmark_NewResponse_MultiBulk_Null(b *testing.B) {
	res := []byte("*-1\r\n")
	for i := 0; i < b.N; i++ {
		NewResponse(res)
	}
}
