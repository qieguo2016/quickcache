package quickcache

import (
	"github.com/cespare/xxhash"
)

// hash分布：|31-16用来加速查询|15-8用来定位bucket|7-0用来定位segment|
func hashFunc(data []byte) uint64 {
	return xxhash.Sum64(data)
}

func convertMBToBytes(value int) int {
	return value * 1024 * 1024
}

func isPowerOfTwo(number int) bool {
	return (number & (number - 1)) == 0
}

// 右对齐copy
func rightAlignCopy(dst, src []byte) int {
	if len(dst) > len(src) {
		return copy(dst[len(dst)-len(src):], src)
	} else {
		return copy(dst, src[len(src)-len(dst):])
	}
}