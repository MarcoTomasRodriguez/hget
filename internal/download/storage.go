package download

import (
	"errors"
	"fmt"
	"github.com/MarcoTomasRodriguez/hget/pkg/codec"
	"github.com/samber/lo"
	"github.com/spf13/afero"
	"io"
	ioFs "io/fs"
	"os"
	"path/filepath"
)

var (
	BrokenDownloadErr = errors.New("download is broken")
)

type Storage interface {
	ListDownloads() ([]Download, error)
	ReadDownloadSpec(id string) (Download, error)
	WriteDownloadSpec(download Download) error
	OpenDownloadOutput(id string) (io.ReadWriteCloser, error)
	DeleteDownload(id string) error
	OpenSegment(id string) (io.ReadWriteCloser, error)
	GetSegmentSize(id string) (int64, error)
	DeleteSegment(id string) error
}

type FilesystemError string

func (e FilesystemError) Error() string {
	return fmt.Sprintf("storage error: %s", string(e))
}

type storage struct {
	afs   afero.Afero
	codec codec.Codec
}

// ListDownloads lists the download specifications from the filesystem.
func (f storage) ListDownloads() ([]Download, error) {
	// List elements inside the internal download directory.
	downloadFolders, err := f.afs.ReadDir(".")
	if err != nil {
		return nil, err
	}

	// Iterate over the elements inside the download folder and read them.
	downloads := lo.FilterMap(downloadFolders, func(fi ioFs.FileInfo, _ int) (Download, bool) {
		d, err := f.ReadDownloadSpec(fi.Name())
		if err != nil {
			return d, false
		}

		return d, true
	})

	return downloads, nil
}

// ReadDownloadSpec reads the download specification from the filesystem.
func (f storage) ReadDownloadSpec(id string) (Download, error) {
	download := Download{}

	// Read download specification.
	in, err := f.afs.ReadFile(filepath.Join(id, "download."+f.codec.Extension()))
	if err != nil {
		return Download{}, BrokenDownloadErr
	}

	// Unmarshal encoded download.
	if err := f.codec.Unmarshal(in, &download); err != nil {
		return Download{}, BrokenDownloadErr
	}

	return download, nil
}

// WriteDownloadSpec saves the download specification on the filesystem.
func (f storage) WriteDownloadSpec(download Download) error {
	_ = f.afs.MkdirAll(download.Id, 0755)

	// Marshall download.
	out, err := f.codec.Marshal(download)
	if err != nil {
		return err
	}

	// Open download specification.
	return f.afs.WriteFile(filepath.Join(download.Id, "download."+f.codec.Extension()), out, 0644)
}

// OpenDownloadOutput opens the download output file by id and returns the file for read and write.
func (f storage) OpenDownloadOutput(id string) (io.ReadWriteCloser, error) {
	return f.afs.OpenFile(filepath.Join(id, "output"), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
}

// DeleteDownload deletes the whole download folder from the filesystem.
func (f storage) DeleteDownload(id string) error {
	return f.afs.RemoveAll(id)
}

// OpenSegment opens the size of a segment by id and returns the file for read and write.
func (f storage) OpenSegment(id string) (io.ReadWriteCloser, error) {
	return f.afs.OpenFile(id, os.O_CREATE|os.O_RDWR, 0644)
}

// GetSegmentSize gets the size of a segment by id.
func (f storage) GetSegmentSize(id string) (int64, error) {
	fileInfo, err := f.afs.Stat(id)
	if err != nil {
		return 0, err
	}

	return fileInfo.Size(), nil
}

// DeleteSegment deletes a segment file by id from the filesystem.
func (f storage) DeleteSegment(id string) error {
	return f.afs.Remove(id)
}

// NewStorage instantiates a new Storage object.
func NewStorage(fs afero.Fs, codec codec.Codec) Storage {
	return storage{afs: afero.Afero{Fs: fs}, codec: codec}
}

var _ Storage = (*storage)(nil)
