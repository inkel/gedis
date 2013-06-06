package server

import "github.com/inkel/gedis"

// Read a bulk as defined in the Redis protocol
//
// This functon is similar to that of gedis.ReadBulk, however given
// that a Redis client can only send a multi-bulk requests that only
// include non-nil bulks of bytes, a simplified version that returns a
// sequence of bytes is provided.
func readBulk(r gedis.Reader) (bs []byte, err error) {
	var b byte

	b, err = readByte(r)
	if err != nil {
		return bs, err
	} else if b != '$' {
		return bs, gedis.NewParseError("Invalid first character")
	}

	n, err := gedis.ReadNumber(r)
	if err != nil {
		return bs, err
	}

	bs = make([]byte, n)

	_, err = r.Read(bs)
	if err != nil {
		return bs, err
	}

	crlf := make([]byte, 2)

	if _, err = r.Read(crlf); err != nil {
		return bs, err
	}

	if crlf[0] != '\r' || crlf[1] != '\n' {
		return bs, gedis.NewParseError("Invalid EOL")
	}

	return
}

// Helper function to read the next byte in a gedis.Reader
func readByte(r gedis.Reader) (byte, error) {
	b := make([]byte, 1)
	_, err := r.Read(b)
	return b[0], err
}

// Read a multi-bulk request from a Redis client
//
// This function is similar in implementation to that of gedis.Read,
// however a Redis client can only send multi-bulk requests to a Redis
// server, so a simplified version is implemented for reading Redis
// commands from clients.
//
// In truth they can also send an inline request, however that is
// currently not covered by this implementation.
func Read(r gedis.Reader) (res [][]byte, err error) {
	var b byte

	b, err = readByte(r)
	if err != nil {
		return
	}

	if b != '*' {
		return res, gedis.NewParseError("Invalid first character")
	} else {
		n, err := gedis.ReadNumber(r)
		if err != nil {
			return res, err
		}

		res = make([][]byte, n)

		for i := int64(0); i < n; i++ {
			res[i], err = readBulk(r)
			if err != nil {
				return res, err
			}
		}
	}

	return res, err
}
