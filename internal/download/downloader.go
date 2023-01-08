package download

import (
	"context"
	"errors"
	"fmt"
	"github.com/MarcoTomasRodriguez/hget/pkg/logger"
	"github.com/MarcoTomasRodriguez/hget/pkg/progressbar"
	"github.com/fatih/color"
	"io"
	"math/rand"
	"sync"
)

var UserCancelledDownloadErr = errors.New("user cancelled download")

type Downloader interface {
	Download(download Download, ctx context.Context) error
	InitDownload(url string, workers uint8) (Download, error)
	FindAllDownloads() ([]Download, error)
	FindDownloadById(id string) (Download, error)
	FindDownloadByUrl(url string) (Download, error)
	DeleteDownloadById(id string) error
}

type downloader struct {
	network     Network
	storage     Storage
	progressbar progressbar.ProgressBar
	logger      logger.Logger
}

// InitDownload extracts the download specification from a web resource.
func (s downloader) InitDownload(url string, workers uint8) (Download, error) {
	resource, err := s.network.FetchResource(url)
	if err != nil {
		return Download{}, err
	}

	// Generate the download id.
	id := make([]byte, 4)
	rand.Read(id)

	// In order for range downloads to work, they should be supported and the content length be provided.
	var segments []Segment
	if resource.Size <= 0 || !resource.AcceptRanges {
		segments = make([]Segment, 1)
	} else {
		segments = make([]Segment, workers)
	}

	// Initialize segments.
	for i := range segments {
		// Compute the segment's starting point.
		start := (resource.Size / int64(len(segments))) * int64(i)

		// Initialize the segment's end point. By default, it is the file size.
		end := resource.Size

		// If the segment is not the last, compute his end point.
		if i < len(segments)-1 {
			end = (resource.Size/int64(len(segments)))*(int64(i)+1) - 1
		}

		segments[i] = Segment{
			Id:    fmt.Sprintf("%x/segment.%02d", id, i),
			Start: start,
			End:   end,
		}
	}

	return Download{
		Id:       fmt.Sprintf("%x", id),
		Name:     resource.Filename,
		URL:      resource.URL,
		Size:     resource.Size,
		Segments: segments,
	}, nil
}

// Download takes a download specification and downloads it.
func (s downloader) Download(download Download, ctx context.Context) error {
	var wg sync.WaitGroup

	if err := s.storage.WriteDownloadSpec(download); err != nil {
		return err
	}

	// Create a channel to listen to the workers' return error.
	workerErrors := make(chan error)

	for i, segment := range download.Segments {
		// Check if segment download already finished.
		segmentSize, _ := s.storage.GetSegmentSize(segment.Id)
		segmentOffset := segment.Start + segmentSize
		if segmentOffset >= segment.End {
			continue
		}

		// Add progress bar to pool.
		total := segment.End - segmentOffset
		prefix := color.CyanString(fmt.Sprintf("Worker #%d", i))
		progressWriter := s.progressbar.Add(total, progressbar.Bytes, prefix)

		// Worker thread.
		wg.Add(1)
		go func(segment Segment, progressWriter io.Writer, segmentOffset int64) {
			defer wg.Done()

			segmentWriter, err := s.storage.OpenSegment(segment.Id)
			if err != nil {
				workerErrors <- err
				return
			}

			defer func() { _ = segmentWriter.Close() }()

			writer := io.MultiWriter(segmentWriter, progressWriter)
			if err := s.network.DownloadResource(download.URL, segmentOffset, segment.End, writer, ctx); err != nil {
				workerErrors <- err
			}
		}(segment, progressWriter, segmentOffset)
	}

	_ = s.progressbar.Start()
	defer func() { _ = s.progressbar.Stop() }()

	waitGroupDone := make(chan struct{})
	go func() {
		defer close(waitGroupDone)
		wg.Wait()
	}()

readChannels:
	for {
		select {
		case err := <-workerErrors:
			return err
		case <-ctx.Done():
			return UserCancelledDownloadErr
		case <-waitGroupDone:
			break readChannels
		}
	}

	// Open output file in write-only mode with permissions: -rw-r--r--.
	downloadWriter, err := s.storage.OpenDownloadOutput(download.Id)
	if err != nil {
		return FilesystemError(err.Error())
	}

	defer func() { _ = downloadWriter.Close() }()

	// Join the segments into the output file.
	s.logger.Info("Merging...")
	for _, segment := range download.Segments {
		// Open segment file.
		segmentReader, err := s.storage.OpenSegment(segment.Id)
		if err != nil {
			return FilesystemError(err.Error())
		}

		// Append worker file to output file.
		if _, err = io.Copy(downloadWriter, segmentReader); err != nil {
			return BufferCopyErr
		}

		// Remove segment file.
		if err := s.storage.DeleteSegment(segment.Id); err != nil {
			return FilesystemError(err.Error())
		}
	}

	return nil
}

// FindAllDownloads finds valid download specifications.
func (s downloader) FindAllDownloads() ([]Download, error) {
	return s.storage.ListDownloads()
}

// FindDownloadById finds a download specification by its id.
func (s downloader) FindDownloadById(id string) (Download, error) {
	return s.storage.ReadDownloadSpec(id)
}

// FindDownloadByUrl finds a download specification by its url.
func (s downloader) FindDownloadByUrl(url string) (Download, error) {
	downloads, err := s.storage.ListDownloads()
	if err != nil {
		return Download{}, nil
	}

	for _, download := range downloads {
		if download.URL == url {
			return download, nil
		}
	}

	return Download{}, nil
}

// DeleteDownloadById deletes the download folder including the specification file, downloaded segments and the
// eventual merged result.
func (s downloader) DeleteDownloadById(id string) error {
	return s.storage.DeleteDownload(id)
}

// NewDownloader instantiates a new Downloader object.
func NewDownloader(network Network, storage Storage, progressbar progressbar.ProgressBar, logger logger.Logger) Downloader {
	return &downloader{network, storage, progressbar, logger}
}

var _ Downloader = (*downloader)(nil)
