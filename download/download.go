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
	"math"
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
	// In the future, make skipTLS optional to improve the user security.
	transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client    = &http.Client{Transport: transport}
	resumable = true
)

const (
	acceptRangeHeader   = "Accept-Ranges"
	contentLengthHeader = "Content-Length"
)

// HttpDownloader represents a download with his information.
type HttpDownloader struct {
	Url         string
	FileName    string
	FileLength  int64
	Parts       []Part
	Parallelism int64
	Resumable   bool
}

// Download downloads the file from the url considering the state of the task using parallelism.
func Download(url string, task *Task, parallelism int) {
	var err error

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// set up parallel
	var files = make([]string, 0)
	var parts = make([]Part, 0)
	var isInterrupted = false

	doneChan := make(chan bool, parallelism)
	fileChan := make(chan string, parallelism)
	errorChan := make(chan error, 1)
	taskChan := make(chan Part, 1)
	interruptChan := make(chan bool, parallelism)
	writtenBytesChan := make(chan int64, parallelism)

	var downloader *HttpDownloader
	if task == nil {
		downloader = NewHttpDownloader(url, int64(parallelism))
	} else {
		downloader = &HttpDownloader{
			Url:         task.Url,
			FileName:    filepath.Base(task.Url),
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
			// send par number of interrupt for each routine
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
					logger.Info("Interrupted, saving task... \n")
					s := &Task{Url: url, Parts: parts}
					err := s.SaveTask()
					if err != nil {
						logger.Info("%v\n", err)
					}
					return
				} else {
					logger.Warn("Interrupted, but downloading url is not resumable, silently die\n")
					return
				}
			} else {
				var outputName string

				downloadTime := time.Since(downloadStart)
				downloadSize := utils.ReadableMemorySize(writtenBytes)
				downloadSpeed := utils.ReadableMemorySize(writtenBytes/int64(math.Max(downloadTime.Seconds(), 1))) + "/s"
				logger.Info("Downloaded %s in %s at an average speed of %s.\n", downloadSize, downloadTime, downloadSpeed)

				if config.SaveWithHash {
					outputName = utils.FilenameWithHash(url)
				} else {
					outputName = utils.FilenameWithoutHash(url)
				}

				fmt.Println("Test")

				logger.Info("Joining process initiated.\n")

				err = JoinParts(files, outputName)
				utils.FatalCheck(err)

				logger.Info("Joining process finished.\n")
				logger.Info("Removing parts.\n")

				err = os.RemoveAll(utils.FolderOf(url))
				utils.FatalCheck(err)

				outputPath, err := filepath.Abs(outputName)
				utils.FatalCheck(err)
				logger.Info("File saved in %s.\n", outputPath)

				return
			}
		}
	}
}

// NewHttpDownloader initializes the download and returns a downloader.
func NewHttpDownloader(url string, parallelism int64) *HttpDownloader {
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
		config.DisplayProgressBar = false
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

	// Return HttpDownloader struct
	httpDownloader := &HttpDownloader{
		Url:         url,
		FileName:    filepath.Base(url),
		FileLength:  fileLength,
		Parts:       CalculateParts(url, parallelism, fileLength),
		Parallelism: parallelism,
		Resumable:   resumable,
	}

	return httpDownloader
}

// Do downloads from the downloader.
func (d *HttpDownloader) Do(doneChan chan bool, fileChan chan string, errorChan chan error, interruptChan chan bool,
	taskSaveChan chan Part, writtenBytesChan chan int64) {
	var ws sync.WaitGroup
	var barPool *pb.Pool
	var err error

	bars := make([]*pb.ProgressBar, 0)

	for i, p := range d.Parts {
		var bar *pb.ProgressBar

		if config.DisplayProgressBar {
			bar = pb.New64(p.RangeTo - p.RangeFrom).SetUnits(pb.U_BYTES).Prefix(color.YellowString(
				fmt.Sprintf("%s-%d", d.FileName, i)))
			bars = append(bars, bar)
		}

		ws.Add(1)
		go func(d *HttpDownloader, part Part, bar *pb.ProgressBar) {
			defer ws.Done()
			var ranges string

			if part.RangeTo != d.FileLength {
				ranges = fmt.Sprintf("bytes=%d-%d", part.RangeFrom, part.RangeTo)
			} else {
				ranges = fmt.Sprintf("bytes=%d-", part.RangeFrom) // get all
			}

			// send request
			req, err := http.NewRequest("GET", d.Url, nil)
			if err != nil {
				errorChan <- err
				return
			}

			if d.Parallelism > 1 {
				req.Header.Add("Range", ranges)
			}

			// write to file
			resp, err := client.Do(req)
			if err != nil {
				errorChan <- err
				return
			}
			defer resp.Body.Close()

			f, err := os.OpenFile(part.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)

			defer f.Close()
			if err != nil {
				errorChan <- err
				return
			}

			var writer io.Writer
			if config.DisplayProgressBar {
				writer = io.MultiWriter(f, bar)
			} else {
				writer = io.MultiWriter(f)
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
					written, err := io.CopyN(writer, resp.Body, config.CopyNBytes)
					writtenBytesChan <- written
					current += written
					if err != nil {
						if err != io.EOF {
							errorChan <- err
						}
						if config.DisplayProgressBar {
							bar.Finish()
						}
						fileChan <- part.Path
						return
					}
				}
			}
		}(d, p, bar)
	}

	if config.DisplayProgressBar {
		barPool, err = pb.StartPool(bars...)
		if err != nil {
			errorChan <- err
			return
		}
	}

	ws.Wait()

	if config.DisplayProgressBar {
		if err = barPool.Stop(); err != nil {
			errorChan <- err
			return
		}
	}

	doneChan <- true
}
