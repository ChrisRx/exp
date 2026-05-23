package random

import (
	"crypto/rand"
	"fmt"
)

var testOnlyRejectionSampling func(int)

func String(length int) string {
	const chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// The provided length must be positive. If length is negative/zero it will
	// result in a zero value string which could go unnoticed. Instead of
	// returning an error, this is setting a minimum length value of 32.
	length = max(length, 32)
	// Only random integers in range [0, rejectAfter) should be used to avoid
	// modulo bias.
	rejectAfter := 255 - (256 % len(chars))

	output := make([]byte, length)
	for i := 0; i < length; {
		buf := make([]byte, 1)
		// Read uses [io.ReadFull] internally, so this will always produce an error
		// if it reads less than len(buf).
		if _, err := rand.Read(buf); err != nil {
			panic(fmt.Errorf("cannot read from crypto/rand.Read: %v", err))
		}
		c := int(buf[0])
		if c > rejectAfter {
			continue
		}
		output[i] = chars[c%len(chars)]
		i++

		if testOnlyRejectionSampling != nil {
			testOnlyRejectionSampling(c % len(chars))
		}
	}
	return string(output)
}
