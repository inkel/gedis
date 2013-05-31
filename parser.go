package gedis

import (
	"io"
	"fmt"
)

func Parse(r io.Reader) (ret interface{}, err error) {
	kind := make([]byte, 1)

	_, err = r.Read(kind)
	if err != nil {
		return
	}

	switch kind[0] {
	case '+':
		ret, err = readLine(r)
		if err == nil {
			if bs, ok := ret.([]byte); ok {
				return string(bs), nil
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
		ret, err = ReadBulk(r)
		if err == nil {
			if bs, ok := ret.([]byte); ok {
				return string(bs), nil
			} else {
				return nil, fmt.Errorf("Cannot convert to string: %#v", ret)
			}
		} else {
			return nil, err
		}
	}

	return
}
