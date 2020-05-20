package quickcache

import (
	"bytes"
)

type ringBuf struct {
	begin int64
	end   int64
	data  []byte
}

func newRingBuf(size int) ringBuf {
	rb := ringBuf{
		begin: 0,
		end:   0,
		data:  make([]byte, size),
	}
	return rb
}

func (r *ringBuf) Begin() int64 {
	return r.begin
}

func (r *ringBuf) End() int64 {
	return r.end
}

func (r *ringBuf) Size() int {
	return len(r.data)
}

func (r *ringBuf) Write(val []byte) error {
	return r.WriteAt(val, r.end)
}

func (r *ringBuf) WriteAt(val []byte, offset int64) error {
	if len(val) > len(r.data) {
		return ErrLargeValue
	}
	if offset < r.begin || offset > r.end {
		return ErrOutOfRange
	}
	pos := offset % int64(len(r.data))
	written := 0
	for written < len(val) {
		written = copy(r.data[pos:], val[written:])
		r.end += int64(written)
		if pos >= int64(len(r.data)) {
			pos -= int64(len(r.data))
		}
	}
	if int(r.end-r.begin) > len(r.data) {
		r.begin = r.end - int64(len(r.data))
	}
	return nil
}

// 读入len(ret)长度的内容
func (r *ringBuf) ReadAt(ret []byte, offset int64, length int) error {
	if offset < r.begin || offset > r.end {
		return ErrOutOfRange
	}
	pos := int(offset % int64(len(r.data)))
	if pos+length < len(r.data) {
		copy(ret, r.data[pos:pos+length])
		return nil
	}
	n := copy(ret, r.data[pos:])
	if n < length {
		copy(ret[n:], r.data[:length-n])
	}
	return nil
}

// 清理部分空间，由于写入是往尾部写入，所以要往头部挤
func (r *ringBuf) Evacuate(offset int64, length int) error {
	if offset < r.begin || offset+int64(length) > r.end {
		return ErrOutOfRange
	}
	b := make([]byte, offset-r.begin)
	_ = r.ReadAt(b, r.begin, len(b))
	_ = r.WriteAt(b, r.begin+int64(length))
	r.begin += int64(length)
	return nil
}

func (r *ringBuf) EqualAt(val []byte, offset int64) bool {
	if offset+int64(len(val)) > r.end || offset < r.begin {
		return false
	}
	pos := int(offset % int64(len(r.data)))
	if pos+len(val) < len(r.data) {
		return bytes.Equal(val, r.data[pos:pos+len(val)])
	}
	beginPos := int(r.begin % int64(len(r.data)))
	equal := bytes.Equal(val[:len(r.data)-beginPos], r.data[pos:])
	if equal {
		equal = bytes.Equal(val[len(r.data)-beginPos:], r.data[:len(val)-len(r.data)+beginPos])
	}
	return equal
}

func (r *ringBuf) CheckSpace(size int) bool {
	return len(r.data)-int(r.end-r.begin) >= size
}
