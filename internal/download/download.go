package download

import (
	"fmt"
	"github.com/MarcoTomasRodriguez/hget/pkg/fsutil"
	"github.com/MarcoTomasRodriguez/hget/pkg/httputil"
	"github.com/fatih/color"
	"math/rand"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
)

type Download struct {
	Id       string    `yaml:"id"`
	Name     string    `yaml:"name"`
	URL      string    `yaml:"url"`
	Size     int64     `yaml:"size"`
	Segments []Segment `yaml:"segments"`
}

func (f *Download) String() string {
	return fmt.Sprintln(
		" ⁕", color.HiCyanString(f.Id), "⇒",
		color.HiCyanString("URL:"), f.URL,
		color.HiCyanString("Size:"), fsutil.ReadableMemorySize(f.Size),
	)
}

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
		return nil, InvalidFilenameError{}
	}

	// Download http request.
	response, err := http.Get(file.URL)
	if err != nil {
		return nil, err
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
		return nil, InvalidFilenameError{}
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
