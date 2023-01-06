// Note: The following tests correspond to a previous version, where mocking each part was not possible.
// The code was refactored for better maintainability and testability, so these tests are temporary and can be
// improved using mock implementations.
package download_test

import (
	"context"
	"fmt"
	"github.com/MarcoTomasRodriguez/hget/internal/download"
	"github.com/MarcoTomasRodriguez/hget/mocks"
	"github.com/MarcoTomasRodriguez/hget/pkg/httputil"
	"github.com/MarcoTomasRodriguez/hget/pkg/logger"
	"github.com/MarcoTomasRodriguez/hget/pkg/progressbar"
	"github.com/jarcoal/httpmock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"net/http"
	"testing"
)

type DownloaderSuite struct {
	suite.Suite
	network  *mocks.Network
	storage  *mocks.Storage
	logger   logger.Logger
	progress progressbar.ProgressBar
}

func (s *DownloaderSuite) SetupTest() {
	s.network = new(mocks.Network)
	s.storage = new(mocks.Storage)
	s.logger = logger.NoopConsoleLogger{}
	s.progress = progressbar.NoopProgressBar{}
}

func (s *DownloaderSuite) TestDownloader_InitDownload_ShouldLoadAllProperties() {
	downloader := download.NewDownloader(s.network, s.storage, s.progress, s.logger)
	s.network.On("FetchResource", javaSample.URL).Return(javaResource, nil)

	spec, err := downloader.InitDownload(javaSample.URL, 4)
	s.NoError(err)
	s.Len(spec.Id, 8)

	s.Equal(javaSample.Name, spec.Name)
	s.Equal(javaSample.URL, spec.URL)
	s.Equal(javaSample.Size, spec.Size)

	for i := range spec.Segments {
		expected := javaSample.Segments[i]
		actual := spec.Segments[i]

		s.Equal(expected.Start, actual.Start)
		s.Equal(expected.End, actual.End)
	}
}

func (s *DownloaderSuite) TestDownloader_InitDownload_ShouldFailIfInvalidURL() {
	downloader := download.NewDownloader(s.network, s.storage, s.progress, s.logger)
	s.network.On("FetchResource", "ftp://go.dev/dl/go1.19.1.src.tar.gz").Return(download.Resource{}, httputil.InvalidUrlErr)

	spec, err := downloader.InitDownload("ftp://go.dev/dl/go1.19.1.src.tar.gz", 8)
	s.Empty(spec)
	s.ErrorIs(err, httputil.InvalidUrlErr)
}

func (s *DownloaderSuite) TestDownloader_InitDownload_ShouldFailIfNotFound() {
	downloader := download.NewDownloader(s.network, s.storage, s.progress, s.logger)
	s.network.On("FetchResource", "https://test.com/dl/filename.ext").Return(download.Resource{}, httputil.ServerNotAvailableErr)

	spec, err := downloader.InitDownload("https://test.com/dl/filename.ext", 8)
	s.Empty(spec)
	s.ErrorIs(err, httputil.ServerNotAvailableErr)
}

func (s *DownloaderSuite) TestDownloader_InitDownload_ShouldFailIfInvalidFilename() {
	downloader := download.NewDownloader(s.network, s.storage, s.progress, s.logger)
	s.network.On("FetchResource", "https://go.dev/dl/-invalid.filename").Return(download.Resource{}, download.InvalidFilenameErr)

	spec, err := downloader.InitDownload("https://go.dev/dl/-invalid.filename", 8)
	s.Empty(spec)
	s.ErrorIs(err, download.InvalidFilenameErr)
}

func (s *DownloaderSuite) TestDownloader_InitDownload_ShouldDisableRangeDownload() {
	resource := golangResource
	resource.AcceptRanges = false

	downloader := download.NewDownloader(s.network, s.storage, s.progress, s.logger)
	s.network.On("FetchResource", golangSample.URL).Return(resource, nil)

	spec, err := downloader.InitDownload(golangSample.URL, 8)
	s.NoError(err)
	s.Equal(golangSample.Name, spec.Name)
	s.Equal(golangSample.Segments[0].Start, spec.Segments[0].Start)
	s.Equal(golangSample.Segments[0].End, spec.Segments[0].End)
}

func (s *DownloaderSuite) TestDownloader_Download() {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	fs := afero.NewMemMapFs()
	afs := afero.Afero{Fs: fs}
	downloader := download.NewDownloader(download.NewNetwork(), download.NewStorage(fs), s.progress, s.logger)

	content := make([]byte, javaSample.Size)
	rand.Read(content)

	httputil.RegisterResponder(javaSample.URL, content, http.Header{"Accept-Ranges": []string{"bytes"}})

	err := downloader.Download(javaSample, context.TODO())
	s.NoError(err)

	fileContent, _ := afs.ReadFile(fmt.Sprintf("%s/output", javaSample.Id))
	s.Equal(content, fileContent)
}

func (s *DownloaderSuite) TestDownloader_GetDownloadByUrl() {
	s.storage.On("ListDownloads").Return([]download.Download{golangSample, javaSample}, nil)

	downloader := download.NewDownloader(s.network, s.storage, s.progress, s.logger)
	spec, err := downloader.FindDownloadByUrl(javaSample.URL)

	s.NoError(err)
	s.Equal(javaSample, spec)
}

func TestDownloaderSuite(t *testing.T) {
	suite.Run(t, new(DownloaderSuite))
}
