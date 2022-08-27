package download

import (
	"context"
	"fmt"
	"github.com/MarcoTomasRodriguez/hget/internal/config"
	"github.com/MarcoTomasRodriguez/hget/pkg/fsutil"
	"github.com/jarcoal/httpmock"
	"github.com/samber/do"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

type WorkerSuite struct {
	suite.Suite
}

func (s *WorkerSuite) SetupSuite() {
	httpmock.Activate()
	httpmock.ActivateNonDefault(httpClient)
}

func (s *WorkerSuite) TearDownTest() {
	httpmock.Reset()
}

func (s *WorkerSuite) TearDownSuite() {
	httpmock.DeactivateAndReset()
}

func (s *WorkerSuite) TestNewWorker() {
	testCases := []struct {
		downloadSize  int64
		workerIndex   int
		workerCount   int64
		startingPoint int64
		endPoint      int64
	}{
		{downloadSize: 15.3 * fsutil.MB, workerIndex: 0, workerCount: 1, startingPoint: 0, endPoint: 15.3 * fsutil.MB},
		{downloadSize: 15.3 * fsutil.MB, workerIndex: 1, workerCount: 2, startingPoint: 7.65 * fsutil.MB, endPoint: 15.3 * fsutil.MB},
		{downloadSize: 15.3 * fsutil.MB, workerIndex: 3, workerCount: 7, startingPoint: 6.557142 * fsutil.MB, endPoint: 8.742855 * fsutil.MB},
		{downloadSize: 1.534 * fsutil.GB, workerIndex: 0, workerCount: 8, startingPoint: 0, endPoint: 0.19175*fsutil.GB - 1},
		{downloadSize: 1.534 * fsutil.GB, workerIndex: 5, workerCount: 8, startingPoint: 0.95875 * fsutil.GB, endPoint: 1.1505*fsutil.GB - 1},
		{downloadSize: 1.534 * fsutil.GB, workerIndex: 7, workerCount: 8, startingPoint: 1.34225 * fsutil.GB, endPoint: 1.534 * fsutil.GB},
		{downloadSize: 1.534 * fsutil.GB, workerIndex: 0, workerCount: 1, startingPoint: 0, endPoint: 1.534 * fsutil.GB},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Size: %d - Index: %d - Workers: %d", tc.downloadSize, tc.workerIndex, tc.workerCount), func() {
			worker := NewWorker(tc.workerIndex, &Download{Size: tc.downloadSize, Workers: make([]*Worker, tc.workerCount)})

			s.Equal(tc.workerIndex, worker.ID)
			s.Equal(tc.startingPoint, worker.StartingPoint)
			s.Equal(tc.endPoint, worker.EndPoint)
		})
	}
}

func (s *WorkerSuite) TestWorker_Execute() {
	ctx := context.TODO()
	fs := &afero.Afero{Fs: afero.NewMemMapFs()}
	cfg := &config.Config{ProgramFolder: "/home/user/.hget"}
	cfg.Download.CopyNBytes = 250

	id := "fdc134c5f503b1bd-go1.17.2.src.tar.gz"
	url := "https://golang.org/dl/go1.17.2.src.tar.gz"
	eTag := "test"

	body := randBytes(51234)

	do.ProvideValue[*config.Config](nil, cfg)
	do.ProvideValue[*afero.Afero](nil, fs)

	httpmock.RegisterResponder("GET", downloadURL, func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewBytesResponse(200, body)
		resp.Header.Set("ETag", eTag)
		return resp, nil
	})

	worker := &Worker{
		ID:            0,
		StartingPoint: 0,
		EndPoint:      int64(len(body)),
		download:      &Download{ID: id, URL: url, ETag: eTag},
	}

	err := worker.Execute(ctx, nil)

	// Assert that no error occurred.
	s.NoError(err)

	// Assert that the mocked endpoint was called.
	s.Equal(1, httpmock.GetCallCountInfo()["GET "+downloadURL])

	// Assert that the worker file exists and contains the response body.
	file, err := fs.ReadFile("/home/user/.hget/downloads/fdc134c5f503b1bd-go1.17.2.src.tar.gz/worker.00000")
	s.NoError(err)
	s.Equal(file, body)
}

func (s *WorkerSuite) TestWorker_filePath() {
	testCases := []struct {
		downloadID string
		workerID   int
		expected   string
	}{
		{"fdc134c5f503b1bd-go1.17.2.src.tar.gz", 0, "fdc134c5f503b1bd-go1.17.2.src.tar.gz/worker.00000"},
		{"fdc134c5f503b1bd-go1.17.2.src.tar.gz", 1, "fdc134c5f503b1bd-go1.17.2.src.tar.gz/worker.00001"},
		{"fdc134c5f503b1bd-go1.17.2.src.tar.gz", 7, "fdc134c5f503b1bd-go1.17.2.src.tar.gz/worker.00007"},
		{"fdc134c5f503b1bd-go1.17.2.src.tar.gz", 12, "fdc134c5f503b1bd-go1.17.2.src.tar.gz/worker.00012"},
		{"fdc134c5f503b1bd-go1.17.2.src.tar.gz", 1435, "fdc134c5f503b1bd-go1.17.2.src.tar.gz/worker.01435"},
		{"fdc134c5f503b1bd-go1.17.2.src.tar.gz", 99999, "fdc134c5f503b1bd-go1.17.2.src.tar.gz/worker.99999"},
	}

	for _, tc := range testCases {
		s.Run(tc.expected, func() {
			do.ProvideValue[*config.Config](nil, &config.Config{ProgramFolder: "/home/user/.hget"})
			worker := &Worker{ID: tc.workerID, download: &Download{ID: tc.downloadID}}

			s.Equal("/home/user/.hget/downloads/"+tc.expected, worker.filePath())
		})
	}
}

func (s *WorkerSuite) TestWorker_fileSize() {
	fs := &afero.Afero{Fs: afero.NewMemMapFs()}
	do.ProvideValue[*config.Config](nil, &config.Config{ProgramFolder: "/home/user/.hget"})
	do.ProvideValue[*afero.Afero](nil, fs)

	worker := &Worker{ID: 0, download: &Download{ID: "fdc134c5f503b1bd-go1.17.2.src.tar.gz"}}
	_ = fs.WriteFile("/home/user/.hget/downloads/fdc134c5f503b1bd-go1.17.2.src.tar.gz/worker.00000", make([]byte, 51233), 0755)
	s.Equal(int64(51233), worker.fileSize())

	nonexistentWorker := &Worker{ID: 1, download: &Download{ID: "fdc134c5f503b1bd-go1.17.2.src.tar.gz"}}
	s.Equal(int64(0), nonexistentWorker.fileSize())
}

func TestWorkerSuite(t *testing.T) {
	suite.Run(t, new(WorkerSuite))
}
