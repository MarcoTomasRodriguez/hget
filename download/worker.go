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

	// URL ...
	URL string `toml:"url"`

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
		Index:      workerIndex,
		DownloadID: downloadId,
		RangeFrom:  rangeFrom,
		RangeTo:    rangeTo,
		URL:        downloadURL,
	}
}

func (w Worker) Reader() (io.ReadCloser, error) {
	return os.OpenFile(w.filePath(), os.O_RDONLY, 0600)
}

func (w Worker) Writer() (io.WriteCloser, error) {
	return os.OpenFile(w.filePath(), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
}

// filePath ...
func (w Worker) filePath() string {
	return filepath.Join(config.Config.DownloadFolder(), w.DownloadID, fmt.Sprintf("worker.%05d", w.Index))
}

// size ...
func (w Worker) size() uint64 {
	fileInfo, err := os.Stat(w.filePath())
	if err != nil {
		return 0
	}

	return uint64(fileInfo.Size())
}

func (w Worker) execute(ctx context.Context, bar *pb.ProgressBar) error {
	// Finish progress bar on exit.
	defer bar.Finish()

	// Compute current range from (defined start + worker file size).
	currentRangeFrom := w.RangeFrom + w.size()

	// Create worker file.
	workerWriter, err := w.Writer()
	if err != nil {
		return err
	}

	// Close worker writer on exit.
	defer workerWriter.Close()

	// Send request.
	httpRequest, err := http.NewRequest("GET", w.URL, nil)
	if err != nil {
		return err
	}

	// Check if file size exceeds range.
	fmt.Println(currentRangeFrom, w.RangeTo)
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
			fmt.Println(w.Index)
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
