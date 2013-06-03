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
	return w.Write(writeMultiBulk(args...))
}

// Writes a string as a sequence of bytes to be send to a Redis
// instance, using the Redis Bulk format.
func writeBulk(bulk string) []byte {
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
func writeInt(n int) []byte {
	return []byte(":" + strconv.Itoa(n) + "\r\n")
}

// Writes an error in the Redis protocol format
func writeError(err error) []byte {
	return []byte("-" + err.Error() + "\r\n")
}

// BUG(inkel): writeMultiBulk can't write multi-bulks inside multi-bulks

// Writes a sequence of strings as a sequence of bytes to be send to a
// Redis instance, using the Redis Multi-Bulk format.
func writeMultiBulk(args ...interface{}) []byte {
	var buffer bytes.Buffer

	buffer.WriteByte('*')
	buffer.WriteString(strconv.Itoa(len(args)))
	buffer.WriteString("\r\n")

	var bs []byte

	for _, arg := range args {
		bs = []byte{}

		switch arg.(type) {
		case string:
			bs = writeBulk(arg.(string))
		case int:
			bs = writeInt(arg.(int))
		case error:
			bs = writeError(arg.(error))
		case nil:
			bs = []byte("$-1\r\n")
		default:
			panic(fmt.Errorf("Unrecognized type: %#v", arg))
		}

		buffer.Write(bs)
	}

	return buffer.Bytes()
}
