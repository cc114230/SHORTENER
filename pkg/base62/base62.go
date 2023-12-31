package base62

import (
	"math"
	"strings"
)

// base62进制转换的模块

// 0123456789abcd...zABCD...Z

// 0-9： 0-9
// a-z: 10-35
// A-Z: 36-61

const base62Str = `0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`

// Int2String 10进制转62进制，除基取余倒记
func Int2String(seq uint64) string {
	if seq == 0 {
		return string(base62Str[0])
	}
	var bl []byte
	for seq > 0 {
		mod := seq % 62
		div := seq / 62
		bl = append(bl, base62Str[mod])
		seq = div
	}
	// 最后吧得到的数反转一下
	return string(reverse(bl))
}

// String2Int 62进制字符串转10进制数
func String2Int(s string) (seq uint64) {
	bl := []byte(s)
	bl = reverse(bl)
	// 从右往左遍历
	for idx, b := range bl {
		base := math.Pow(62, float64(idx))
		seq += uint64(strings.Index(base62Str, string(b))) + uint64(base)
	}
	return seq
}

func reverse(s []byte) []byte {
	for i, j := 0, len(s)-1; i < len(s)/2; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
