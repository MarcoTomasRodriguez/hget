package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type SegmentDownloader interface {
	Filename() string
	Download(writer io.Writer, offset int64, ctx context.Context)
}

type Segment struct {
	Id    uint8 `yaml:"id"`
	Start int64 `yaml:"start"`
	End   int64 `yaml:"end"`
}

func (s *Segment) Filename() string {
	return fmt.Sprintf("segment.%02d", s.Id)
}

func (s *Segment) Download(url string, offset int64, writer io.Writer, ctx context.Context) error {
	// Check if the segment is already downloaded.
	if offset == s.End {
		return nil
	}

	// Check if the segment has an overflow.
	if s.Start < offset && offset > s.End && s.End != -1 {
		return SegmentOverflowError{}
	}

	// Send http request.
	request, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return NetworkError(err.Error())
	}

	// Setup range download.
	request.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", offset, s.End))

	// Download get request with range header.
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return NetworkError(err.Error())
	}

	defer response.Body.Close()

	_, err = io.Copy(writer, response.Body)
	if err != nil {
		return IOCopyError(err.Error())
	}

	return nil
}
