package gedis

import (
	"bytes"
	"errors"
	"strconv"
)

/*
 Redis protocoal

 Redis uses a very simple text protocol, which is binary safe.

 *<num args> CR LF
 $<num bytes arg1> CR LF
 <arg data> CR LF
 ...
 $<num bytes argn> CR LF
 <arg data>
*/

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

// Writes a sequence of strings as a sequence of bytes to be send to a
// Redis instance, using the Redis Multi-Bulk format.
func WriteMultiBulk(cmd string, args ...string) []byte {
	var buffer bytes.Buffer

	buffer.WriteByte('*')
	buffer.WriteString(strconv.Itoa(1 + len(args)))
	buffer.WriteString("\r\n")

	buffer.Write(WriteBulk(cmd))

	for _, elem := range args {
		buffer.Write(WriteBulk(elem))
	}

	return buffer.Bytes()
}

const (
	StatusReply    = '+'
	ErrorReply     = '-'
	IntegerReply   = ':'
	BulkReply      = '$'
	MultiBulkReply = '*'
)

// Represents a response sent by a Redis server.
type Response struct {
	kind   byte
	status string
	value  []byte
	null   bool
	values [][]byte
}

type ResponseError struct {
	kind, msg string
}

// Returns true if the response is a status response.
func (r *Response) IsStatus() bool {
	return r.kind == StatusReply
}

// Returns true if the response is an integer response.
func (r *Response) IsInteger() bool {
	return r.kind == IntegerReply
}

// Returns true if the response is a bulk response
func (r *Response) IsBulk() bool {
	return r.kind == BulkReply
}

// Returns true if the response is a multi-bulk response.
func (r *Response) IsMultiBulk() bool {
	return r.kind == MultiBulkReply
}

// Returns the status returned in a status response.
func (r *Response) Status() string {
	return string(r.value)
}

// Returns the Int64 in an integer response.
func (r *Response) Integer() (int64, error) {
	return strconv.ParseInt(string(r.value), 10, 64)
}

// Returns the raw value of the response.
func (r *Response) Value() []byte {
	return r.value
}

// Returns true if the bulk or multi-bulk response was nil.
func (r *Response) IsNull() bool {
	return r.null
}

// Returns the raw values in a multi-bulk response.
func (r *Response) Values() [][]byte {
	return r.values
}

func readLine(data []byte, offset int64) ([]byte, int64) {
	n := int64(len(data))
	i := 0

	buffer := make([]byte, n-offset)

	for offset < n {
		if data[offset] == '\r' && data[offset+1] == '\n' {
			break
		} else {
			buffer[i] = data[offset]
			i++
			offset++
		}
	}

	return buffer[:i], int64(offset + 2)
}

// Reads the first bulk response from a sequence of bytes.
func ReadBulk(data []byte, offset int64) ([]byte, int64, error) {
	var value []byte
	var n_offset int64

	value, offset = readLine(data, offset+1)

	num_bytes, err := strconv.ParseInt(string(value), 10, 64)
	if err != nil {
		return value, offset, err
	}

	if num_bytes == int64(-1) {
		return []byte{}, offset, nil
	}

	n_offset = offset + num_bytes

	// This magic 2 is for ending \r\n
	if offset+num_bytes+2 <= int64(len(data)) {
		return data[offset:n_offset], n_offset + 2, nil
	} else {
		return []byte{}, offset, errors.New("not enough bytes in data stream")
	}
}

// Creates a new Response object from a sequence of bytes.
func NewResponse(data []byte) (r Response, err error) {
	r.kind = data[0]

	switch r.kind {
	case StatusReply, IntegerReply, ErrorReply:
		value, _ := readLine(data, 1)

		if r.kind == '-' {
			return r, errors.New(string(value))
		} else {
			r.value = value
		}
	case BulkReply:
		r.value, _, err = ReadBulk(data, 0)

		if len(r.value) == 0 {
			r.null = true
		}

		if err != nil {
			return
		}
	case MultiBulkReply:
		value, offset := readLine(data, 1)

		num_values, err := strconv.ParseInt(string(value), 10, 64)
		if err != nil {
			return r, err
		}

		if num_values == -1 {
			r.null = true
		} else {
			r.values = make([][]byte, num_values)

			for i := int64(0); i < num_values; i++ {
				r.values[i], offset, err = ReadBulk(data, offset)
			}
		}
	}

	return
}
