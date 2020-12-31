package download

import (
	"crypto/tls"
	"fmt"
	"github.com/MarcoTomasRodriguez/hget/config"
	"github.com/MarcoTomasRodriguez/hget/logger"
	"github.com/MarcoTomasRodriguez/hget/utils"
	"github.com/fatih/color"
	"gopkg.in/cheggaaa/pb.v1"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"
)

var (
	client = &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	resumable = true
)

const (
	acceptRangeHeader   = "Accept-Ranges"
	contentLengthHeader = "Content-Length"
)

// HTTPDownloader represents a download with his information.
type HTTPDownloader struct {
	URL         string
	FileName    string
	FileLength  int64
	Parts       []Part
	Parallelism int64
	Resumable   bool
}

// Download downloads the file from the url considering the state of the task using parallelism.
func Download(url string, task *Task, parallelism int) {
	// Set up parallelism
	var parts = make([]Part, 0)
	var isInterrupted = false

	// signalChan listens to a syscall to interrupt the download
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// doneChan represents the end of the main download thread
	doneChan := make(chan bool, parallelism)

	// errorChan represents an unrecoverable error
	errorChan := make(chan error, 1)

	// taskChan represents the last state of an interrupted task/part
	taskChan := make(chan Part, 1)

	// interruptChan signalizes the interruption of a task
	interruptChan := make(chan bool, parallelism)

	// writtenBytesChan represents the written bytes of every task
	writtenBytesChan := make(chan int64, parallelism)

	// Start or resume download
	var downloader *HTTPDownloader
	if task == nil {
		downloader = NewHTTPDownloader(url, int64(parallelism))
	} else {
		downloader = &HTTPDownloader{
			URL:         task.URL,
			FileName:    filepath.Base(task.URL),
			Parallelism: int64(len(task.Parts)),
			Parts:       task.Parts,
			Resumable:   true,
		}
	}

	logger.Info("Starting download with %d connection(s).\n", downloader.Parallelism)

	downloadStart := time.Now()
	writtenBytes := int64(0)

	go downloader.Do(doneChan, errorChan, interruptChan, taskChan, writtenBytesChan)

	for {
		select {
		case err := <-errorChan:
			logger.Panic(err)
		case part := <-taskChan:
			parts = append(parts, part)
		case wb := <-writtenBytesChan:
			writtenBytes += wb
		case <-signalChan:
			isInterrupted = true
			for i := 0; i < parallelism; i++ {
				interruptChan <- true
			}
		case <-doneChan:
			if isInterrupted {
				if downloader.Resumable {
					logger.Info("Interrupted. Saving task... \n")
					s := &Task{URL: url, Parts: parts}
					err := s.SaveTask()
					if err != nil {
						logger.Info("%v\n", err)
					}
					return
				}
				logger.Warn("Interrupted. Task was not saved because it is not resumable.\n")
				return
			}

			outputName := ""
			downloadTime := time.Since(downloadStart).Round(time.Millisecond)
			downloadSize := utils.ReadableMemorySize(writtenBytes)
			downloadSpeed := utils.ReadableMemorySize(int64(float64(writtenBytes)/downloadTime.Seconds())) + "/s"
			logger.Info("Downloaded %s in %s at an average speed of %s.\n", downloadSize, downloadTime, downloadSpeed)

			// Save with/without hash depending on configuration
			if config.Config.SaveWithHash {
				outputName = utils.FilenameWithHash(url)
			} else {
				outputName = utils.FilenameWithoutHash(url)
			}

			logger.Info("Joining process initiated.\n")

			// Get output path
			outputPath, err := filepath.Abs(filepath.Join(config.Config.DownloadFolder, outputName))
			utils.FatalCheck(err)

			// Get the full path to the download parts
			var files = make([]string, 0, len(parts))
			for _, part := range parts {
				files = append(files, part.Path)
			}

			// Join the parts and save the complete file in the output path
			err = JoinParts(files, outputPath)
			utils.FatalCheck(err)
			logger.Info("Joining process finished.\n")

			// Remove download parts
			logger.Info("Removing parts.\n")
			err = os.RemoveAll(utils.FolderOf(url))
			utils.FatalCheck(err)

			logger.Info("File saved in %s.\n", outputPath)

			return
		}
	}
}

// NewHTTPDownloader initializes the download and returns a downloader.
func NewHTTPDownloader(url string, parallelism int64) *HTTPDownloader {
	// Create request
	req, err := http.NewRequest("GET", url, nil)
	utils.FatalCheck(err)

	// Log download server host
	logger.Info("Download server: %s.\n", req.Host)

	// Execute request
	resp, err := client.Do(req)
	utils.FatalCheck(err)

	// Check for Content-Length and Accept-Ranges headers
	contentLength := resp.Header.Get(contentLengthHeader)
	acceptRange := resp.Header.Get(acceptRangeHeader)
	if contentLength == "" {
		logger.Info("Target url doesn't contain Content-Length header. Thus, the following features were disabled: progress bar, resumable downloads and parallelism.\n")
		contentLength = "0"
		parallelism = 1
		resumable = false
		config.Config.DisplayProgressBar = false
	} else if acceptRange == "" {
		logger.Info("Target url doesn't contain Accept-Range header. Thus, the parallelism feature was disabled.\n")
		parallelism = 1
	}

	// Get file length
	fileLength, err := strconv.ParseInt(contentLength, 10, 64)
	utils.FatalCheck(err)

	// Display download target size
	if fileLength > 0 {
		logger.Info("Download size: %s.\n", utils.ReadableMemorySize(fileLength))
	} else {
		logger.Info("Download size: not specified.\n")
	}

	// Return HTTPDownloader struct
	httpDownloader := &HTTPDownloader{
		URL:         url,
		FileName:    filepath.Base(url),
		FileLength:  fileLength,
		Parts:       CalculateParts(url, parallelism, fileLength),
		Parallelism: parallelism,
		Resumable:   resumable,
	}

	return httpDownloader
}

// Do downloads from the downloader.
func (downloader *HTTPDownloader) Do(doneChan chan bool, errorChan chan error, interruptChan chan bool, taskChan chan Part, writtenBytesChan chan int64) {
	var waitGroup sync.WaitGroup
	var barPool *pb.Pool
	var err error

	bars := make([]*pb.ProgressBar, 0, len(downloader.Parts))

	for partIndex, part := range downloader.Parts {
		var bar *pb.ProgressBar
		var partIndex = partIndex

		// Setup progress bar
		if config.Config.DisplayProgressBar {
			bar = pb.New64(part.RangeTo - part.RangeFrom).SetUnits(pb.U_BYTES).Prefix(
				color.YellowString(fmt.Sprintf("%s-%d", downloader.FileName, partIndex)),
			)
			bars = append(bars, bar)
		}

		// Add downloader part to wait group
		waitGroup.Add(1)

		go func(downloader *HTTPDownloader, part Part, bar *pb.ProgressBar) {
			// On function end, remove downloader part from wait group
			defer waitGroup.Done()

			// Send request
			req, err := http.NewRequest("GET", downloader.URL, nil)
			if err != nil {
				errorChan <- err
				return
			}

			// Setup range download if parallelism is greater than 1
			if downloader.Parallelism > 1 {
				var ranges string

				if part.RangeTo != downloader.FileLength {
					ranges = fmt.Sprintf("bytes=%d-%d", part.RangeFrom, part.RangeTo)
				} else {
					ranges = fmt.Sprintf("bytes=%d-", part.RangeFrom) // get all
				}

				req.Header.Add("Range", ranges)
			}

			// Write response body to file
			resp, err := client.Do(req)
			defer resp.Body.Close()
			if err != nil {
				errorChan <- err
				return
			}

			// Create part file
			file, err := os.OpenFile(part.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
			defer file.Close()
			if err != nil {
				errorChan <- err
				return
			}

			// Create writer with progress bar if enabled
			var writer io.Writer
			if config.Config.DisplayProgressBar {
				writer = io.MultiWriter(file, bar)
			} else {
				writer = io.MultiWriter(file)
			}

			current := int64(0)
			for {
				select {
				case <-interruptChan:
					// Save task and stop download
					taskChan <- Part{
						Index:     partIndex,
						Path:      part.Path,
						RangeFrom: current + part.RangeFrom,
						RangeTo:   part.RangeTo,
					}
					return
				default:
					// Copy from response to writer
					written, err := io.CopyN(writer, resp.Body, config.Config.CopyNBytes)
					writtenBytesChan <- written
					current += written

					if err != nil {
						// Throw error if any (in this case, EOF is not considered an error)
						if err != io.EOF {
							errorChan <- err
						}

						// Finish progress bar
						if config.Config.DisplayProgressBar {
							bar.Finish()
						}

						// Save task and stop download
						taskChan <- Part{
							Index:     partIndex,
							Path:      part.Path,
							RangeFrom: part.RangeFrom,
							RangeTo:   part.RangeTo,
						}
						return
					}
				}
			}
		}(downloader, part, bar)
	}

	// Setup progress bar pool
	if config.Config.DisplayProgressBar {
		barPool, err = pb.StartPool(bars...)
		if err != nil {
			errorChan <- err
			return
		}
	}

	// Wait until the last part finished his download
	waitGroup.Wait()

	// Stop progress bar pool
	if config.Config.DisplayProgressBar {
		if err = barPool.Stop(); err != nil {
			errorChan <- err
			return
		}
	}

	doneChan <- true
}
