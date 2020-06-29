package quickcache

import (
	"bytes"
)

type ringBuf struct {
	begin int64
	end   int64
	size  int
	data  []byte
}

func newRingBuf(size int) ringBuf {
	rb := ringBuf{
		begin: 0,
		end:   0,
		size:  size,
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
	return r.size
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
	pos := offset & int64(r.size-1)
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
	pos := int(offset & int64(r.size-1))
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

	if r.begin == offset {
		r.begin += int64(length)
		return nil
	}

	beginPos := int(r.begin & int64(r.size-1))
	evaBeginPos := int(offset & int64(r.size-1))
	evaEndPos := evaBeginPos + length
	var n int
	if evaEndPos >= len(r.data) { // 左右两侧都有
		// 先复制左端
		m := evaEndPos - len(r.data)
		n = rightAlignCopy(r.data[:m], r.data[beginPos:evaBeginPos])
		if n >= m {
			// 再复制右端
			n += rightAlignCopy(r.data[beginPos+length:len(r.data)], r.data[beginPos:evaBeginPos-m])
		}
	} else if evaBeginPos < beginPos { // 全在左侧
		n = rightAlignCopy(r.data[:evaEndPos], r.data[:evaBeginPos])
		n += rightAlignCopy(r.data[:n], r.data[beginPos:])
	} else { // 全在右侧
		n = rightAlignCopy(r.data[beginPos:evaBeginPos+length], r.data[beginPos:evaBeginPos])
	}
	r.begin += int64(length)
	return nil
}

func (r *ringBuf) EqualAt(val []byte, offset int64) bool {
	if offset+int64(len(val)) > r.end || offset < r.begin {
		return false
	}
	pos := int(offset & int64(r.size-1))
	if pos+len(val) < len(r.data) {
		return bytes.Equal(val, r.data[pos:pos+len(val)])
	}
	beginPos := int(r.begin & int64(r.size-1))
	equal := bytes.Equal(val[:len(r.data)-beginPos], r.data[pos:])
	if equal {
		equal = bytes.Equal(val[len(r.data)-beginPos:], r.data[:len(val)-len(r.data)+beginPos])
	}
	return equal
}

func (r *ringBuf) CheckSpace(size int) int {
	return size - len(r.data) + int(r.end-r.begin)
}
