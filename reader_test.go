package gedis

import (
	"runtime"
	"strings"
	"testing"
)

func pass_readNumber(t *testing.T, line string, expected int) {
	n, err := readNumber(strings.NewReader(line))
	assertNotError(t, 2, err)
	assertIntegerEq(t, 2, expected, n)
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

func pass_readLine(t *testing.T, expected string) {
	input := []byte(expected + "\r\n")
	reader := strings.NewReader(string(input))

	res, err := readLine(reader)

	assertNotError(t, 2, err)
	assertStringEq(t, 2, expected, res)
}

func Test_readLine(t *testing.T) {
	pass_readLine(t, "Lorem ipsum dolor sit amet")
	pass_readLine(t, "Lorem\ripsum")

	res, err := readLine(strings.NewReader("Lorem ipsum\r\ndolor sit amet"))

	assertNotError(t, 2, err)
	assertStringEq(t, 2, "Lorem ipsum", res)
}

func Benchmark_readLine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		line := strings.NewReader("Lorem ipsum\rdolor sit amet\r\n")
		b.StartTimer()
		readLine(line)
	}
}

func Test_readBulk(t *testing.T) {
	res, err := readBulk(strings.NewReader("6\r\nlipsum\r\n"))
	assertNotError(t, 1, err)
	assertStringEq(t, 1, "lipsum", res)

	res, err = readBulk(strings.NewReader("-1\r\n"))
	assertNotError(t, 1, err)

	if res != nil {
		t.Errorf("Expected nil, returned %#v", res)
		t.FailNow()
	}

	res, err = readBulk(strings.NewReader("12\r\nlorem\r\nipsum\r\n"))
	assertNotError(t, 1, err)
	assertStringEq(t, 1, "lorem\r\nipsum", res)

	if res, err := readBulk(strings.NewReader("PONG")); err == nil {
		t.Errorf("readBulk() should've returned an error, returned: %#v", res)
	}
}

func Benchmark_readBulk(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		reader := strings.NewReader("12\r\nlorem\r\nipsum\r\n")
		b.StartTimer()
		readBulk(reader)
	}
}

func TestRead_status(t *testing.T) {
	res, err := Read(strings.NewReader("+OK\r\n"))
	assertNotError(t, 1, err)
	assertStringEq(t, 1, "OK", res)
}

func TestRead_error(t *testing.T) {
	res, err := Read(strings.NewReader("-ERR unknown\r\n"))

	if err == nil {
		t.Errorf("Error expected, returned: %#v", res)
	}
}

func TestRead_integer(t *testing.T) {
	res, err := Read(strings.NewReader(":1234\r\n"))
	assertNotError(t, 1, err)
	assertIntegerEq(t, 1, 1234, res)

	res, err = Read(strings.NewReader(":-1234\r\n"))
	assertNotError(t, 1, err)
	assertIntegerEq(t, 1, -1234, res)

	res, err = Read(strings.NewReader(":lorem\r\n"))
	if err == nil {
		t.Errorf("Error expected, returned: %#v", res)
	}
}

func TestRead_bulk(t *testing.T) {
	var res interface{}
	var err error

	res, err = Read(strings.NewReader("$5\r\nlorem\r\n"))
	assertNotError(t, 1, err)
	assertStringEq(t, 1, "lorem", res)

	res, err = Read(strings.NewReader("$12\r\nlorem\r\nipsum\r\n"))
	assertNotError(t, 1, err)
	assertStringEq(t, 1, "lorem\r\nipsum", res)

	res, err = Read(strings.NewReader("MUST FAIL"))
	if err == nil {
		t.Errorf("Error expected, returned: %#v", res)
		t.FailNow()
	}

	res, err = Read(strings.NewReader("$-1\r\n"))
	assertNotError(t, 1, err)
	if res != nil {
		t.Errorf("nil expected, returned: %#v", res)
	}
}

func TestRead_multiBulk(t *testing.T) {
	input := "*4\r\n$5\r\nlorem\r\n$-1\r\n*2\r\n$5\r\nipsum\r\n$5\r\ndolor\r\n:-1234\r\n"
	reader := strings.NewReader(input)

	res, err := Read(reader)

	assertNotError(t, 1, err)

	data, ok := res.([]interface{})

	if !ok {
		t.Errorf("Read() can't convert multi-bulk to []interface{}: %#v", res)
		t.FailNow()
	}

	assertStringEq(t, 1, "lorem", data[0])

	if data[1] != nil {
		t.Errorf("nil expected, got: %#v", data[1])
		t.FailNow()
	}

	if bulks, ok := data[2].([]interface{}); ok {
		assertStringEq(t, 1, "ipsum", bulks[0])
		assertStringEq(t, 1, "dolor", bulks[1])
	} else {
		t.Errorf("can't convert to []interface{}: %#v", data[2])
		t.FailNow()
	}

	assertIntegerEq(t, 1, -1234, data[3])
}
