package download

import (
	"context"
	"github.com/MarcoTomasRodriguez/hget/pkg/httputil"
	"github.com/samber/lo"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
	ioFs "io/fs"
	"os"
	"path/filepath"
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
		return nil, NonexistentDownloadErr
	}

	// Read download information.
	fileBytes, err := m.afs.ReadFile("downloads/" + id + "/download.yml")
	if err != nil {
		return nil, BrokenDownloadErr
	}

	// Unmarshal toml download into the file struct.
	if err := yaml.Unmarshal(fileBytes, &download); err != nil {
		return nil, BrokenDownloadErr
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
func (m *Manager) saveDownloadToFilesystem(download *Download, afs afero.Afero) error {
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
	// Create download folder and initialize a restrictive filesystem.
	afs, err := m.initDownloadFilesystem(file.Id)
	if err != nil {
		return err
	}

	// Save download object to disk.
	if file.Size > 0 {
		if err := m.saveDownloadToFilesystem(file, afs); err != nil {
			return err
		}
	}

	return file.Download(afs, ctx)
}

// NewManager initializes the manager object.
func NewManager(fs afero.Fs) *Manager {
	return &Manager{afero.Afero{Fs: fs}}
}
