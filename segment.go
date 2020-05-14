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
	bucketCap int32              // 拉链法中的链表长度
	buckets   []entryIndex       // 拉链法中的数组+链表，链表可以看成展开在数组内部
}

type entryIndex struct {
	offset      int64
	hash16      uint16
	keyLen      uint16
	valCap      uint32
	accessTime  uint32
	expireTime  uint32
	accessCount uint32
	reserved32 uint32 // 64bit，对齐cache line
}

// 长度是entryHeaderSize
type entryHeader struct {
	valLen     uint32
	valCap     uint32
	keyLen     uint16
	reserved16 uint16
	reserved32 uint32 // 64bit，对齐cache line
}

func newSegment(size, i int) segment {
	return segment{
		segId: i,
	}
}

func (s *segment) Set(key, value []byte, hash uint64) error {
	// pre check
	if len(key) > 65535 {
		return ErrLargeKey
	}
	maxKeyValLen := len(s.rb.data)/4 - entryHeaderSize
	if len(key)+len(value) > maxKeyValLen {
		return ErrLargeEntry
	}

	bucketId := uint8(hash >> 8)    // bucket序号
	bucket := s.getBucket(bucketId) // 连续数组代表链表，作为一个桶
	hash16 := uint16(hash >> 16)
	var headerBuf [entryHeaderSize]byte
	header := (*entryHeader)(unsafe.Pointer(&headerBuf[0])) // 共用内存地址，方便byte转换
	pos, found := s.getIndexOfBucket(bucket, hash16, key)   // key是在rb上，读取不便，所以用hash16加速查询
	if found {
		entryIdx := bucket[pos]
		_ = s.rb.ReadAt(headerBuf[:], entryIdx.offset) // 复用[]byte空间
		header.keyLen = uint16(len(key))
		header.valLen = uint32(len(value))
		if header.valCap > uint32(len(value)) { // rb中该位置有足够空间
			// 原子性？相同的key，不考虑
			_ = s.rb.WriteAt(headerBuf[:], entryIdx.offset)
			_ = s.rb.WriteAt(value, entryIdx.offset+entryHeaderSize+int64(header.keyLen))
			return nil
		}
		// 不够空间则清理该段空间，将空间挤到头部或者尾部
		header.valCap = header.valLen
		s.delEntryIndex(bucketId, bucket, pos)
		s.rb.Evacuate(entryIdx.offset, entryHeaderSize+int64(header.keyLen)+int64(header.valCap))
	} else {
		// new
		header.keyLen = uint16(len(key))
		header.valLen = uint32(len(value))
		header.valCap = header.valLen
	}

	// 从尾部插入，也需要判断是否有足够空间
	enough := s.rb.CheckSpace(entryHeaderSize + int(header.keyLen) + int(header.valCap))
	if !enough {
		// 不够空间，需要淘汰掉老数据
		s.evacuate(entryHeaderSize + int(header.keyLen) + int(header.valCap))
	}
	s.addEntryIndex(bucketId, pos, s.rb.end, hash16, header.keyLen, header.valCap)
	_ = s.rb.Write(headerBuf[:])
	_ = s.rb.Write(key)
	_ = s.rb.Write(value)
	return nil
}

func (s *segment) Get(key []byte, hash uint64) (ret []byte, err error) {
	bucketId := uint8(hash >> 8)    // bucket序号
	bucket := s.getBucket(bucketId) // 连续数组代表链表，作为一个桶
	hash16 := uint16(hash >> 16)
	pos, found := s.getIndexOfBucket(bucket, hash16, key) // key是在rb上，读取不便，所以用hash16加速查询
	if !found {
		err = ErrNotFound
		return
	}
	entryIdx := bucket[pos]
	err = s.rb.ReadAt(ret, entryIdx.offset+entryHeaderSize+int64(entryIdx.keyLen))
	return
}

func (s *segment) Del(key []byte, hash uint64) error {
	bucketId := uint8(hash >> 8)    // bucket序号
	bucket := s.getBucket(bucketId) // 连续数组代表链表，作为一个桶
	hash16 := uint16(hash >> 16)
	pos, found := s.getIndexOfBucket(bucket, hash16, key) // key是在rb上，读取不便，所以用hash16加速查询
	if !found {
		return ErrNotFound
	}
	entryIdx := bucket[pos]
	s.delEntryIndex(bucketId, bucket, pos)
	s.rb.Evacuate(entryIdx.offset, entryHeaderSize+int64(entryIdx.keyLen)+int64(entryIdx.valCap))
	return nil
}

func (s *segment) addEntryIndex(bucketId uint8, pos int, offset int64, hash16 uint16, keyLen uint16, valCap uint32) {
	if s.bucketLen[bucketId] == s.bucketCap {
		s.expand()
	}
	s.bucketLen[bucketId]++
	bucket := s.getBucket(bucketId)
	copy(bucket[pos+1:], bucket[pos:]) // 插入到pos位置
	bucket[pos].offset = offset
	bucket[pos].hash16 = hash16
	bucket[pos].keyLen = keyLen
	bucket[pos].valCap = valCap
}

func (s *segment) delEntryIndex(bucketId uint8, bucket []entryIndex, pos int) {
	copy(bucket[pos:], bucket[pos+1:]) // 删除pos位置
	s.bucketLen[bucketId]--
}

// 拉链索引2倍扩容
func (s *segment) expand() {
	newBuckets := make([]entryIndex, s.bucketCap*bucketCount*2)
	for i := 0; i < bucketCount; i++ {
		offset := int32(i) * s.bucketCap
		copy(newBuckets[offset*2:], s.buckets[offset:offset+s.bucketLen[i]])
	}
	s.bucketCap *= 2
	s.buckets = newBuckets
}

// 驱逐元素获取足够空间
func (s *segment) evacuate(size int) {

}

// 从数组里面切出一个bucket
func (s *segment) getBucket(idx uint8) []entryIndex {
	offset := int32(idx) * s.bucketCap
	return s.buckets[offset : offset+s.bucketLen[idx] : offset+s.bucketCap]
}

func (s *segment) getIndexOfBucket(bucket []entryIndex, hash16 uint16, key []byte) (idx int, match bool) {
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
