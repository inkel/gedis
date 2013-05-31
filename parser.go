package gedis

import (
	"io"
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
	}

	return
}
