package gedis

import (
	"io"
	"errors"
)

func Parse(r io.Reader) (bs []byte, err error) {
	kind := make([]byte, 1)

	_, err = r.Read(kind)
	if err != nil {
		return
	}

	switch kind[0] {
	case '+':
		bs, err = readLine(r)
	case '-':
		bs, err = readLine(r)
		if err == nil {
			err = errors.New(string(bs))
			bs = make([]byte, 1)
		}
	}

	return
}
