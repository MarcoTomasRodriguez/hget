package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/MarcoTomasRodriguez/hget/config"
	"github.com/MarcoTomasRodriguez/hget/utils"
	"github.com/cheggaaa/pb"
)

// Worker ...
type Worker struct {
	// Index ...
	Index uint16 `toml:"index"`

	// DownloadID ...
	DownloadID string `toml:"download_id"`

	// DownloadURL ...
	DownloadURL string `toml:"download_url"`

	// RangeFrom ...
	RangeFrom uint64 `toml:"range_from"`

	// RangeTo ...
	RangeTo uint64 `toml:"range_to"`
}

// NewWorker ...
func NewWorker(workerIndex uint16, totalWorkers uint16, downloadId string, downloadURL string, downloadSize uint64) Worker {
	// Calculate beginning of range.
	rangeFrom := (downloadSize / uint64(totalWorkers)) * uint64(workerIndex)

	// By default, a thread is in charge of downloading the whole file.
	rangeTo := downloadSize

	// If the worker index is not the last, calculate the desired range.
	if workerIndex < totalWorkers-1 {
		rangeTo = (downloadSize/uint64(totalWorkers))*(uint64(workerIndex)+1) - 1
	}

	return Worker{
		Index:       workerIndex,
		DownloadID:  downloadId,
		RangeFrom:   rangeFrom,
		RangeTo:     rangeTo,
		DownloadURL: downloadURL,
	}
}

// FilePath returns the worker file path.
func (w Worker) FilePath() string {
	return filepath.Join(config.Config.DownloadFolder(), w.DownloadID, fmt.Sprintf("worker.%05d", w.Index))
}

// Reader opens the worker file in read-only mode.
func (w Worker) Reader() (io.ReadCloser, error) {
	return os.OpenFile(w.FilePath(), os.O_RDONLY, 0644)
}

// Writer opens the worker file in append-write-only mode.
func (w Worker) Writer() (io.WriteCloser, error) {
	return os.OpenFile(w.FilePath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
}

// DownloadSize calculates the difference between the maximum and minimum range.
func (w Worker) DownloadSize() uint64 {
	return w.RangeTo - w.RangeFrom
}

// CurrentSize returns the size of the worker file.
func (w Worker) CurrentSize() uint64 {
	fileInfo, err := os.Stat(w.FilePath())
	if err != nil {
		return 0
	}

	return uint64(fileInfo.Size())
}

// Execute starts the download of the file slice.
// This operation is blocking and must be called inside a goroutine.
func (w Worker) Execute(ctx context.Context, bar *pb.ProgressBar) error {
	// Compute current range from (defined start + worker file size).
	currentRangeFrom := w.RangeFrom + w.CurrentSize()

	// Create worker file.
	workerWriter, err := w.Writer()
	if err != nil {
		return err
	}

	// Close worker writer on exit.
	defer workerWriter.Close()

	// Send request.
	httpRequest, err := http.NewRequest("GET", w.DownloadURL, nil)
	if err != nil {
		return err
	}

	// Check if file size exceeds range.
	if currentRangeFrom > w.RangeTo {
		return nil
	}

	// Setup range download.
	httpRequest.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", currentRangeFrom, w.RangeTo))

	// Execute get request with range header.
	httpResponse, err := utils.HTTPClient.Do(httpRequest)
	if err != nil {
		return err
	}

	defer httpResponse.Body.Close()

	writer := io.MultiWriter(workerWriter, bar)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// Copy from response to writer.
			_, err := io.CopyN(writer, httpResponse.Body, config.Config.Download.CopyNBytes)

			if err != nil {
				// Throw error if any (in this case, EOF is not considered an error).
				if err != io.EOF {
					return err
				}

				return nil
			}
		}
	}
}
