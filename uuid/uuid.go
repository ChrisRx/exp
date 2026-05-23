package uuid

import (
	"encoding/hex"
	"hash/fnv"
)

// NewV8 returns an RFC4122 compliant UUID (Version 8). The resulting UUID will
// be deterministic based upon the provided input string.
//
// The input is hashed using the 128-bit version of the FNV-1 hash algorithm.
func NewV8(input string) string {
	h := fnv.New128()
	_, _ = h.Write([]byte(input))
	hash := h.Sum(nil)
	var u [16]byte
	copy(u[:], hash[:])
	u[6] = (u[6] & 0x0f) | 0x80 // Version 8       (0b1000)
	u[8] = (u[8] & 0x3f) | 0x80 // Version RFC4122 (0b10)
	var buf [36]byte
	hex.Encode(buf[:8], u[:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], u[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], u[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], u[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], u[10:])
	return string(buf[:])
}
