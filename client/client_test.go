package client

import (
	"bytes"
	"path"
	"runtime"
	"testing"
)

var _ = bytes.Equal

func notErr(t *testing.T, err error) {
	if err != nil {
		_, file, ln, _ := runtime.Caller(1)
		t.Fatalf("\r\t%s:%d: Unexpected error: %v", path.Base(file), ln, err)
	}
}

const key = "gedis.client:key"

func TestClient(t *testing.T) {
	c, err := Dial("tcp", ":26739")
	if err != nil {
		t.Skip("NOTE: This test needs a *live* test redis server running on port 26379:", err)
	}

	var res interface{}

	res, err = c.Send("SET", key, "lorem\r\nipsum")
	notErr(t, err)
	if s, ok := res.(string); !ok || s != "OK" {
		t.Fatalf("Unexpected: %#v", res)
	}

	res, err = c.Send("GET", key)
	notErr(t, err)
	if s, ok := res.(string); !ok || s != "lorem\r\nipsum" {
		t.Fatalf("Unexpected: %#v", res)
	}

	res, err = c.Send("DEL", key)
	notErr(t, err)
	if n, ok := res.(int64); !ok || n != 1 {
		t.Fatalf("Unexpected: %#v", res)
	}

	_, err = c.Send("SADD", key, "lorem", "1234")
	notErr(t, err)

	res, err = c.Send("SMEMBERS", key)
	notErr(t, err)

	arr, ok := res.([]interface{})
	if !ok {
		t.Fatalf("Unexpected: %#v", res)
	}

	if s, ok := arr[0].(string); !ok {
		t.Fatalf("Unexpected: %#v", arr[0])
	} else if s != "1234" {
		t.Fatalf("Unexpected: %#v", s)
	}

	if s, ok := arr[1].(string); !ok {
		t.Fatalf("Unexpected: %#v", arr[0])
	} else if s != "lorem" {
		t.Fatalf("Unexpected: %#v", s)
	}

	res, err = c.Send("DEL", key)
	notErr(t, err)
	if n, ok := res.(int64); !ok || n != 1 {
		t.Fatalf("Unexpected: %#v", res)
	}

	res, err = c.Send("GET", key)
	notErr(t, err)
	if res != nil {
		t.Fatalf("Unexpected: %#v", res)
	}
}
