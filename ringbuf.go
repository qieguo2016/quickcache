package quickcache

import (
	"errors"
)

var ErrOutOfRange = errors.New("out of range")

type ringBuf struct {
	begin int64
	end   int64
	data  []byte
}

func newRingBuf(size int) *ringBuf {
	rb := ringBuf{
		begin: 0,
		end:   0,
		data:  make([]byte, size),
	}
	return &rb
}

func (r *ringBuf) Write(val []byte) error {
	if len(p) > len(rb.data) {
		err = ErrOutOfRange
		return
	}
	for n < len(p) {
		written := copy(rb.data[rb.index:], p[n:])
		rb.end += int64(written)
		n += written
		rb.index += written
		if rb.index >= len(rb.data) {
			rb.index -= len(rb.data)
		}
	}
	if int(rb.end-rb.begin) > len(rb.data) {
		rb.begin = rb.end - int64(len(rb.data))
	}
	return
}

func (r *ringBuf) WriteAt(val []byte, offset int64) error {

}


func (r *ringBuf) Read(offset int64, limit int64) ([]byte, error) {

}

func (r *ringBuf) ReadAt(ret []byte, offset int64) error {

}


func (r *ringBuf) Evacuate(offset, length int64) {

}

func (r *ringBuf) EqualAt(val []byte, offset int64) bool {

}

func (r *ringBuf) CheckSpace(size int) bool {
	return len(r.data) - int(r.end - r.begin) >= size
}