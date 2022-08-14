package download

import (
	"errors"
	"fmt"
	"github.com/MarcoTomasRodriguez/hget/internal/config"
	"github.com/MarcoTomasRodriguez/hget/pkg/fsutil"
	"github.com/jarcoal/httpmock"
	"github.com/pelletier/go-toml"
	"github.com/samber/do"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"path/filepath"
	"testing"
)

const (
	downloadID     = "fdc134c5f503b1bd-go1.17.2.src.tar.gz"
	downloadName   = "go1.17.2.src.tar.gz"
	downloadURL    = "https://golang.org/dl/go1.17.2.src.tar.gz"
	programFolder  = "/home/user/.hget"
	downloadFolder = "/home/user/downloads"
)

type DownloadSuite struct {
	suite.Suite
	afs *afero.Afero
}

func (s *DownloadSuite) SetupSuite() {
	httpmock.Activate()
}

func (s *DownloadSuite) SetupTest() {
	cfg := &config.Config{}
	cfg.ProgramFolder = programFolder
	cfg.Download.Folder = downloadFolder

	// Initialize Afero filesystem.
	s.afs = &afero.Afero{Fs: afero.NewMemMapFs()}

	do.ProvideValue[*config.Config](nil, cfg)
	do.ProvideValue[*afero.Afero](nil, s.afs)
}

func (s *DownloadSuite) TearDownTest() {
	httpmock.Reset()
}

func (s *DownloadSuite) TearDownSuite() {
	httpmock.DeactivateAndReset()
}

func (s *DownloadSuite) TestResolveUrl() {
	okResponder := httpmock.NewStringResponder(200, "OK")
	noResponder := httpmock.NewErrorResponder(errors.New("timeout"))

	httpmock.RegisterResponder("GET", "=~(https|http)://www.google.com", okResponder)
	httpmock.RegisterResponder("GET", "https://localhost/files/test.txt", noResponder)
	httpmock.RegisterResponder("GET", "http://localhost/files/test.txt", okResponder)
	httpmock.RegisterResponder("GET", "=~(https|http)://234.112.93.22:4123", noResponder)

	testCases := []struct {
		rawURL string
		url    string
		err    error
	}{
		{rawURL: "", url: "", err: errors.New("url is empty")},
		{rawURL: "www.google.com", url: "https://www.google.com", err: nil},
		{rawURL: "https://www.google.com", url: "https://www.google.com", err: nil},
		{rawURL: "http://www.google.com", url: "http://www.google.com", err: nil},
		{rawURL: "localhost/files/test.txt", url: "http://localhost/files/test.txt", err: nil},
		{rawURL: "https://localhost/files/test.txt", url: "", err: errors.New("server unavailable")},
		{rawURL: "http://localhost/files/test.txt", url: "http://localhost/files/test.txt", err: nil},
		{rawURL: "https://234.112.93.22:4123", url: "", err: errors.New("server unavailable")},
		{rawURL: "http://234.112.93.22:4123", url: "", err: errors.New("server unavailable")},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Raw URL: %s", tc.rawURL), func() {
			url, err := resolveURL(tc.rawURL)
			s.Equal(tc.err, err)
			s.Equal(tc.url, url)
		})
	}
}

func (s *DownloadSuite) TestGetDownload() {
	// Create download object and populate the directory with valid data.
	d := &Download{ID: downloadID, Resumable: true}
	dToml, _ := toml.Marshal(d)

	_ = s.afs.MkdirAll(d.FolderPath(), 0644)
	_ = s.afs.WriteFile(d.FilePath(), dToml, 0644)

	// Assert that the retrieved download is equal to the original one.
	download, err := GetDownload(downloadID)
	s.Equal(d, download)
	s.NoError(err)
}

func (s *DownloadSuite) TestGetDownload_NotExists() {
	// If the download folder was not created, it doesn't exist.
	d, err := GetDownload(downloadID)
	s.Nil(d)
	s.ErrorIs(ErrDownloadNotExist, err)
}

func (s *DownloadSuite) TestGetDownload_Broken_EmptyDir() {
	// If the download directory is empty, it is broken.
	d := &Download{ID: downloadID}
	_ = s.afs.MkdirAll(d.FolderPath(), 0644)

	// Assert that the download is broken.
	d, err := GetDownload(downloadID)
	s.Nil(d)
	s.ErrorIs(ErrDownloadBroken, err)
}

func (s *DownloadSuite) TestGetDownload_Broken_InvalidFile() {
	// If the download directory is empty, it is broken.
	d := &Download{ID: downloadID}
	_ = s.afs.MkdirAll(d.FolderPath(), 0644)
	_ = s.afs.WriteFile(d.FilePath(), []byte("invalid content"), 0644)

	// Assert that the download is broken.
	d, err := GetDownload(downloadID)
	s.Nil(d)
	s.ErrorIs(ErrDownloadBroken, err)
}

func (s *DownloadSuite) TestDownload_String() {
	d := &Download{ID: downloadID, URL: downloadURL, Size: 4.3243 * fsutil.GB}
	s.Equal(fmt.Sprintf(" ⁕ %s ⇒ URL: %s Size: 4.3 GB\n", downloadID, downloadURL), d.String())
}

func (s *DownloadSuite) TestDownload_Delete() {
	// Populate filesystem.
	d := &Download{ID: downloadID}
	folderPath := d.FolderPath()
	_ = s.afs.MkdirAll(folderPath, 0644)
	_ = s.afs.WriteFile(filepath.Join(folderPath, "download.toml"), []byte{}, 0644)

	// Delete download.
	s.NoError(d.Delete())

	// Check that the download folder does not exist.
	folderPathExists, err := s.afs.DirExists(folderPath)
	s.False(folderPathExists)
	s.NoError(err)

	// Check that the download object file does not exist.
	downloadFileExists, err := s.afs.Exists(filepath.Join(folderPath, "download.toml"))
	s.False(downloadFileExists)
	s.NoError(err)
}

func (s *DownloadSuite) TestDownload_OutputFilePath() {
	d := &Download{Name: downloadName}
	s.Equal(filepath.Join(downloadFolder, downloadName), d.OutputFilePath())
}

func (s *DownloadSuite) TestDownload_FolderPath() {
	d := &Download{ID: downloadID}
	s.Equal(filepath.Join(programFolder, "downloads", downloadID), d.FolderPath())
}

func (s *DownloadSuite) TestDownload_FilePath() {
	d := &Download{ID: downloadID}
	s.Equal(filepath.Join(programFolder, "downloads", downloadID, "download.toml"), d.FilePath())
}

func TestDownloadSuite(t *testing.T) {
	suite.Run(t, new(DownloadSuite))
}
