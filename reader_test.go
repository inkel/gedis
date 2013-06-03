package gedis

import (
	"strings"
	"testing"
)

func Test_readNumber(t *testing.T) {
	var n int
	var err error

	n, err = readNumber(strings.NewReader("1234\r\n"))
	assertNil(t, 1, err)
	assertIntegerEq(t, 1, 1234, n)

	n, err = readNumber(strings.NewReader("-1234\r\n"))
	assertNil(t, 1, err)
	assertIntegerEq(t, 1, -1234, n)

	_, err = readNumber(strings.NewReader("abc\r\n"))
	assertNotNil(t, 1, err)
	_, err = readNumber(strings.NewReader("12ab34\r\n"))
	assertNotNil(t, 1, err)
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

	assertNil(t, 2, err)
	assertStringEq(t, 2, expected, res)
}

func Test_readLine(t *testing.T) {
	pass_readLine(t, "Lorem ipsum dolor sit amet")
	pass_readLine(t, "Lorem\ripsum")

	res, err := readLine(strings.NewReader("Lorem ipsum\r\ndolor sit amet"))

	assertNil(t, 2, err)
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
	assertNil(t, 1, err)
	assertStringEq(t, 1, "lipsum", res)

	res, err = readBulk(strings.NewReader("-1\r\n"))
	assertNil(t, 1, err)
	assertNil(t, 1, res)

	res, err = readBulk(strings.NewReader("12\r\nlorem\r\nipsum\r\n"))
	assertNil(t, 1, err)
	assertStringEq(t, 1, "lorem\r\nipsum", res)

	res, err = readBulk(strings.NewReader("PONG"))
	assertNotNil(t, 1, err)
	assertNil(t, 1, res)
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
	assertNil(t, 1, err)
	assertStringEq(t, 1, "OK", res)
}

func TestRead_error(t *testing.T) {
	res, err := Read(strings.NewReader("-ERR unknown\r\n"))
	assertNotNil(t, 1, err)
	assertNil(t, 1, res)
}

func TestRead_integer(t *testing.T) {
	res, err := Read(strings.NewReader(":1234\r\n"))
	assertNil(t, 1, err)
	assertIntegerEq(t, 1, 1234, res)

	res, err = Read(strings.NewReader(":-1234\r\n"))
	assertNil(t, 1, err)
	assertIntegerEq(t, 1, -1234, res)

	_, err = Read(strings.NewReader(":lorem\r\n"))
	assertNotNil(t, 1, err)
}

func TestRead_bulk(t *testing.T) {
	var res interface{}
	var err error

	res, err = Read(strings.NewReader("$5\r\nlorem\r\n"))
	assertNil(t, 1, err)
	assertStringEq(t, 1, "lorem", res)

	res, err = Read(strings.NewReader("$12\r\nlorem\r\nipsum\r\n"))
	assertNil(t, 1, err)
	assertStringEq(t, 1, "lorem\r\nipsum", res)

	res, err = Read(strings.NewReader("MUST FAIL"))
	assertNotNil(t, 1, err)
	assertNil(t, 1, res)

	res, err = Read(strings.NewReader("$-1\r\n"))
	assertNil(t, 1, err)
	assertNil(t, 1, res)
}

func TestRead_multiBulk(t *testing.T) {
	input := "*4\r\n$5\r\nlorem\r\n$-1\r\n*2\r\n$5\r\nipsum\r\n$5\r\ndolor\r\n:-1234\r\n"
	reader := strings.NewReader(input)

	res, err := Read(reader)

	assertNil(t, 1, err)

	data, ok := res.([]interface{})

	if !ok {
		t.Errorf("Read() can't convert multi-bulk to []interface{}: %#v", res)
		t.FailNow()
	}

	assertStringEq(t, 1, "lorem", data[0])

	assertNil(t, 1, data[1])

	if bulks, ok := data[2].([]interface{}); ok {
		assertStringEq(t, 1, "ipsum", bulks[0])
		assertStringEq(t, 1, "dolor", bulks[1])
	} else {
		t.Errorf("can't convert to []interface{}: %#v", data[2])
		t.FailNow()
	}

	assertIntegerEq(t, 1, -1234, data[3])
}
