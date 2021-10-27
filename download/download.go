package download

import (
	"context"
	"fmt"
	"io"

	"github.com/MarcoTomasRodriguez/hget/config"
	"github.com/MarcoTomasRodriguez/hget/logger"

	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/MarcoTomasRodriguez/hget/utils"
	"github.com/cheggaaa/pb"
	"github.com/fatih/color"
	"github.com/pelletier/go-toml"
)

// download ...
type download struct {
	// ID is the task unique identifier.
	// It is used to allow the download of many files with the same name from different sources.
	// This field will be initialized on creation.
	ID string `toml:"-"`

	// URL represents the url from which the manager will download the file.
	URL string `toml:"url"`

	// Name is the original filename.
	Name string `toml:"name"`

	// Size is the total file size in bytes.
	Size uint64 `toml:"size"`

	// IsResumable ...
	// This field will be initialized on creation.
	IsResumable bool `toml:"-"`

	// Workers is the number of parallel connections configured for the manager.
	// Initially it is set by the user, but falls to 1 if the server does not accept range or does not provide a
	// content length.
	Workers []Worker `toml:"workers"`
}

func ParseDownloadID(downloadURL string) (string, error) {
	// Resolve download url.
	downloadURL, err := utils.ResolveURL(downloadURL)
	if err != nil {
		return "", err
	}

	return utils.HashFilename(downloadURL, filepath.Base(downloadURL)), nil
}

// NewDownload ...
func NewDownload(downloadURL string, totalWorkers uint16) (*download, error) {
	// Resolve download url.
	downloadURL, err := utils.ResolveURL(downloadURL)
	if err != nil {
		return nil, err
	}

	// Create request.
	httpResponse, err := utils.HTTPClient.Get(downloadURL)
	if err != nil {
		return nil, err
	}

	// Get response headers.
	contentLengthHeader := httpResponse.Header.Get("Content-Length")
	acceptRangeHeader := httpResponse.Header.Get("Accept-Ranges")

	// If Content-size or Accept-Ranges headers are not provided by the download server,
	// set the number of Workers to 1.
	if contentLengthHeader == "" || acceptRangeHeader == "" {
		totalWorkers = 1
	}

	// By default, downloads are resumable.
	isResumable := true

	// If Content-size is not provided by the download server, disable the resumable feature and the progress bar.
	if contentLengthHeader == "" {
		isResumable = false
		// config.ConfigOld.DisplayProgressBar = false
	}

	// Extract the download name from the url.
	downloadName := filepath.Base(downloadURL)

	// Check the extracted downloadName validity.
	if err := utils.CheckFilenameValidity(downloadName); err != nil {
		return nil, err
	}

	// Get the internal downloadID.
	// It is the concatenation of the first N bytes of the URL's hash and the download name.
	downloadID, err := ParseDownloadID(downloadURL)
	if err != nil {
		return nil, err
	}

	// Read the download size from the header.
	downloadSize, _ := strconv.ParseUint(contentLengthHeader, 10, 64)

	// Create download Workers.
	workers := make([]Worker, totalWorkers)
	for i := range workers {
		workers[i] = NewWorker(uint16(i), totalWorkers, downloadID, downloadURL, downloadSize)
	}

	return &download{
		ID:          downloadID,
		URL:         downloadURL,
		Name:        downloadName,
		Size:        downloadSize,
		IsResumable: isResumable,
		Workers:     workers,
	}, nil
}

// GetDownload gets a download by his id.
func GetDownload(downloadID string) (*download, error) {
	d := &download{ID: downloadID, IsResumable: true}

	// Check if download folder exists.
	if _, err := os.Stat(d.folderPath()); os.IsNotExist(err) {
		return nil, utils.ErrDownloadNotExist
	}

	// Read download file.
	downloadFile, err := ioutil.ReadFile(d.filePath())
	if err != nil {
		return nil, utils.ErrDownloadBroken
	}

	// Unmarshall toml file into download struct.
	if err := toml.Unmarshal(downloadFile, d); err != nil {
		return nil, utils.ErrDownloadBroken
	}

	return d, nil
}

// ListDownloads lists all the saved downloads.
func ListDownloads() ([]*download, error) {
	var downloads []*download

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

// DeleteDownload removes the download folder.
func DeleteDownload(downloadID string) error {
	return os.RemoveAll(filepath.Join(config.Config.DownloadFolder(), downloadID))
}

// String ...
func (d download) String() string {
	return fmt.Sprintln(" ⁕ ", color.HiCyanString(d.ID), " ⇒ ", color.HiCyanString("URL:"), d.URL, color.HiCyanString("Size:"), utils.ReadableMemorySize(d.Size))
}

// Writer ...
func (d download) Writer() (io.WriteCloser, error) {
	return os.OpenFile(d.outputFilePath(), os.O_CREATE|os.O_WRONLY, os.ModePerm)
}

// outputFilePath ...
func (d download) outputFilePath() string {
	if config.Config.Download.CollisionProtection {
		return filepath.Join(config.Config.Download.Folder, utils.HashFilename(d.URL, d.Name))
	}

	return filepath.Join(config.Config.Download.Folder, d.Name)
}

// folderPath gets the path to the download folder.
func (d download) folderPath() string {
	return filepath.Join(config.Config.DownloadFolder(), d.ID)
}

// filePath gets the path to the download TOML file.
func (d download) filePath() string {
	return filepath.Join(d.folderPath(), "download.toml")
}

// Execute ...
func (d download) Execute(ctx context.Context) error {
	// Initialize uplink channel.
	downloadChannel := make(chan interface{})
	workerProgressBars := make([]*pb.ProgressBar, len(d.Workers))

	if err := os.MkdirAll(d.folderPath(), os.ModePerm); err != nil {
		return err
	}

	go func(ctx context.Context) {
		waitGroup := new(sync.WaitGroup)

		// Start download Workers.
		for i, w := range d.Workers {
			// Setup progress bar.
			bar := pb.New64(int64(w.RangeTo - w.RangeFrom - w.size())).SetUnits(pb.U_BYTES).Prefix(
				color.CyanString(fmt.Sprintf("Worker #%d", w.Index)),
			)

			// Add worker to wait group.
			waitGroup.Add(1)

			go func() {
				// When the worker finished its execution remove it from wait group.
				defer waitGroup.Done()

				if err := w.execute(ctx, bar); err != nil {
					downloadChannel <- ErrorEvent{Payload: err}
				}
			}()

			workerProgressBars[i] = bar
		}

		// Setup progress bar pool.
		barPool, _ := pb.StartPool(workerProgressBars...)

		// Stop progress bar pool on exit.
		defer func() {
			_ = barPool.Stop()
		}()

		// Wait until the last worker has finished his download.
		waitGroup.Wait()

		downloadChannel <- DoneEvent{}
	}(ctx)

	for {
		select {
		case <-ctx.Done():
			// If download was interrupted, and it is NOT resumable, then delete it.
			if !d.IsResumable {
				return DeleteDownload(d.ID)
			}

			// If download was interrupted, and it is resumable, then save it.
			return d.save()
		case event := <-downloadChannel:
			switch ev := event.(type) {
			case ErrorEvent:
				return ev.Payload
			case DoneEvent:
				return d.finish()
			}
		}
	}
}

// joinWorkers joins the worker files into the output file.
func (d download) joinWorkers() error {
	// Open output file.
	downloadWriter, err := d.Writer()
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
			workerReader, err := worker.Reader()
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

// save ...
func (d download) save() error {
	// Create directories.
	if err := os.MkdirAll(d.folderPath(), 0644); err != nil {
		return err
	}

	// Parse download struct as toml.
	downloadToml, err := toml.Marshal(d)
	if err != nil {
		return err
	}

	// Save download as toml.
	if err := ioutil.WriteFile(d.filePath(), downloadToml, 0644); err != nil {
		return err
	}

	logger.LogInfo("Resumable download saved in %s.", d.folderPath())
	return nil
}

// finish ...
func (d download) finish() error {
	// Join Workers.
	if err := d.joinWorkers(); err != nil {
		return err
	}

	// Remove internal files.
	if err := os.RemoveAll(d.folderPath()); err != nil {
		return err
	}

	logger.LogInfo("Download saved in %s.", d.outputFilePath())
	return nil
}
