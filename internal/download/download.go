package download

import (
	"context"
	"fmt"
	"github.com/MarcoTomasRodriguez/hget/pkg/fsutil"
	"github.com/MarcoTomasRodriguez/hget/pkg/httputil"
	"github.com/cheggaaa/pb"
	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
	"github.com/spf13/afero"
	"io"
	"math/rand"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
)

// Download stores the information of a resource that can be downloaded.
type Download struct {
	Id       string    `yaml:"id"`
	Name     string    `yaml:"name"`
	URL      string    `yaml:"url"`
	Size     int64     `yaml:"size"`
	Segments []Segment `yaml:"segments"`
}

// String returns a colored formatted string with the download's Id, URL and Size.
func (d *Download) String() string {
	return fmt.Sprintln(
		" ⁕", color.HiCyanString(d.Id), "⇒",
		color.HiCyanString("URL:"), d.URL,
		color.HiCyanString("Size:"), fsutil.ReadableMemorySize(d.Size),
	)
}

func (d *Download) Download(afs afero.Afero, ctx context.Context) error {
	var wg sync.WaitGroup
	var progressBars []*pb.ProgressBar

	// Create a channel to listen to the workers' return error.
	workerErrors := make(chan error)

	// Progress bar utilities.
	showProgressBar := d.Size > 0 && isatty.IsTerminal(os.Stdout.Fd())

	for i, segment := range d.Segments {
		var progressBar *pb.ProgressBar

		// Create the segment file with permissions: -rw-r--r--.
		segmentFile, err := afs.OpenFile(segment.Filename(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return FilesystemError(err.Error())
		}

		segmentStat, err := segmentFile.Stat()
		segmentOffset := segment.Start + segmentStat.Size()
		if segmentOffset >= segment.End {
			continue
		}

		// Add progress bar to pool.
		if showProgressBar {
			progressBar = pb.New64(segment.End - segmentOffset).SetUnits(pb.U_BYTES).Prefix(
				color.CyanString(fmt.Sprintf("Worker #%d", i)),
			)
			progressBars = append(progressBars, progressBar)
		}

		// Worker thread.
		wg.Add(1)
		go func(segment Segment, segmentFile afero.File, progressBar *pb.ProgressBar, segmentOffset int64) {
			var segmentWriter io.Writer = segmentFile

			defer wg.Done()
			defer segmentFile.Close()

			if progressBar != nil {
				defer progressBar.Finish()
				segmentWriter = io.MultiWriter(segmentWriter, progressBar)
			}

			if err := segment.Download(d.URL, segmentOffset, segmentWriter, ctx); err != nil {
				workerErrors <- err
			}
		}(segment, segmentFile, progressBar, segmentOffset)
	}

	// Start progress bar.
	if showProgressBar {
		pool, _ := pb.StartPool(progressBars...)
		defer func() { _ = pool.Stop() }()
	}

	waitGroupDone := make(chan struct{})
	go func() {
		defer close(waitGroupDone)
		wg.Wait()
	}()

	contextDone := ctx.Done()

readChannels:
	for {
		select {
		case err := <-workerErrors:
			return err
		case <-contextDone:
			return UserCancelledDownloadErr
		case <-waitGroupDone:
			break readChannels
		}
	}

	// Open output file in write-only mode with permissions: -rw-r--r--.
	outputFile, err := afs.OpenFile(d.Name, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return FilesystemError(err.Error())
	}

	defer outputFile.Close()

	// Setup progress progressBar.
	var progressBar *pb.ProgressBar
	if showProgressBar {
		progressBar = pb.StartNew(len(d.Segments)).Prefix(color.CyanString("Merging"))
		defer progressBar.Finish()
	}

	// Join the segments into the output file.
	for _, s := range d.Segments {
		// Open segment file.
		segmentFile, err := afs.Open(s.Filename())
		if err != nil {
			return FilesystemError(err.Error())
		}

		// Append worker file to output file.
		if _, err = io.Copy(outputFile, segmentFile); err != nil {
			return BufferCopyErr
		}

		// Remove segment file.
		if err := afs.Remove(s.Filename()); err != nil {
			return FilesystemError(err.Error())
		}

		if showProgressBar {
			progressBar.Increment()
		}
	}

	return nil
}

// NewDownload initializes the download object by executing a GET request to the resource and analyzing its headers.
func NewDownload(rawUrl string, segmentsNum uint8) (*Download, error) {
	var err error
	file := &Download{}

	// Resolve url.
	file.URL, err = httputil.ResolveURL(rawUrl)
	if err != nil {
		return nil, err
	}

	// Extract the download filename from the headers or from the url.
	path, err := url.PathUnescape(file.URL)
	if err != nil {
		return nil, InvalidFilenameErr
	}

	// Download http request.
	response, err := http.Get(file.URL)
	if err != nil {
		return nil, NetworkError(err.Error())
	}

	defer response.Body.Close()

	// Extract Content-Length and Accept-Ranges from response headers.
	file.Size = response.ContentLength
	acceptRanges := response.Header.Get("Accept-Ranges")

	// In order for range downloads to work, they should be supported and the content length be provided.
	if file.Size <= 0 || acceptRanges != "bytes" {
		file.Segments = make([]Segment, 1)
	} else {
		file.Segments = make([]Segment, segmentsNum)
	}

	file.Name = filepath.Base(path)
	_, contentDispositionParams, err := mime.ParseMediaType(response.Header.Get("Content-Disposition"))
	if filename, ok := contentDispositionParams["filename"]; ok {
		file.Name = filename
	}

	// Validate the download filename.
	if !fsutil.ValidateFilename(file.Name) {
		return nil, InvalidFilenameErr
	}

	// Generate the download file id.
	id := make([]byte, 4)
	rand.Read(id)
	file.Id = fmt.Sprintf("%x", id)

	// Initialize segments.
	for i := range file.Segments {
		// Compute the segment's starting point.
		start := (file.Size / int64(len(file.Segments))) * int64(i)

		// Initialize the segment's end point.
		// By default, it is the download size.
		end := file.Size

		// If the segment is not the last, compute his end point.
		if i < len(file.Segments)-1 {
			end = (file.Size/int64(len(file.Segments)))*(int64(i)+1) - 1
		}

		file.Segments[i] = Segment{Id: uint8(i), Start: start, End: end}
	}

	return file, nil
}
