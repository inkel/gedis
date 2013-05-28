package gedis

import (
	"bytes"
	"errors"
	"strconv"
)

// Redis protocoal

// *<num args> CR LF
// $<num bytes arg1> CR LF
// <arg data> CR LF
// ...
// $<num bytes argn> CR LF
// <arg data>

func WriteBulk(bulk string) []byte {
	var buffer bytes.Buffer

	buffer.WriteByte('$')
	buffer.WriteString(strconv.Itoa(len(bulk)))
	buffer.WriteString("\r\n")

	buffer.WriteString(bulk)
	buffer.WriteString("\r\n")

	return buffer.Bytes()
}

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

func (r *Response) IsStatus() bool {
	return r.kind == StatusReply
}

func (r *Response) IsInteger() bool {
	return r.kind == IntegerReply
}

func (r *Response) IsBulk() bool {
	return r.kind == BulkReply
}

func (r *Response) IsMultiBulk() bool {
	return r.kind == MultiBulkReply
}

func (r *Response) Status() string {
	return string(r.value)
}

func (r *Response) Integer() (int64, error) {
	return strconv.ParseInt(string(r.value), 10, 64)
}

func (r *Response) Value() []byte {
	return r.value
}

func (r *Response) IsNull() bool {
	return r.null
}

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

func NewResponse(data []byte) (r Response, err error) {
	r.kind = data[0]

	switch r.kind {
	case '+', ':', '-':
		value, _ := readLine(data, 1)

		if r.kind == '-' {
			return r, errors.New(string(value))
		} else {
			r.value = value
		}
	case '$':
		r.value, _, err = ReadBulk(data, 0)

		if len(r.value) == 0 {
			r.null = true
		}

		if err != nil {
			return
		}
	case '*':
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
