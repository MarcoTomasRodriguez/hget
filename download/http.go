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
	"net"
	"net/http"
	netUrl "net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var (
	tr = &http.Transport{ TLSClientConfig: &tls.Config{InsecureSkipVerify: true} }
	client = &http.Client{Transport: tr}
)

var (
	resumable 			= true
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
	Ips         []string
	SkipTLS     bool
	Resumable   bool
}

// partCalculate calculates the parts.
func partCalculate(url string, parallelism int64, length int64) []Part {
	ret := make([]Part, 0)
	for current := int64(0); current < parallelism; current++ {
		from := (length / parallelism) * current
		to := int64(0)

		if current < parallelism - 1 {
			to = (length/parallelism) * (current + 1) - int64(1)
		} else {
			to = length
		}

		file := filepath.Base(url)
		folder := utils.FolderOf(url)
		if err := utils.MkdirIfNotExist(folder); err != nil {
			logger.Info("%v", err)
			os.Exit(1)
		}

		filename := fmt.Sprintf("%s.part%d", file, current)
		path := filepath.Join(folder, filename) // ~/.hget/download-file-name/part-name
		ret = append(ret, Part{Url: url, Path: path, RangeFrom: from, RangeTo: to})
	}

	return ret
}

// NewHttpDownloader initializes the download and returns a downloader.
func NewHttpDownloader(url string, parallelism int64, skipTLS bool) *HttpDownloader {
	// Parse the raw-url into a URL structure
	parsedUrl, err := netUrl.Parse(url)
	utils.FatalCheck(err)

	// Lookup for the ips
	ips, err := net.LookupIP(parsedUrl.Host)
	utils.FatalCheck(err)

	// Convert ipv4 ips to string
	ipsStr := utils.StringifyIpsV4(ips)
	logger.Info("Resolve ip: %s\n", strings.Join(ipsStr, " | "))

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	utils.FatalCheck(err)

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
		Parts:       partCalculate(url, parallelism, fileLength),
		Parallelism: parallelism,
		Ips:         ipsStr,
		SkipTLS:     skipTLS,
		Resumable:   resumable,
	}

	return httpDownloader
}

// Do downloads from the downloader.
func (d *HttpDownloader) Do(doneChan chan bool, fileChan chan string, errorChan chan error, interruptChan chan bool,
	taskSaveChan chan Part, writtenBytesChan chan int64) {
	var ws sync.WaitGroup
	var bars []*pb.ProgressBar
	var barPool *pb.Pool
	var err error

	if config.DisplayProgressBar {
		bars = make([]*pb.ProgressBar, 0)
		for i, part := range d.Parts {
			partLength := part.RangeTo - part.RangeFrom
			bar := pb.New64(partLength).SetUnits(pb.U_BYTES).Prefix(color.YellowString(
				fmt.Sprintf("%s-%d", d.FileName, i)))
			bars = append(bars, bar)
		}
		barPool, err = pb.StartPool(bars...)
		utils.FatalCheck(err)
	}

	for i, p := range d.Parts {
		ws.Add(1)
		go func(d *HttpDownloader, loop int64, part Part) {
			defer ws.Done()
			var bar *pb.ProgressBar
			var ranges string

			if config.DisplayProgressBar { bar = bars[loop] }

			if part.RangeTo != d.FileLength {
				ranges = fmt.Sprintf("bytes=%d-%d", part.RangeFrom, part.RangeTo)
			} else {
				ranges = fmt.Sprintf("bytes=%d-", part.RangeFrom) // get all
			}

			// send request
			req, err := http.NewRequest("GET", d.Url, nil)
			if err != nil { errorChan <- err; return }

			if d.Parallelism > 1 { // support range download just in case parallel factor is over 1
				req.Header.Add("Range", ranges)
				if err != nil { errorChan <- err; return }
			}

			// write to file
			resp, err := client.Do(req)
			if err != nil { errorChan <- err; return }
			defer resp.Body.Close()

			f, err := os.OpenFile(part.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)

			defer f.Close()
			if err != nil {
				logger.Error("%v\n", err)
				errorChan <- err
				return
			}

			var writer io.Writer
			if config.DisplayProgressBar {
				writer = io.MultiWriter(f, bar)
			} else {
				writer = io.MultiWriter(f)
			}

			// make copy interruptible by copy 100 bytes each loop
			current := int64(0)
			for {
				select {
				case <- interruptChan:
					taskSaveChan <- Part{
						Url: d.Url,
						Path: part.Path,
						RangeFrom: current + part.RangeFrom,
						RangeTo: part.RangeTo,
					}
					return
				default:
					written, err := io.CopyN(writer, resp.Body, config.CopyNBytes)
					writtenBytesChan <- written
					current += written
					if err != nil {
						if err != io.EOF { errorChan <- err }
						if config.DisplayProgressBar { bar.Finish() }
						fileChan <- part.Path
						return
					}
				}
			}
		}(d, int64(i), p)
	}

	ws.Wait()

	if config.DisplayProgressBar {
		if err = barPool.Stop(); err != nil { errorChan <- err; return }
	}

	doneChan <- true
}

