package download

import (
	"context"
	"errors"
	"fmt"
	"github.com/cheggaaa/pb"
	"github.com/samber/do"
	"github.com/spf13/afero"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/MarcoTomasRodriguez/hget/internal/config"
)

// A Worker is the entity in charge of downloading a specific file segment.
type Worker struct {
	// Index is used to identify and sequentially order the workers during the final concatenation process.
	Index int `toml:"index"`

	// StartingPoint specifies the segment starting byte.
	StartingPoint int64 `toml:"starting_point"`

	// EndPoint specifies the segment end byte.
	EndPoint int64 `toml:"end_point"`

	// ...
	download *Download
}

// NewWorker computes the byte endpoints and returns a new worker.
func NewWorker(download *Download, workerIndex int) *Worker {
	// ...
	workerCount := len(download.Workers)

	// Compute the worker's starting point.
	startingPoint := (download.Size / int64(workerCount)) * int64(workerIndex)

	// Initialize the worker's end point. By default, it is the download size.
	endPoint := download.Size

	// If the worker is not the last, compute his end point.
	if workerIndex < workerCount-1 {
		endPoint = (download.Size/int64(workerCount))*(int64(workerIndex)+1) - 1
	}

	return &Worker{
		Index:         workerIndex,
		StartingPoint: startingPoint,
		EndPoint:      endPoint,
		download:      download,
	}
}

// Execute starts the worker's segment download blocking the execution, hence it must be called inside a goroutine.
func (w *Worker) Execute(ctx context.Context, bar *pb.ProgressBar) error {
	fs := do.MustInvoke[afero.Fs](do.DefaultInjector)
	cfg := do.MustInvoke[*config.Config](do.DefaultInjector)

	// Computes the actual starting point by taking into account the worker file size.
	startingPoint := w.StartingPoint + w.fileSize()

	// Check if file size exceeds range.
	if startingPoint > w.EndPoint {
		return nil
	}

	// Create the worker file with permissions: -rw-r--r--.
	workerFile, err := fs.OpenFile(w.filePath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	defer workerFile.Close()

	// Send http request.
	httpRequest, err := http.NewRequestWithContext(ctx, "GET", w.download.URL, nil)
	if err != nil {
		return err
	}

	// Setup range download.
	httpRequest.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", startingPoint, w.EndPoint))

	// Execute get request with range header.
	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		return err
	}
	defer httpResponse.Body.Close()

	if httpResponse.Header.Get("ETag") != w.download.ETag {
		return errors.New("ETag does not match") // Restart download if etag does not match.
	}

	writer := io.MultiWriter(workerFile, bar)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// Copy from response to writer.
			_, err := io.CopyN(writer, httpResponse.Body, cfg.Download.CopyNBytes)

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
	cfg := do.MustInvoke[*config.Config](do.DefaultInjector)
	return filepath.Join(cfg.DownloadFolder(), w.download.ID, fmt.Sprintf("worker.%05d", w.Index))
}

// fileSize returns the size of the worker file.
func (w *Worker) fileSize() int64 {
	fs := do.MustInvoke[afero.Fs](do.DefaultInjector)
	fileInfo, err := fs.Stat(w.filePath())
	if err != nil {
		return 0
	}

	return fileInfo.Size()
}
