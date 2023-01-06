package download

import (
	"context"
	"errors"
	"fmt"
	"github.com/MarcoTomasRodriguez/hget/pkg/fsutil"
	"github.com/MarcoTomasRodriguez/hget/pkg/httputil"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
)

var (
	InvalidFilenameErr = errors.New("invalid filename")
	SegmentOverflowErr = errors.New("segment overflow")
	BufferCopyErr      = errors.New("could not copy buffer")
)

type Resource struct {
	URL          string
	Filename     string
	Size         int64
	AcceptRanges bool
}

type Network interface {
	FetchResource(url string) (Resource, error)
	DownloadResource(url string, start int64, end int64, writer io.Writer, ctx context.Context) error
}

type NetworkError string

func (e NetworkError) Error() string {
	return fmt.Sprintf("network error: %s", string(e))
}

type network struct{}

func (n network) FetchResource(URL string) (Resource, error) {
	// Resolve URL.
	URL, err := httputil.ResolveURL(URL)
	if err != nil {
		return Resource{}, err
	}

	// Extract the download filename from the headers or from the URL.
	path, err := url.PathUnescape(URL)
	if err != nil {
		return Resource{}, InvalidFilenameErr
	}

	// Download http request.
	response, err := http.Get(URL)
	if err != nil {
		return Resource{}, NetworkError(err.Error())
	}

	_ = response.Body.Close()

	// Extract Accept-Ranges from headers.
	acceptRanges := response.Header.Get("Accept-Ranges")

	filename := filepath.Base(path)
	_, contentDispositionParams, err := mime.ParseMediaType(response.Header.Get("Content-Disposition"))
	if name, ok := contentDispositionParams["filename"]; ok {
		filename = name
	}

	// Validate the download filename.
	if !fsutil.ValidateFilename(filename) {
		return Resource{}, InvalidFilenameErr
	}

	return Resource{
		URL:          URL,
		Filename:     filename,
		Size:         response.ContentLength,
		AcceptRanges: acceptRanges == "bytes",
	}, nil
}

func (n network) DownloadResource(url string, start int64, end int64, writer io.Writer, ctx context.Context) error {
	// Check if the segment has an overflow.
	if start < 0 || start > end {
		return SegmentOverflowErr
	}

	// Check if download already finished.
	if start == end {
		return nil
	}

	// Send HTTP GET request.
	request, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return NetworkError(err.Error())
	}

	// Start range download.
	request.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return NetworkError(err.Error())
	}

	defer response.Body.Close()

	_, err = io.Copy(writer, response.Body)
	if err != nil {
		return BufferCopyErr
	}

	return nil
}

func NewNetwork() Network {
	return &network{}
}

var _ Network = (*network)(nil)
