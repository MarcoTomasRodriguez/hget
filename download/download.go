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
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: false}},
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
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Set up parallelism
	var files = make([]string, 0)
	var parts = make([]Part, 0)
	var isInterrupted = false

	doneChan := make(chan bool, parallelism)
	fileChan := make(chan string, parallelism)
	errorChan := make(chan error, 1)
	taskChan := make(chan Part, 1)
	interruptChan := make(chan bool, parallelism)
	writtenBytesChan := make(chan int64, parallelism)

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

	go downloader.Do(doneChan, fileChan, errorChan, interruptChan, taskChan, writtenBytesChan)

	for {
		select {
		case <-signalChan:
			isInterrupted = true
			for i := 0; i < parallelism; i++ {
				interruptChan <- true
			}
		case file := <-fileChan:
			files = append(files, file)
		case err := <-errorChan:
			logger.Panic(err)
		case part := <-taskChan:
			parts = append(parts, part)
		case wb := <-writtenBytesChan:
			writtenBytes += wb
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
			downloadTime := time.Since(downloadStart)
			downloadSize := utils.ReadableMemorySize(writtenBytes)
			downloadSpeed := utils.ReadableMemorySize(int64(float64(writtenBytes)/downloadTime.Seconds())) + "/s"
			logger.Info("Downloaded %s in %s at an average speed of %s.\n", downloadSize, downloadTime, downloadSpeed)

			// Save with/without hash depending on configuration
			if config.Config.SaveWithHash {
				outputName = utils.FilenameWithHash(url)
			} else {
				outputName = utils.FilenameWithoutHash(url)
			}

			// Join the parts and save the complete file in the output path
			logger.Info("Joining process initiated.\n")
			outputPath, err := filepath.Abs(filepath.Join(config.Config.DownloadFolder, outputName))
			utils.FatalCheck(err)

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

	// Check support for range download, if not, change parallelism to 1
	if resp.Header.Get(acceptRangeHeader) == "" {
		logger.Info("Target url doesn't support range download. Changing parallelism to 1.\n")
		parallelism = 1
	}

	// Get download range
	contentLength := resp.Header.Get(contentLengthHeader)
	if contentLength == "" {
		logger.Info("Target url doesn't contain Content-Length header. Changing parallelism to 1.\n")
		contentLength = "0"
		parallelism = 1
		resumable = false
		config.Config.DisplayProgressBar = false
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
func (d *HTTPDownloader) Do(doneChan chan bool, fileChan chan string, errorChan chan error, interruptChan chan bool,
	taskSaveChan chan Part, writtenBytesChan chan int64) {
	var ws sync.WaitGroup
	var barPool *pb.Pool
	var err error

	bars := make([]*pb.ProgressBar, 0)

	for i, p := range d.Parts {
		var bar *pb.ProgressBar

		if config.Config.DisplayProgressBar {
			bar = pb.New64(p.RangeTo - p.RangeFrom).SetUnits(pb.U_BYTES).Prefix(color.YellowString(
				fmt.Sprintf("%s-%d", d.FileName, i)))
			bars = append(bars, bar)
		}

		ws.Add(1)
		go func(d *HTTPDownloader, part Part, bar *pb.ProgressBar) {
			defer ws.Done()
			var ranges string

			if part.RangeTo != d.FileLength {
				ranges = fmt.Sprintf("bytes=%d-%d", part.RangeFrom, part.RangeTo)
			} else {
				ranges = fmt.Sprintf("bytes=%d-", part.RangeFrom) // get all
			}

			// Send request
			req, err := http.NewRequest("GET", d.URL, nil)
			if err != nil {
				errorChan <- err
				return
			}

			if d.Parallelism > 1 {
				req.Header.Add("Range", ranges)
			}

			// Write response body to file
			resp, err := client.Do(req)
			defer resp.Body.Close()
			if err != nil {
				errorChan <- err
				return
			}

			file, err := os.OpenFile(part.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
			defer file.Close()
			if err != nil {
				errorChan <- err
				return
			}

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
					taskSaveChan <- Part{
						Path:      part.Path,
						RangeFrom: current + part.RangeFrom,
						RangeTo:   part.RangeTo,
					}
					return
				default:
					written, err := io.CopyN(writer, resp.Body, config.Config.CopyNBytes)
					writtenBytesChan <- written
					current += written
					if err != nil {
						if err != io.EOF {
							errorChan <- err
						}
						if config.Config.DisplayProgressBar {
							bar.Finish()
						}
						fileChan <- part.Path
						return
					}
				}
			}
		}(d, p, bar)
	}

	if config.Config.DisplayProgressBar {
		barPool, err = pb.StartPool(bars...)
		if err != nil {
			errorChan <- err
			return
		}
	}

	ws.Wait()

	if config.Config.DisplayProgressBar {
		if err = barPool.Stop(); err != nil {
			errorChan <- err
			return
		}
	}

	doneChan <- true
}
