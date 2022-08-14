package download

import (
	"errors"
	"fmt"
	"github.com/MarcoTomasRodriguez/hget/internal/config"
	"github.com/MarcoTomasRodriguez/hget/pkg/fsutil"
	"github.com/jarcoal/httpmock"
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
}

func (s *DownloadSuite) SetupSuite() {
	httpmock.Activate()
}

func (s *DownloadSuite) SetupTest() {
	cfg := &config.Config{}
	cfg.ProgramFolder = programFolder
	cfg.Download.Folder = downloadFolder

	do.ProvideValue[*config.Config](nil, cfg)
	do.ProvideValue[*afero.Afero](nil, &afero.Afero{Fs: afero.NewMemMapFs()})
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
		{rawURL: "localhost/files/test.txt", url: "http://localhost/files/test.txt", err: nil},        // cannot resolve raw url
		{rawURL: "https://localhost/files/test.txt", url: "", err: errors.New("server unavailable")},  // empty
		{rawURL: "http://localhost/files/test.txt", url: "http://localhost/files/test.txt", err: nil}, // server unavailable
		{rawURL: "https://234.112.93.22:4123", url: "", err: errors.New("server unavailable")},        // empty
		{rawURL: "http://234.112.93.22:4123", url: "", err: errors.New("server unavailable")},         // empty
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Raw URL: %s", tc.rawURL), func() {
			url, err := resolveURL(tc.rawURL)
			s.Equal(tc.err, err)
			s.Equal(tc.url, url)
		})
	}
}

func (s *DownloadSuite) TestDownload_String() {
	d := &Download{ID: downloadID, URL: downloadURL, Size: 4.3243 * fsutil.GB}
	s.Equal(fmt.Sprintf(" ⁕ %s ⇒ URL: %s Size: 4.3 GB\n", downloadID, downloadURL), d.String())
}

func (s *DownloadSuite) TestDownload_Delete() {

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
