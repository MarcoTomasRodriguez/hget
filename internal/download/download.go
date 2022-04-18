package download

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/cheggaaa/pb"
	"github.com/fatih/color"
	"github.com/pelletier/go-toml"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/MarcoTomasRodriguez/hget/internal/config"
	"github.com/MarcoTomasRodriguez/hget/pkg/fsutil"
	"github.com/MarcoTomasRodriguez/hget/pkg/logger"
)

var (
	// ErrDownloadNotExist is an error thrown when trying to fetch an inexistent download.
	ErrDownloadNotExist = errors.New("download does not exist")

	// ErrDownloadBroken is an error thrown when trying to fetch a broken download.
	ErrDownloadBroken = errors.New("download is broken")

	// ErrFilenameEmpty is an error thrown when a cleaned download filename is empty.
	ErrFilenameEmpty = errors.New("filename is empty")

	// ErrFilenameTooLong is an error thrown when a download filename is too long.
	ErrFilenameTooLong = errors.New("filename is too long")
)

// HTTPClient is a custom client with tls insecure skip verify enabled.
// TODO: Find a way to enable tls verify, and thus improve security, while allowing multi-threaded downloads.
var httpClient = &http.Client{
	Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
}

// ResolveURL resolves the rawURL adding the http prefix, preferring https over http.
func resolveURL(rawURL string) (string, error) {
	// Parse the raw rawURL.
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// Check if rawURL is empty.
	if parsedURL.String() == "" {
		return "", errors.New("rawURL is empty")
	}

	// Check if a scheme is provided.
	if parsedURL.Scheme == "https" || parsedURL.Scheme == "http" {
		return parsedURL.String(), nil
	}

	// Resolve using https.
	parsedURL.Scheme = "https"
	if res, err := http.Get(parsedURL.String()); err == nil && res != nil {
		return parsedURL.String(), nil
	}

	// Resolve using http.
	parsedURL.Scheme = "http"
	if res, err := http.Get(parsedURL.String()); err == nil && res != nil {
		return parsedURL.String(), nil
	}

	return "", errors.New("cannot resolve raw url using https or http")
}

// Download stores the information relative to a download, including the Workers.
type Download struct {
	// ID is the task unique identifier.
	// It is used to allow the download of many files with the same name from different sources.
	// This field will be initialized on runtime.
	ID string `toml:"-"`

	// URL represents the url from which the manager will download the file.
	URL string `toml:"url"`

	// Name is the original filename.
	Name string `toml:"name"`

	// Size is the total file size in bytes.
	Size uint64 `toml:"size"`

	// Resumable flags a download as Resumable.
	// If false, the download will be automatically removed on cancellation.
	// This field will be initialized on runtime.
	Resumable bool `toml:"-"`

	// Workers is the number of parallel connections configured for the manager.
	// Initially it is set by the user, but falls to 1 if the server does not accept range or does not provide a
	// content length.
	Workers []*Worker `toml:"Workers"`
}

// NewDownload fetches the download url, obtains all the information required to start a download and finally returns the download struct.
func NewDownload(downloadURL string, totalWorkers uint16) (*Download, error) {
	// Resolve download url.
	downloadURL, err := resolveURL(downloadURL)
	if err != nil {
		return nil, err
	}

	// Create request.
	httpResponse, err := httpClient.Get(downloadURL)
	if err != nil {
		return nil, err
	}

	defer httpResponse.Body.Close()

	// Get response headers.
	contentLengthHeader := httpResponse.Header.Get("Content-Length")
	acceptRangeHeader := httpResponse.Header.Get("Accept-Ranges")

	// If Content-size or Accept-Ranges headers are not provided by the download server,
	// set the number of Workers to 1.
	if contentLengthHeader == "" || acceptRangeHeader == "" {
		totalWorkers = 1
	}

	// If Content-Length is not provided by the download server, disable the Resumable feature.
	isResumable := contentLengthHeader != ""

	// Extract the download name from the url.
	downloadName := filepath.Base(downloadURL)

	// Check the extracted downloadName validity.
	if !fsutil.ValidateFilename(downloadName) {
		return nil, err
	}

	// Generate the internal downloadID.
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	downloadID := fmt.Sprintf("%x-%s", randomBytes, downloadName)

	// Read the download size from the header.
	downloadSize, _ := strconv.ParseUint(contentLengthHeader, 10, 64)

	// Create download Workers.
	workers := make([]*Worker, totalWorkers)
	for i := range workers {
		workers[i] = NewWorker(uint16(i), totalWorkers, downloadID, downloadURL, downloadSize)
	}

	return &Download{
		ID:        downloadID,
		URL:       downloadURL,
		Name:      downloadName,
		Size:      downloadSize,
		Resumable: isResumable,
		Workers:   workers,
	}, nil
}

// GetDownload gets a download by his id.
func GetDownload(downloadID string) (*Download, error) {
	d := &Download{ID: downloadID, Resumable: true}

	// Check if download folder exists.
	if _, err := os.Stat(d.FolderPath()); os.IsNotExist(err) {
		return nil, ErrDownloadNotExist
	}

	// Read download file.
	downloadFile, err := ioutil.ReadFile(d.FilePath())
	if err != nil {
		return nil, ErrDownloadBroken
	}

	// Unmarshal toml file into download struct.
	if err := toml.Unmarshal(downloadFile, d); err != nil {
		return nil, ErrDownloadBroken
	}

	return d, nil
}

// ListDownloads lists all the saved downloads.
func ListDownloads() ([]*Download, error) {
	var downloads []*Download

	// List elements inside the internal downloads' directory.
	downloadFolders, err := ioutil.ReadDir(config.Config.DownloadFolder())
	if err != nil {
		return nil, err
	}

	// Iterate over the elements inside the task folder, and append to the downloads array if valid.
	for _, downloadFolder := range downloadFolders {
		// A valid task should be located inside a directory.
		if downloadFolder.IsDir() {
			download, _ := GetDownload(downloadFolder.Name())

			if download != nil {
				downloads = append(downloads, download)
			}
		}
	}

	return downloads, nil
}

// Execute downloads the specified file.
// This operation blocks the execution until it finishes or is cancelled by the context.
func (d *Download) Execute(ctx context.Context) error {
	// Initialize uplink channels.
	errorChannel := make(chan error)
	doneChannel := make(chan struct{})
	workerProgressBars := make([]*pb.ProgressBar, len(d.Workers))

	// Create download folder.
	if err := os.MkdirAll(d.FolderPath(), 0755); err != nil {
		return err
	}

	go func(ctx context.Context) {
		waitGroup := new(sync.WaitGroup)

		// Start download Workers.
		for i, w := range d.Workers {
			// Calculate bytes left to download.
			currentSize := w.currentSize()
			downloadSize := w.downloadSize()
			bytesLeft := int64(0)
			if downloadSize > currentSize {
				bytesLeft = int64(downloadSize - currentSize)
			}

			// Setup progress bar.
			bar := pb.New64(bytesLeft).SetUnits(pb.U_BYTES).Prefix(
				color.CyanString(fmt.Sprintf("Worker #%d", w.Index)),
			)

			// Add worker to wait group.
			waitGroup.Add(1)

			go func(w *Worker, bar *pb.ProgressBar) {
				// Finish progress bar on exit.
				defer bar.Finish()

				// When the worker finished its execution remove it from wait group.
				defer waitGroup.Done()

				// Recover from panic.
				defer func() {
					if r := recover(); r != nil {
						errorChannel <- fmt.Errorf("Worker panic: %v", r)
					}
				}()

				if err := w.Execute(ctx, bar); err != nil {
					errorChannel <- err
				}
			}(w, bar)

			workerProgressBars[i] = bar
		}

		// Setup progress bar pool.
		barPool, _ := pb.StartPool(workerProgressBars...)

		// Wait until the last worker has finished his download.
		waitGroup.Wait()

		// Stop the progress bar pool.
		_ = barPool.Stop()

		doneChannel <- struct{}{}
	}(ctx)

	for {
		select {
		case <-doneChannel:
			if errors.Is(ctx.Err(), context.Canceled) {
				// Attempt to save the download.
				return d.attemptSave()
			}

			// Join Workers.
			if err := d.joinWorkers(); err != nil {
				return err
			}

			// Remove internal files.
			if err := d.Delete(); err != nil {
				return err
			}

			logger.Info("Download saved in %s.", d.OutputFilePath())
			return nil
		case err := <-errorChannel:
			// Attempt to save the download.
			if err := d.attemptSave(); err != nil {
				logger.Error("Could not save download: %v", err)
			}

			return err
		}
	}
}

// String returns a pretty string with the download's information.
func (d *Download) String() string {
	return fmt.Sprintln(
		" ⁕ ", color.HiCyanString(d.ID), " ⇒ ",
		color.HiCyanString("URL:"), d.URL,
		color.HiCyanString("Size:"), fsutil.ReadableMemorySize(d.Size),
	)
}

// writer opens the output file in write-only mode.
func (d *Download) writer() (io.WriteCloser, error) {
	return os.OpenFile(d.OutputFilePath(), os.O_CREATE|os.O_WRONLY, 0644)
}

// OutputFilePath returns the path of the download output.
func (d *Download) OutputFilePath() string {
	if config.Config.Download.CollisionProtection {
		return filepath.Join(config.Config.Download.Folder, d.ID)
	}

	return filepath.Join(config.Config.Download.Folder, d.Name)
}

// FolderPath gets the path to the download folder.
func (d *Download) FolderPath() string {
	return filepath.Join(config.Config.DownloadFolder(), d.ID)
}

// FilePath gets the path to the download TOML file.
func (d *Download) FilePath() string {
	return filepath.Join(d.FolderPath(), "download.toml")
}

// joinWorkers joins the worker files into the output file.
func (d *Download) joinWorkers() error {
	// Open output file.
	downloadWriter, err := d.writer()
	if err != nil {
		return err
	}

	defer downloadWriter.Close()

	// Setup progress bar.
	bar := pb.StartNew(len(d.Workers)).Prefix(color.CyanString("Joining"))

	// Stop progress bar on exit.
	defer bar.Finish()

	// Join the Workers' files into output file.
	for _, worker := range d.Workers {
		err := func() error {
			// Get worker reader.
			workerReader, err := worker.reader()
			if err != nil {
				return err
			}

			defer workerReader.Close()

			// Append worker file to output file.
			if _, err = io.Copy(downloadWriter, workerReader); err != nil {
				return err
			}

			bar.Increment()
			return nil
		}()

		if err != nil {
			return err
		}
	}

	return nil
}

// attemptSave saves the download struct as a toml file if it is Resumable, otherwise it deletes it.
func (d *Download) attemptSave() error {
	// If download is not Resumable, delete it.
	if !d.Resumable {
		return d.Delete()
	}

	// Parse download struct as toml.
	downloadToml, err := toml.Marshal(d)
	if err != nil {
		return err
	}

	// Save download as a toml file.
	if err := ioutil.WriteFile(d.FilePath(), downloadToml, 0644); err != nil {
		return err
	}

	logger.Info("Resumable download saved in %s.", d.FolderPath())
	return nil
}

// Delete removes all related files with the download.
func (d *Download) Delete() error {
	return os.RemoveAll(d.FolderPath())
}
