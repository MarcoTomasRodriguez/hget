package download

import (
	"context"
	"fmt"
	"github.com/cheggaaa/pb"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/MarcoTomasRodriguez/hget/internal/config"
)

// Worker represents a goroutine in charge of downloading a file part/segment.
type Worker struct {
	// Index is the index of the worker.
	// During the merge process, the worker downloads will be concatenated using this index.
	Index uint16 `toml:"index"`

	// DownloadID stores the id of the download.
	DownloadID string `toml:"download_id"`

	// DownloadURL stores the url of the download.
	DownloadURL string `toml:"download_url"`

	// RangeFrom is the start point of the worker download.
	RangeFrom uint64 `toml:"range_from"`

	// RangeTo is the end position of the worker download.
	RangeTo uint64 `toml:"range_to"`
}

// NewWorker computes the start & end point of the worker download and returns a new worker.
func NewWorker(workerIndex uint16, totalWorkers uint16, downloadId string, downloadURL string, downloadSize uint64) *Worker {
	// Calculate beginning of range.
	rangeFrom := (downloadSize / uint64(totalWorkers)) * uint64(workerIndex)

	// By default, a thread is in charge of downloading the whole file.
	rangeTo := downloadSize

	// If the worker index is not the last, calculate the desired range.
	if workerIndex < totalWorkers-1 {
		rangeTo = (downloadSize/uint64(totalWorkers))*(uint64(workerIndex)+1) - 1
	}

	return &Worker{
		Index:       workerIndex,
		DownloadID:  downloadId,
		RangeFrom:   rangeFrom,
		RangeTo:     rangeTo,
		DownloadURL: downloadURL,
	}
}

// Execute starts the download of the file slice.
// This operation is blocking and must be called inside a goroutine.
func (w *Worker) Execute(ctx context.Context, bar *pb.ProgressBar) error {
	// Compute current range from (defined start + worker file size).
	currentRangeFrom := w.RangeFrom + w.currentSize()

	// Create worker file.
	workerWriter, err := w.writer()
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
	httpResponse, err := httpClient.Do(httpRequest)
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

// filePath returns the worker file path.
func (w *Worker) filePath() string {
	return filepath.Join(config.Config.DownloadFolder(), w.DownloadID, fmt.Sprintf("worker.%05d", w.Index))
}

// reader opens the worker file in read-only mode.
func (w *Worker) reader() (io.ReadCloser, error) {
	return os.OpenFile(w.filePath(), os.O_RDONLY, 0644)
}

// writer opens the worker file in append-write-only mode.
func (w *Worker) writer() (io.WriteCloser, error) {
	return os.OpenFile(w.filePath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
}

// downloadSize calculates the difference between the maximum and minimum range.
func (w *Worker) downloadSize() uint64 {
	return w.RangeTo - w.RangeFrom
}

// currentSize returns the size of the worker file.
func (w *Worker) currentSize() uint64 {
	fileInfo, err := os.Stat(w.filePath())
	if err != nil {
		return 0
	}

	return uint64(fileInfo.Size())
}
