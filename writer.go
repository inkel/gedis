package gedis

import (
	"bytes"
	"fmt"
	"strconv"
)

// Interface for writing Redis commands
type Writer interface {
	Write(p []byte) (n int, err error)
}

func Write(w Writer, args ...interface{}) (n int, err error) {
	if len(args) == 0 {
		return -1, fmt.Errorf("Must write at least one argument")
	}
	return w.Write(WriteMultiBulk(args...))
}

// Writes a string as a sequence of bytes to be send to a Redis
// instance, using the Redis Bulk format.
func WriteBulk(bulk string) []byte {
	bulk_len := strconv.Itoa(len(bulk))

	// '$' + len(string(len(bulk))) + "\r\n" + len(bulk) + "\r\n"
	n := 1 + len(bulk_len) + 2 + len(bulk) + 2

	bytes := make([]byte, n)

	bytes[0] = '$'

	j := 1

	for _, c := range bulk_len {
		bytes[j] = byte(c)
		j++
	}

	bytes[j] = '\r'
	bytes[j+1] = '\n'
	j += 2

	for _, c := range bulk {
		bytes[j] = byte(c)
		j++
	}

	bytes[j] = '\r'
	bytes[j+1] = '\n'

	return bytes
}

// Writes a number in the Redis protocol format
func WriteInt(n int64) []byte {
	return []byte(":" + strconv.FormatInt(n, 10) + "\r\n")
}

// Writes an error in the Redis protocol format
func WriteError(err error) []byte {
	return []byte("-ERR " + err.Error() + "\r\n")
}

// Writes a status in the Redis protocol format
func WriteStatus(status string) []byte {
	bs := make([]byte, len(status)+3)
	bs[0] = '+'
	l := 1
	for _, r := range status {
		bs[l] = byte(r)
		l++
	}
	bs[l] = '\r'
	bs[l+1] = '\n'
	return bs
}

// BUG(inkel): writeMultiBulk can't write multi-bulks inside multi-bulks

// Writes a sequence of strings as a sequence of bytes to be send to a
// Redis instance, using the Redis Multi-Bulk format.
func WriteMultiBulk(args ...interface{}) []byte {
	var buffer bytes.Buffer

	buffer.WriteByte('*')
	buffer.WriteString(strconv.Itoa(len(args)))
	buffer.WriteString("\r\n")

	var bs []byte

	for _, arg := range args {
		bs = []byte{}

		switch arg := arg.(type) {
		case string:
			bs = WriteBulk(arg)
		case int:
			bs = WriteInt(int64(arg))
		case int64:
			bs = WriteInt(arg)
		case error:
			bs = WriteError(arg)
		case nil:
			bs = []byte("$-1\r\n")
		default:
			panic(fmt.Errorf("Unrecognized type: %#v", arg))
		}

		buffer.Write(bs)
	}

	return buffer.Bytes()
}
