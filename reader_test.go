package gedis

import (
	"strings"
	"testing"
)

func TestReadNumber(t *testing.T) {
	var n int64
	var err error

	a := Asserter{t, 1}

	n, err = ReadNumber(strings.NewReader("1234\r\n"))
	a.Nil(err)
	a.IntegerEq(1234, n)

	n, err = ReadNumber(strings.NewReader("-1234\r\n"))
	a.Nil(err)
	a.IntegerEq(-1234, n)

	_, err = ReadNumber(strings.NewReader("abc\r\n"))
	a.NotNil(err)
	_, err = ReadNumber(strings.NewReader("12ab34\r\n"))
	a.NotNil(err)
}

func BenchmarkReadNumber(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		line := strings.NewReader("1234\r\n")
		b.StartTimer()
		ReadNumber(line)
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

	res, err = readBulk(strings.NewReader("6\r\nlorem\r\n"))
	a.NotNil(err)
	a.Nil(res)

	res, err = readBulk(strings.NewReader("6\r\nlor\r\n"))
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

	res, err := Read(strings.NewReader("+PONG\r\n"))
	a.Nil(err)

	if status, ok := res.(Status); ok {
		a.StringEq("PONG", string(status))
	} else {
		t.Errorf("Can't convert to Status: %#v", res)
	}
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

func TestRead_all(t *testing.T) {
	a := Asserter{t, 1}

	/* Given a Redis server populated with the following commands
	 *
	 *   HSET hash name leandro nick inkel
	 *   INCR counter
	 *   SADD set inkel lean
	 *
	 * Test that the response of the following MULTI/EXEC block
	 *
	 *   MULTI
	 *   GET nonexisting
	 *   HGETALL hash
	 *   INCR counter
	 *   SMEMBERS set
	 *   EXEC
	 *
	 * Executing this in redis-cli returns the following:
	 *
	 *   1) (nil)
	 *   2) 1) "name"
	 *      2) "leandro"
	 *      3) "nick"
	 *      4) "inkel"
	 *   3) (integer) 3
	 *   4) 1) "inkel"
	 *      2) "lean"
	 *
	 */
	input := "*4\r\n$-1\r\n*4\r\n$4\r\nname\r\n$7\r\nleandro\r\n$4\r\nnick\r\n$5\r\ninkel\r\n:2\r\n*2\r\n$5\r\ninkel\r\n$4\r\nlean\r\n"
	reader := strings.NewReader(input)
	res, err := Read(reader)

	a.Nil(err)

	data, ok := res.([]interface{})

	if !ok {
		t.Fatalf("Can't convert to multi-bulk response: %#v", res)
	}

	a.Nil(data[0])

	hash, ok := data[1].([]interface{})

	if !ok {
		t.Fatalf("Can't convert HGETALL response: %#v", data[1])
	}

	for i, expected := range []string{"name", "leandro", "nick", "inkel"} {
		a.StringEq(expected, hash[i])
	}

	a.IntegerEq(2, data[2])

	set, ok := data[3].([]interface{})

	if !ok {
		t.Fatalf("Can't convert SMEMBERS response: %#v", data[3])
	}

	a.StringEq("inkel", set[0])
	a.StringEq("lean", set[1])
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

func TestRead(t *testing.T) {
	a := Asserter{t, 1}

	replies := "+OK\r\n" +
		"-ERR unknown\r\n" +
		":1234\r\n" +
		":-1234\r\n" +
		"$5\r\nlorem\r\n"
	input := replies + "*7\r\n$-1\r\n" + replies + "*1\r\n$2\r\nOK\r\n"
	reader := strings.NewReader(input)

	t.Logf("Using %q as input", input)

	var err error

	status, err := Read(reader)
	a.Nil(err)
	if st, ok := status.(Status); ok {
		a.StringEq("OK", string(st))
	} else {
		t.Errorf("can't convert to Status: %#v", status)
	}

	rerr, err := Read(reader)
	a.NotNil(err)
	a.Nil(rerr)
	if err.Error() != "ERR unknown" {
		t.Errorf("Unexpected: %q", err)
	}

	n, err := Read(reader)
	a.Nil(err)
	a.IntegerEq(1234, n)

	n, err = Read(reader)
	a.Nil(err)
	a.IntegerEq(-1234, n)

	bulk, err := Read(reader)
	a.Nil(err)
	a.StringEq("lorem", bulk)

	mbulk, err := Read(reader)

	a.Nil(err)

	data, ok := mbulk.([]interface{})

	if !ok {
		t.Errorf("Can't convert to []interface{}: %#v", mbulk)
		t.FailNow()
	}

	a.Nil(data[0])

	if st, ok := data[1].(Status); ok {
		a.StringEq("OK", string(st))
	} else {
		t.Errorf("Can't convert to Status: %#v", data[1])
	}

	if err, ok = data[2].(error); ok {
		if err.Error() != "ERR unknown" {
			t.Errorf("Unexpected: %q", err)
		}
	} else {
		t.Errorf("Can't convert to error: %#v", data[2])
	}

	a.IntegerEq(1234, data[3])
	a.IntegerEq(-1234, data[4])
	a.StringEq("lorem", data[5])

	if bulks, ok := data[6].([]interface{}); ok {
		a.StringEq("OK", bulks[0])
	} else {
		t.Errorf("Can't convert to []interface{}: %#v", data[6])
	}
}
