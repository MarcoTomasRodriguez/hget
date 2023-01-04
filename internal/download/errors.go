package download

import (
	"errors"
	"fmt"
)

var (
	NonexistentDownloadErr   = errors.New("download does not exist")
	UserCancelledDownloadErr = errors.New("user cancelled download")
	BrokenDownloadErr        = errors.New("download is broken")
	InvalidFilenameErr       = errors.New("invalid filename")
	SegmentOverflowErr       = errors.New("segment overflow")
	BufferCopyErr            = errors.New("could not copy buffer")
)

type NetworkError string

func (e NetworkError) Error() string {
	return fmt.Sprintf("network error: %s", string(e))
}

type FilesystemError string

func (e FilesystemError) Error() string {
	return fmt.Sprintf("filesystem error: %s", string(e))
}
