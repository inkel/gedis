package gedis

import (
	"fmt"
	"io"
)

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

func readLine(r io.Reader) (bs []byte, err error) {
	bs = make([]byte, 1024)
	l := 0

	b := make([]byte, 1)

	for {
		_, err = r.Read(b)

		if err == io.EOF {
			err = nil
			break
		} else if err != nil {
			return []byte{}, err
		} else if b[0] == '\r' {
			_, err = r.Read(b)
			if err != nil {
				return []byte{}, err
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

	return bs[:l], err
}
