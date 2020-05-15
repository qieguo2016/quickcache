package quickcache

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
		written := copy(r.data[pos:], val[written:])
		r.end += int64(written)
		if pos >= int64(len(r.data)) {
			pos -= int64(len(r.data))
		}
	}
	return nil
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