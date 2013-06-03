package gedis

import (
	"fmt"
	"io"
)

// Interface for reading Redis replies
type Reader interface {
	Read(b []byte) (n int, err error)
}

func readNumber(r io.Reader) (n int, err error) {
	b := make([]byte, 1)

	sign := 1

	_, err = r.Read(b)
	if err != nil {
		return
	}
	if b[0] == '-' {
		sign = -1
		b[0] = '0'
	}

	for {
		if b[0] >= '0' && b[0] <= '9' {
			n = n*10 + int(b[0]-'0')
		} else if b[0] == '\r' {
			_, err = r.Read(b)
			if b[0] == '\n' {
				break
			} else {
				return 0, fmt.Errorf("Invalid character after '\r': %q", b)
			}
		} else {
			return 0, fmt.Errorf("Invalid character: %q", b)
		}

		_, err = r.Read(b)
		if err == io.EOF {
			break
		} else if err != nil {
			return
		}
	}

	return sign * n, nil
}

func readLine(r io.Reader) (line string, err error) {
	bs := make([]byte, 1024)
	l := 0

	b := make([]byte, 1)

	for {
		_, err = r.Read(b)

		if err == io.EOF {
			err = nil
			break
		} else if err != nil {
			return "", err
		} else if b[0] == '\r' {
			_, err = r.Read(b)
			if err != nil {
				return "", err
			}
			if b[0] == '\n' {
				break
			} else {
				bs[l] = '\r'
				l++
			}
		}

		bs[l] = b[0]
		l++
	}

	line = string(bs[:l])

	return line, err
}

func readBulk(r Reader) (interface{}, error) {
	var bs []byte

	num_bytes, err := readNumber(r)
	if err != nil {
		return nil, err
	}

	if num_bytes == -1 {
		return nil, nil
	}

	bs = make([]byte, num_bytes)
	b := make([]byte, 1)

	for i := 0; i < num_bytes; i++ {
		_, err = r.Read(b)
		if err != nil {
			return nil, err
		}
		bs[i] = b[0]
	}

	// Must read following two bytes for \r\n
	crlf := make([]byte, 2)
	r.Read(crlf)

	return string(bs), nil
}

func Read(r Reader) (ret interface{}, err error) {
	kind := make([]byte, 1)

	_, err = r.Read(kind)
	if err != nil {
		return
	}

	switch kind[0] {
	case '+':
		ret, err = readLine(r)
		if err == nil {
			if bs, ok := ret.(string); ok {
				return bs, nil
			} else {
				return nil, fmt.Errorf("Cannot convert to string: %#v", ret)
			}
		} else {
			return nil, err
		}
	case '-':
		ret, err = readLine(r)

		if err == nil {
			if bs, ok := ret.([]byte); ok {
				err = fmt.Errorf(string(bs))
			} else {
				err = fmt.Errorf("Cannot convert to []byte: %#v", ret)
			}
			ret = nil
		}
	case ':':
		ret, err = readNumber(r)
	case '$':
		ret, err = readBulk(r)
		if err == nil {
			if ret == nil {
				return nil, nil
			}
			if bs, ok := ret.(string); ok {
				return bs, nil
			} else {
				return nil, fmt.Errorf("Cannot convert to string: %#v", ret)
			}
		} else {
			return nil, err
		}
	case '*':
		n, err := readNumber(r)
		if err != nil {
			return nil, err
		}

		if n == -1 {
			return nil, nil
		}

		res := make([]interface{}, n)

		for i := 0; i < n; i++ {
			ret, err := Read(r)
			if err == nil {
				res[i] = ret
			} else {
				res[i] = err
			}
		}

		ret = res
	}

	return
}
