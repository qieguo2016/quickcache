package quickcache

import (
	"testing"
)

func TestSegment(t *testing.T) {
	seg := newSegment(128, 3)
	k1, k2, k3 := []byte("a"), []byte("bb"), []byte("ccc")
	h1, h2, h3 := hashFunc(k1), hashFunc(k2), hashFunc(k3)
	v1, v2, v3 := []byte("1111"), []byte("222222222"), []byte("33333")
	seg.Set(k1, v1, h1)
	ret1, err := seg.Get(k1, h1)
	t.Log(string(ret1), err)
	seg.Set(k2, v2, h2)
	seg.Set(k3, v3, h3)

}
