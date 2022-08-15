package download

import (
	"context"
	"errors"
	"fmt"
	"github.com/cheggaaa/pb"
	"github.com/fatih/color"
	"github.com/pelletier/go-toml"
	"github.com/samber/do"
	"github.com/samber/lo"
	"github.com/spf13/afero"
	"io"
	"io/fs"
	"math/rand"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/MarcoTomasRodriguez/hget/internal/config"
	"github.com/MarcoTomasRodriguez/hget/pkg/fsutil"
	"github.com/MarcoTomasRodriguez/hget/pkg/logger"
)

var (
	// ErrDownloadNotExist is an error thrown when trying to fetch a non-existent download.
	ErrDownloadNotExist = errors.New("download does not exist")

	// ErrDownloadBroken is an error thrown when trying to fetch a broken download.
	ErrDownloadBroken = errors.New("download is broken")

	// ErrDownloadNotResumable is an error thrown when trying to save a non-resumable download.
	ErrDownloadNotResumable = errors.New("download is not resumable")

	urlRegex = regexp.MustCompile("(\\bhttps?://)?[-A-Za-z\\d+&@#/%?=~_|!:,.;]+[-A-Za-z\\d+&@#/%=~_|]")
)

// randBytes generates an array with a specific downloadSize containing random data.
func randBytes(size int) []byte {
	randomBytes := make([]byte, size)
	rand.Read(randomBytes)

	return randomBytes
}

// resolveURL resolves the rawURL adding the http prefix, preferring https over http.
func resolveURL(rawURL string) (string, error) {
	match := urlRegex.FindStringSubmatch(rawURL)

	// Check if rawURL is empty.
	if len(match) < 2 {
		return "", errors.New("url is empty")
	}

	// If scheme is provided, attempt to execute a request.
	scheme := match[1]
	if scheme != "" {
		res, err := http.Get(rawURL)
		if err != nil || res == nil {
			return "", errors.New("server unavailable")
		}

		return rawURL, nil
	}

	// Resolve using https.
	if url, err := resolveURL("https://" + rawURL); err == nil {
		return url, nil
	}

	// Resolve using http.
	if url, err := resolveURL("http://" + rawURL); err == nil {
		return url, nil
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

	// Size is the total file downloadSize in bytes.
	Size int64 `toml:"downloadSize"`

	// ETag stores the http response ETag.
	// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/ETag
	ETag string `toml:"etag"`

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
func NewDownload(downloadURL string, workerCount int) (*Download, error) {
	// Resolve download url.
	downloadURL, err := resolveURL(downloadURL)
	if err != nil {
		return nil, err
	}

	// Execute http request.
	httpResponse, err := http.Get(downloadURL)
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()

	// Extract Content-Length and Accept-Ranges from response headers.
	contentLength := httpResponse.ContentLength
	acceptRanges := httpResponse.Header.Get("Accept-Ranges")
	eTag := httpResponse.Header.Get("ETag")

	// In order for range downloads to work, they should be supported and the content length be provided.
	if contentLength == 0 || acceptRanges == "" || acceptRanges == "none" {
		logger.Warn("Range downloads are not supported by the server: setting worker count to 1")
		workerCount = 1
	}

	// Check if ETag was provided.
	if eTag == "" {
		logger.Warn("ETag was not provided by the server: resumable's download integrity cannot be guaranteed")
	}

	// Extract the download filename from the headers or from the url.
	downloadFilename := filepath.Base(downloadURL)
	_, contentDispositionParams, err := mime.ParseMediaType(httpResponse.Header.Get("Content-Disposition"))
	if filename, ok := contentDispositionParams["downloadFilename"]; ok {
		downloadFilename = filename
	}

	// Validate the download filename.
	if !fsutil.ValidateFilename(downloadFilename) {
		return nil, errors.New("invalid download filename")
	}

	// Generate the internal download id.
	downloadID := fmt.Sprintf("%x-%s", randBytes(8), downloadFilename)

	// Initialize download.
	download := &Download{
		ID:        downloadID,
		URL:       downloadURL,
		Name:      downloadFilename,
		ETag:      eTag,
		Size:      contentLength,
		Resumable: contentLength != 0, // Disable if Content-Length is not provided.
		Workers:   make([]*Worker, workerCount),
	}

	// Initialize download workers.
	for i := range download.Workers {
		download.Workers[i] = NewWorker(i, download)
	}

	return download, nil
}

// GetDownload gets a download by his id.
func GetDownload(downloadID string) (*Download, error) {
	afs := do.MustInvoke[*afero.Afero](nil)
	d := &Download{ID: downloadID, Resumable: true}

	// Check if download folder exists.
	if exists, _ := afs.DirExists(d.FolderPath()); !exists {
		return nil, ErrDownloadNotExist
	}

	// Read download file.
	downloadFile, err := afs.ReadFile(d.FilePath())
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
// TODO: Add tests.
func ListDownloads() ([]*Download, error) {
	afs := do.MustInvoke[*afero.Afero](nil)
	cfg := do.MustInvoke[*config.Config](nil)

	// List elements inside the internal downloads' directory.
	downloadFolders, err := afs.ReadDir(cfg.DownloadFolder())
	if err != nil {
		return nil, err
	}

	// Iterate over the elements inside the download folder and read them.
	downloads := lo.FilterMap(downloadFolders, func(fi fs.FileInfo, _ int) (*Download, bool) {
		download, err := GetDownload(fi.Name())
		if err != nil {
			return nil, false
		}

		return download, true
	})

	return downloads, nil
}

// Delete removes all related files with the download.
func (d *Download) Delete() error {
	afs := do.MustInvoke[*afero.Afero](nil)
	return afs.RemoveAll(d.FolderPath())
}

// String returns a pretty string with the download's information.
func (d *Download) String() string {
	return fmt.Sprintln(
		" ⁕", color.HiCyanString(d.ID), "⇒",
		color.HiCyanString("URL:"), d.URL,
		color.HiCyanString("Size:"), fsutil.ReadableMemorySize(d.Size),
	)
}

// OutputFilePath returns an available output path for the download file.
func (d *Download) OutputFilePath() string {
	afs := do.MustInvoke[*afero.Afero](nil)
	cfg := do.MustInvoke[*config.Config](nil)

	filename := filepath.Join(cfg.Download.Folder, d.Name)

	// If a download with the same name exists, add a number after the original name in parentheses.
	// Example: go1.17.2.src.tar.gz => go1.17.2.src.tar(1).gz
	if exists, _ := afs.Exists(filename); exists {
		// Split download file name by ".".
		parts := strings.Split(d.Name, ".")
		count := 1

		for {
			// Check if a filename with the current count already exists. If so, continue iterating until a filename
			// is available.
			filename = filepath.Join(cfg.Download.Folder, fmt.Sprintf("%s(%d).%s", strings.Join(parts[:len(parts)-1], "."), count, parts[len(parts)-1]))
			if exists, _ := afs.Exists(filename); !exists {
				break
			}

			count++
		}
	}

	return filename
}

// FolderPath gets the path to the download folder.
func (d *Download) FolderPath() string {
	cfg := do.MustInvoke[*config.Config](nil)
	return filepath.Join(cfg.DownloadFolder(), d.ID)
}

// FilePath gets the path to the download TOML file.
func (d *Download) FilePath() string {
	return filepath.Join(d.FolderPath(), "download.toml")
}

// Execute downloads the specified file.
// This operation blocks the execution until it finishes or is cancelled by the context.
func (d *Download) Execute(ctx context.Context) error {
	afs := do.MustInvoke[*afero.Afero](nil)

	// Initialize uplink channels.
	errorChannel := make(chan error)
	doneChannel := make(chan struct{})
	workerProgressBars := make([]*pb.ProgressBar, len(d.Workers))

	// Create download folder with the default unix directory permissions: drwxr-xr-x.
	if err := afs.MkdirAll(d.FolderPath(), 0755); err != nil {
		return err
	}

	go func(ctx context.Context) {
		var waitGroup sync.WaitGroup

		// Start download Workers.
		for i, w := range d.Workers {
			// Calculate bytes left to download.
			currentSize := w.fileSize()
			downloadSize := w.EndPoint - w.StartingPoint
			bytesLeft := int64(0)
			if downloadSize > currentSize {
				bytesLeft = downloadSize - currentSize
			}

			// Setup progress bar.
			bar := pb.New64(bytesLeft).SetUnits(pb.U_BYTES).Prefix(
				color.CyanString(fmt.Sprintf("Worker #%d", w.ID)),
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
						errorChannel <- fmt.Errorf("worker panic: %v", r)
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
				// If download is resumable, save it; otherwise, delete it.
				if !d.Resumable {
					return d.Delete()
				}

				return d.Save()
			}

			// Join Workers.
			if err := d.joinWorkers(); err != nil {
				return err
			}

			// Remove internal files.
			if err := d.Delete(); err != nil {
				return err
			}

			return nil
		case err := <-errorChannel:
			// If download is resumable, save it; otherwise, delete it.
			if !d.Resumable {
				return d.Delete()
			}

			if err := d.Save(); err != nil {
				logger.Error("Could not save download: %v", err)
			}

			return err
		}
	}
}

// joinWorkers joins the worker files into the output file.
func (d *Download) joinWorkers() error {
	afs := do.MustInvoke[*afero.Afero](nil)

	// Open output file in write-only mode with permissions: -rw-r--r--.
	downloadFilepath := d.OutputFilePath()
	downloadFile, err := afs.OpenFile(downloadFilepath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer downloadFile.Close()

	// Setup progress bar.
	bar := pb.StartNew(len(d.Workers)).Prefix(color.CyanString("Joining"))

	// Stop progress bar on exit.
	defer bar.Finish()

	// Join the Workers' files into output file.
	for _, w := range d.Workers {
		err := func() error {
			// Get worker reader.
			workerReader, err := afs.Open(w.filePath())
			if err != nil {
				return err
			}

			defer workerReader.Close()

			// Append worker file to output file.
			if _, err = io.Copy(downloadFile, workerReader); err != nil {
				return err
			}

			bar.Increment()
			return nil
		}()

		if err != nil {
			return err
		}
	}

	logger.Info("Download saved in %s.", downloadFilepath)

	return nil
}

// Save saves the download information inside a toml file if resumable, otherwise returns an error.
func (d *Download) Save() error {
	afs := do.MustInvoke[*afero.Afero](nil)

	// If it is non-resumable, return an error.
	if !d.Resumable {
		return ErrDownloadNotResumable
	}

	// Parse download struct as toml.
	downloadToml, err := toml.Marshal(d)
	if err != nil {
		return err
	}

	// Save download as a toml file with permissions: -rw-r--r--.
	if err := afs.WriteFile(d.FilePath(), downloadToml, 0644); err != nil {
		return err
	}

	logger.Info("Resumable download saved in %s.", d.FolderPath())
	return nil
}
