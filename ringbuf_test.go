package quickcache

import (
	"testing"
)

func TestRingBuf(t *testing.T) {
	rb := newRingBuf(16)
	rb.Write([]byte("abcdefghijklmnop"))
	rb.Write([]byte("qrstuvw"))
	t.Log(string(rb.data))
	rb.Evacuate(9, 3)
	if string(rb.data) != "qrstuvwhijhimnop" {
		t.Fatalf("expect %v, got %v", "qrstuvwhijhimnop", string(rb.data))
	}
	t.Log(string(rb.data))
	t.Log(rb.begin)

	a := make([]byte, 5)
	rb.ReadAt(a, rb.begin, 5)
	if string(a) != string(rb.data[10:15]) {
		t.Fatalf("expect %v, got %v", string(rb.data[10:15]), string(a))
	}

	b := []byte("abcdefg")
	c := rb.CheckSpace(len(b))
	t.Log(c)
	if c > 0 {
		rb.Evacuate(rb.Begin(), c)
	}
	t.Log(string(rb.data))
	rb.Write(b)
	t.Log(string(rb.data))
}
