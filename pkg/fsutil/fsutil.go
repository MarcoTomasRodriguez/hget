package fsutil

import (
	"fmt"
	"regexp"
)

const (
	SI = 1000
	KB = SI
	MB = SI * KB
	GB = SI * MB
	TB = SI * GB
	PB = SI * TB
)

type UnsignedNumber interface {
	uint | uint16 | uint32 | uint64
}

// ValidateFilename checks the validity of a filename.
func ValidateFilename(filename string) bool {
	valid, _ := regexp.MatchString("^(\\w+[\\w\\-.]+\\.\\w+){1,255}$", filename)

	return valid
}

// ReadableMemorySize returns a prettier form of some memory size expressed in bytes.
// Note: Do not exceed the float64 limit, as the result will overflow.
func ReadableMemorySize[T UnsignedNumber](bytes T) string {
	if bytes < SI {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := uint64(SI), 0
	for n := bytes / SI; n >= SI; n /= SI {
		div *= SI
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "kMGTPE"[exp])
}