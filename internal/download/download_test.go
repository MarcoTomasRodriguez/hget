package download

import (
	"github.com/MarcoTomasRodriguez/hget/pkg/httputil"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

var golangSample = &Download{
	Id:       "v5pra7bt",
	Name:     "go1.19.1.src.tar.gz",
	URL:      "https://go.dev/dl/go1.19.1.src.tar.gz",
	Size:     1300,
	Segments: []Segment{{0, 0, 1300}},
}

var javaSample = &Download{
	Id:   "ita2qybt",
	Name: "jre-8u351-macosx-x64.dmg",
	URL:  "https://java.com/download/jre/jre-8u351-macosx-x64.dmg",
	Size: 2583,
	Segments: []Segment{
		{0, 0, 644},
		{1, 645, 1289},
		{2, 1290, 1934},
		{3, 1935, 2583},
	},
}

type DownloaderSuite struct {
	suite.Suite
}

func (s *DownloaderSuite) SetupSuite() {
	httpmock.Activate()
}

func (s *DownloaderSuite) TearDownTest() {
	httpmock.Reset()
}

func (s *DownloaderSuite) TearDownSuite() {
	httpmock.DeactivateAndReset()
}

func (s *DownloaderSuite) TestNewDownload_ShouldLoadAllProperties() {
	httputil.RegisterResponder(javaSample.URL, make([]byte, javaSample.Size), http.Header{"Accept-Ranges": []string{"bytes"}})

	download, err := NewDownload(javaSample.URL, 4)

	s.NoError(err)
	s.Len(download.Id, 8)

	download.Id = javaSample.Id
	s.Equal(javaSample, download)
}

func (s *DownloaderSuite) TestNewDownload_ShouldFailIfInvalidURL() {
	download, err := NewDownload("ftp://go.dev/dl/go1.19.1.src.tar.gz", 8)

	s.Nil(download)
	s.ErrorIs(err, httputil.InvalidUrlErr)
}

func (s *DownloaderSuite) TestNewDownload_ShouldFailIfNotFound() {
	download, err := NewDownload("https://urlshouldnotrespondotherwisefail.com/dl/filename.ext", 8)

	s.Nil(download)
	s.Error(err)
}

func (s *DownloaderSuite) TestNewDownload_ShouldFailIfInvalidFilename() {
	url := "https://go.dev/dl/-invalid.filename"

	httputil.RegisterResponder(url, make([]byte, golangSample.Size), http.Header{"Accept-Ranges": []string{"bytes"}})

	download, err := NewDownload(url, 8)

	s.Nil(download)
	s.EqualError(err, "invalid filename")
}

func (s *DownloaderSuite) TestNewDownload_ShouldDisableRangeDownload() {
	httputil.RegisterResponder(golangSample.URL, make([]byte, golangSample.Size), http.Header{})

	download, err := NewDownload(golangSample.URL, 8)

	s.NoError(err)
	s.Equal(golangSample.Name, download.Name)
	s.Equal(golangSample.Segments, download.Segments)
}

func (s *DownloaderSuite) TestNewDownload_ShouldUseContentDispositionFilename() {
	url := "https://go.dev/dl/go1.19.1.src.tar.gz"
	header := http.Header{
		"Accept-Ranges":       []string{"bytes"},
		"Content-Disposition": []string{"attachment; filename=\"filename.tar.gz\""},
	}

	httputil.RegisterResponder(url, make([]byte, golangSample.Size), header)

	download, err := NewDownload(url, 8)

	s.NoError(err)
	s.Equal("filename.tar.gz", download.Name)
}

func (s *DownloaderSuite) TestDownload_String() {
	url := "https://go.dev/dl/go1.19.1.src.tar.gz"
	header := http.Header{
		"Accept-Ranges":       []string{"bytes"},
		"Content-Disposition": []string{"attachment; filename=\"filename.tar.gz\""},
	}

	httputil.RegisterResponder(url, make([]byte, golangSample.Size), header)

	download, err := NewDownload(url, 8)

	s.NoError(err)
	s.Equal(" ⁕ 52fdfc07 ⇒ URL: https://go.dev/dl/go1.19.1.src.tar.gz Size: 1.3 kB\n", download.String())
}

func TestDownloadSuite(t *testing.T) {
	suite.Run(t, new(DownloaderSuite))
}
