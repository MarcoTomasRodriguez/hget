package download

import "fmt"

type NonexistentDownloadError struct{}

func (e NonexistentDownloadError) Error() string {
	return fmt.Sprintf("download does not exist")
}

type BrokenDownloadError struct{}

func (e BrokenDownloadError) Error() string {
	return fmt.Sprintf("download is broken")
}

type InvalidFilenameError struct{}

func (e InvalidFilenameError) Error() string {
	return "invalid filename"
}

type WorkerError string

func (e WorkerError) Error() string {
	return fmt.Sprintf("worker error: %s", string(e))
}

type CancelledDownloadError string

func (e CancelledDownloadError) Error() string {
	return fmt.Sprintf("cancelled download: %s", string(e))
}

type SegmentOverflowError struct{}

func (e SegmentOverflowError) Error() string {
	return "segment overflow"
}

type NetworkError string

func (e NetworkError) Error() string {
	return fmt.Sprintf("network error: %s", string(e))
}

type FilesystemError string

func (e FilesystemError) Error() string {
	return fmt.Sprintf("filesystem error: %s", string(e))
}

type IOCopyError string

func (e IOCopyError) Error() string {
	return fmt.Sprintf("buffer copy error: %s", string(e))
}
