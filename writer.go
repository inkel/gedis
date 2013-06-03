package gedis

import (
	"bytes"
	"strconv"
)

// Interface for writing Redis commands
type Writer interface {
	Write(p []byte) (n int, err error)
}

func Write(w Writer, cmd string, args ...string) (n int, err error) {
	return w.Write(writeMultiBulk(cmd, args...))
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

// Writes a sequence of strings as a sequence of bytes to be send to a
// Redis instance, using the Redis Multi-Bulk format.
func writeMultiBulk(cmd string, args ...string) []byte {
	var buffer bytes.Buffer

	buffer.WriteByte('*')
	buffer.WriteString(strconv.Itoa(1 + len(args)))
	buffer.WriteString("\r\n")

	buffer.Write(writeBulk(cmd))

	for _, elem := range args {
		buffer.Write(writeBulk(elem))
	}

	return buffer.Bytes()
}
