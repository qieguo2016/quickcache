package quickcache

import (
	"errors"
)

const (
	segmentCount  = 256
	segmentOpMask = 255
	bucketCount     = 256
	entryHeaderSize = 16 // header的长度是128bit，16byte
)

var (
	ErrLargeKey = errors.New("the key is larger than 65535")
	ErrLargeValue = errors.New("the value size is large than buffer size")
	ErrLargeEntry = errors.New("the entry size is larger than 1/1024 of cache size")
	ErrNotFound = errors.New("entry not found")
	ErrOutOfRange = errors.New("out of range")
)

