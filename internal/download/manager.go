package download

import (
	"context"
	"fmt"
	"github.com/MarcoTomasRodriguez/hget/pkg/httputil"
	"github.com/cheggaaa/pb"
	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
	"github.com/samber/lo"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
	"io"
	ioFs "io/fs"
	"os"
	"path/filepath"
	"sync"
)

type Manager struct {
	// afs is the filesystem used by the Manager.
	// Remember to configure a root path, as it works on the root directory of the filesystem, which in the real
	// implementation should never be the OS root path.
	// Note: afs is a wrapper on top of afero.Fs that provides useful utilities.
	afs afero.Afero
}

// GetDownloadById loads the download object from disk.
func (m *Manager) GetDownloadById(id string) (*Download, error) {
	// Initialize file struct.
	download := &Download{}

	// Check if download file exists.
	if exists, _ := m.afs.DirExists("downloads/" + id); !exists {
		return nil, NonexistentDownloadError{}
	}

	// Read download information.
	fileBytes, err := m.afs.ReadFile("downloads/" + id + "/download.yml")
	if err != nil {
		return nil, BrokenDownloadError{}
	}

	// Unmarshal toml download into the file struct.
	if err := yaml.Unmarshal(fileBytes, &download); err != nil {
		return nil, BrokenDownloadError{}
	}

	return download, nil
}

// GetDownloadByUrl loads the download object from disk.
func (m *Manager) GetDownloadByUrl(url string) (*Download, error) {
	download := &Download{}

	url, err := httputil.ResolveURL(url)
	if err != nil {
		return nil, err
	}

	downloads, err := m.ListDownloads()
	if err != nil {
		return nil, err
	}

	download, _ = lo.Find(downloads, func(f *Download) bool { return f.URL == url })

	return download, nil
}

// ListDownloads loads all download objects from disk.
func (m *Manager) ListDownloads() ([]*Download, error) {
	// List elements inside the internal download directory.
	downloadFolders, err := m.afs.ReadDir("downloads")
	if err != nil {
		return nil, err
	}

	// Iterate over the elements inside the download folder and read them.
	downloads := lo.FilterMap(downloadFolders, func(fi ioFs.FileInfo, _ int) (*Download, bool) {
		d, err := m.GetDownloadById(fi.Name())
		if err != nil {
			return d, false
		}

		return d, true
	})

	return downloads, nil
}

// DeleteDownloadById deletes a download object from disk and removes all the child files.
func (m *Manager) DeleteDownloadById(id string) error {
	return m.afs.RemoveAll(filepath.Join("downloads", id))
}

// initDownloadFilesystem creates the download folder and initializes a restrictive base path filesystem.
func (m *Manager) initDownloadFilesystem(downloadId string) (afero.Afero, error) {
	// Create download folder.
	downloadFolder := filepath.Join("downloads", downloadId)
	if err := m.afs.MkdirAll(downloadFolder, os.ModePerm); err != nil {
		return afero.Afero{}, FilesystemError(err.Error())
	}

	return afero.Afero{Fs: afero.NewBasePathFs(m.afs, downloadFolder)}, nil
}

// saveDownloadToFilesystem writes the download object into disk.
func (m *Manager) saveDownloadToFilesystem(afs afero.Afero, download *Download) error {
	// Create and open download.yml file.
	downloadYml, err := afs.OpenFile("download.yml", os.O_CREATE|os.O_WRONLY, 0644)
	defer func() { _ = downloadYml.Close() }()
	if err != nil {
		return err
	}

	encoder := yaml.NewEncoder(downloadYml)
	defer func() { _ = encoder.Close() }()

	return encoder.Encode(download)
}

// StartDownload downloads all file segments and joins them in the output folder.
func (m *Manager) StartDownload(file *Download, ctx context.Context) error {
	var wg sync.WaitGroup
	var progressBars []*pb.ProgressBar

	// Create download folder and initialize a restrictive filesystem.
	afs, err := m.initDownloadFilesystem(file.Id)
	if err != nil {
		return err
	}

	// Save download object to disk.
	if file.Size > 0 {
		if err := m.saveDownloadToFilesystem(afs, file); err != nil {
			return err
		}
	}

	// Create a channel to listen to the workers' return error.
	workerErrors := make(chan error)

	// Progress bar utilities.
	showProgressBar := file.Size > 0 && isatty.IsTerminal(os.Stdout.Fd())

	for i, segment := range file.Segments {
		var progressBar *pb.ProgressBar

		// Create the segment file with permissions: -rw-r--r--.
		segmentFile, err := afs.OpenFile(segment.Filename(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return FilesystemError(err.Error())
		}

		segmentStat, err := segmentFile.Stat()
		segmentOffset := segment.Start + segmentStat.Size()
		if segmentOffset >= segment.End {
			continue
		}

		// Add progress bar to pool.
		if showProgressBar {
			progressBar = pb.New64(segment.End - segmentOffset).SetUnits(pb.U_BYTES).Prefix(
				color.CyanString(fmt.Sprintf("Worker #%d", i)),
			)
			progressBars = append(progressBars, progressBar)
		}

		// Worker thread.
		wg.Add(1)
		go func(segment Segment, segmentFile afero.File, progressBar *pb.ProgressBar, segmentOffset int64) {
			var segmentWriter io.Writer = segmentFile

			defer wg.Done()
			defer segmentFile.Close()

			if progressBar != nil {
				defer progressBar.Finish()
				segmentWriter = io.MultiWriter(segmentWriter, progressBar)
			}

			if err := segment.Download(file.URL, segmentOffset, segmentWriter, ctx); err != nil {
				workerErrors <- err
			}
		}(segment, segmentFile, progressBar, segmentOffset)
	}

	// Start progress bar.
	if showProgressBar {
		pool, _ := pb.StartPool(progressBars...)
		defer func() { _ = pool.Stop() }()
	}

	waitGroupDone := make(chan struct{})
	go func() {
		defer close(waitGroupDone)
		wg.Wait()
	}()

	contextDone := ctx.Done()

readChannels:
	for {
		select {
		case err := <-workerErrors:
			return err
		case <-contextDone:
			return CancelledDownloadError("user cancelled")
		case <-waitGroupDone:
			break readChannels
		}
	}

	// Open output file in write-only mode with permissions: -rw-r--r--.
	outputFile, err := afs.OpenFile(file.Name, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return FilesystemError(err.Error())
	}

	defer outputFile.Close()

	// Setup progress progressBar.
	var progressBar *pb.ProgressBar
	if showProgressBar {
		progressBar = pb.StartNew(len(file.Segments)).Prefix(color.CyanString("Merging"))
		defer progressBar.Finish()
	}

	// Join the segments into the output file.
	for _, s := range file.Segments {
		// Open segment file.
		segmentFile, err := afs.Open(s.Filename())
		if err != nil {
			return FilesystemError(err.Error())
		}

		// Append worker file to output file.
		if _, err = io.Copy(outputFile, segmentFile); err != nil {
			return IOCopyError(err.Error())
		}

		// Remove segment file.
		if err := afs.Remove(s.Filename()); err != nil {
			return FilesystemError(err.Error())
		}

		if showProgressBar {
			progressBar.Increment()
		}
	}

	return nil
}

// NewManager initializes the manager object.
func NewManager(fs afero.Fs) *Manager {
	return &Manager{afero.Afero{Fs: fs}}
}
