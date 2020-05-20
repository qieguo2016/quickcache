package quickcache

import (
	"testing"
)

func TestRingBuf(t *testing.T) {
	rb := newRingBuf(16)
	rb.Write([]byte("fghibbbbccccddde"))
	rb.Write([]byte("fghibbbbc"))
	t.Log(string(rb.data))
	rb.Evacuate(9, 3)
	t.Log(string(rb.data))
	t.Log(rb.begin)

	data := make([]byte, 5)
	rb.ReadAt(data, rb.begin, 5)
	if string(data) != "dddef" {
		t.Fatalf("read at should be ddde, got %v", string(data))
	}
}
