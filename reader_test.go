package gedis

import (
	"strings"
	"testing"
)

func Test_readNumber(t *testing.T) {
	var n int64
	var err error

	a := Asserter{t, 1}

	n, err = readNumber(strings.NewReader("1234\r\n"))
	a.Nil(err)
	a.IntegerEq(1234, n)

	n, err = readNumber(strings.NewReader("-1234\r\n"))
	a.Nil(err)
	a.IntegerEq(-1234, n)

	_, err = readNumber(strings.NewReader("abc\r\n"))
	a.NotNil(err)
	_, err = readNumber(strings.NewReader("12ab34\r\n"))
	a.NotNil(err)
}

func Benchmark_readNumber(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		line := strings.NewReader("1234\r\n")
		b.StartTimer()
		readNumber(line)
	}
}

func Test_readLine(t *testing.T) {
	var res string
	var err error

	a := Asserter{t, 1}

	res, err = readLine(strings.NewReader("Lorem ipsum\r\n"))
	a.Nil(err)
	a.StringEq("Lorem ipsum", res)

	res, err = readLine(strings.NewReader("Lorem\ripsum\ndolor\r\n"))
	a.Nil(err)
	a.StringEq("Lorem\ripsum\ndolor", res)

	res, err = readLine(strings.NewReader("Lorem ipsum\r\ndolor sit amet"))

	a.Nil(err)
	a.StringEq("Lorem ipsum", res)
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
	a := Asserter{t, 1}

	res, err := readBulk(strings.NewReader("6\r\nlipsum\r\n"))
	a.Nil(err)
	a.StringEq("lipsum", res)

	res, err = readBulk(strings.NewReader("-1\r\n"))
	a.Nil(err)
	a.Nil(res)

	res, err = readBulk(strings.NewReader("12\r\nlorem\r\nipsum\r\n"))
	a.Nil(err)
	a.StringEq("lorem\r\nipsum", res)

	res, err = readBulk(strings.NewReader("PONG"))
	a.NotNil(err)
	a.Nil(res)
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
	a := Asserter{t, 1}

	res, err := Read(strings.NewReader("+OK\r\n"))
	a.Nil(err)
	a.StringEq("OK", res)
}

func TestRead_error(t *testing.T) {
	a := Asserter{t, 1}

	res, err := Read(strings.NewReader("-ERR unknown\r\n"))
	a.NotNil(err)
	a.Nil(res)
}

func TestRead_integer(t *testing.T) {
	a := Asserter{t, 1}

	res, err := Read(strings.NewReader(":1234\r\n"))
	a.Nil(err)
	a.IntegerEq(1234, res)

	res, err = Read(strings.NewReader(":-1234\r\n"))
	a.Nil(err)
	a.IntegerEq(-1234, res)

	_, err = Read(strings.NewReader(":lorem\r\n"))
	a.NotNil(err)
}

func TestRead_bulk(t *testing.T) {
	var res interface{}
	var err error

	a := Asserter{t, 1}

	res, err = Read(strings.NewReader("$5\r\nlorem\r\n"))
	a.Nil(err)
	a.StringEq("lorem", res)

	res, err = Read(strings.NewReader("$12\r\nlorem\r\nipsum\r\n"))
	a.Nil(err)
	a.StringEq("lorem\r\nipsum", res)

	res, err = Read(strings.NewReader("MUST FAIL"))
	a.NotNil(err)
	a.Nil(res)

	res, err = Read(strings.NewReader("$-1\r\n"))
	a.Nil(err)
	a.Nil(res)
}

func TestRead_multiBulk(t *testing.T) {
	a := Asserter{t, 1}

	input := "*4\r\n$5\r\nlorem\r\n$-1\r\n*2\r\n$5\r\nipsum\r\n$5\r\ndolor\r\n:-1234\r\n"
	reader := strings.NewReader(input)

	res, err := Read(reader)

	a.Nil(err)

	data, ok := res.([]interface{})

	if !ok {
		t.Errorf("Read() can't convert multi-bulk to []interface{}: %#v", res)
		t.FailNow()
	}

	a.StringEq("lorem", data[0])

	a.Nil(data[1])

	if bulks, ok := data[2].([]interface{}); ok {
		a.StringEq("ipsum", bulks[0])
		a.StringEq("dolor", bulks[1])
	} else {
		t.Errorf("can't convert to []interface{}: %#v", data[2])
		t.FailNow()
	}

	a.IntegerEq(-1234, data[3])
}
