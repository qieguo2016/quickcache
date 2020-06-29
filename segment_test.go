package quickcache

import (
	"testing"
)

func TestSegment(t *testing.T) {
	seg := newSegment(128, 3)
	k1, k2, k3 := []byte("aaaa"), []byte("bbbb"), []byte("cccc")
	h1, h2, h3 := hashFunc(k1), hashFunc(k2), hashFunc(k3)
	v1, v2, v3 := []byte("1111111111"), []byte("22222222222"), []byte("3333333333")
	t.Log(seg.Set(k1, v1, h1))
	ret1, err := seg.Get(k1, h1)
	t.Log(string(ret1), err)
	t.Log(seg.Set(k2, v2, h2))
	t.Log(seg.Set(k3, v3, h3))
	ret2, err := seg.Get(k2, h2)
	t.Log(string(ret2), err)
	ret3, err := seg.Get(k3, h3)
	t.Log(string(ret3), err)
	ret1, err = seg.Get(k1, h1)
	t.Log(string(ret1), err)
	t.Log(seg.Set(k1, v2, h1))
	t.Log(seg.Set(k3, v2, h3))
	ret1, err = seg.Get(k1, h1)
	t.Log(string(ret1), err)
}
