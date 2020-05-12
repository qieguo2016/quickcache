package quickcache

import (
	"errors"
	"unsafe"
)

const (
	bucketCount     = 256
	entryHeaderSize = 16 // header的长度是128bit，16byte
)

var ErrLargeKey = errors.New("The key is larger than 65535")
var ErrLargeEntry = errors.New("The entry size is larger than 1/1024 of cache size")
var ErrNotFound = errors.New("Entry not found")

type segment struct {
	segId     int
	rb        ringBuf
	bucketLen [bucketCount]int32 // 拉链法中的数组
	bucketCap int32
	buckets   []entry // 拉链法中的数组+链表，链表可以看成展开在数组内部
}

type entry struct {
	offset int64
	hash16 uint16
	keyLen uint16
	valLen uint32
}

// 长度是entryHeaderSize
type entryHeader struct {
	hash16   uint16
	keyLen   uint16
	valLen   uint32
	valCap   uint32
	deleted  bool
	bucketId uint8
	reserved uint16 // 64bit，对齐cache line
}

func newSegment(size, i int) segment {
	return segment{
		segId: i,
	}
}

func (s *segment) Set(key, value []byte, hash uint64) error {
	if len(key) > 65535 {
		return ErrLargeKey
	}
	maxKeyValLen := len(s.rb.data)/4 - entryHeaderSize
	if len(key)+len(value) > maxKeyValLen {
		// Do not accept large entry.
		return ErrLargeEntry
	}
	bucketId := uint8(hash >> 8)    // bucket序号
	bucket := s.getBucket(bucketId) // 连续数组代表链表
	hash16 := uint16(hash >> 16)
	//header.hash16 = hash16
	//header.keyLen = uint16(len(key))
	//header.valLen = uint32(len(value))
	//header.deleted = false
	var headerBuf [entryHeaderSize]byte
	header := (*entryHeader)(unsafe.Pointer(&headerBuf[0])) // 共用内存地址，方便byte转换
	pos, found := s.lookup(bucket, hash16, key)             // key是在rb上，读取不便，所以用hash16加速查询
	if found {
		entry := &bucket[pos]
		err := s.rb.ReadAt(headerBuf[:], entry.offset) // 复用[]byte空间
		if err != nil {
			return err
		}
		header.hash16 = hash16
		header.keyLen = uint16(len(key))
		header.valLen = uint32(len(value))
		header.deleted = false
		header.bucketId = bucketId
		if header.valCap > uint32(len(value)) { // rb中该位置有足够空间
			// 原子性？相同的key，不考虑
			_ = s.rb.WriteAt(headerBuf[:], entry.offset)
			_ = s.rb.WriteAt(value, entry.offset+entryHeaderSize+int64(header.keyLen))
			return nil
		}
		// 不够空间则清理该段空间，将空间挤到头部或者尾部，再从rb尾部插入
	} else {
		// new
	}
	// 从尾部插入，也需要判断是否有足够空间
	// 够，直接插入
	// 不够，则要淘汰掉老数据，1.从
}

func (s *segment) Get(key []byte, hash uint64) ([]byte, error) {

}

func (s *segment) Del(key []byte, hash uint64) error {

}

func (s *segment) expand() {

}

func (s *segment) addEntry() {

}

func (s *segment) getBucket(idx uint8) []entry {
	offset := int32(idx) * s.bucketCap
	return s.buckets[offset : offset+s.bucketLen[idx] : offset+s.bucketCap]
}

func (s *segment) lookup(bucket []entry, hash16 uint16, key []byte) (idx int, match bool) {
	// todo: hash16是按倍数递增的，所以可以使用二分查找
	for idx < len(bucket) {
		entry := &bucket[idx]
		if entry.hash16 != hash16 {
			continue
		}
		match = int(entry.keyLen) == len(key) && s.rb.EqualAt(key, entry.offset+entryHeaderSize)
		if match {
			return
		}
		idx++
	}
	return
}
