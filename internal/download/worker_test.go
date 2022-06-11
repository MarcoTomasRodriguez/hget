package download

import (
	"fmt"
	"github.com/MarcoTomasRodriguez/hget/internal/config"
	"github.com/MarcoTomasRodriguez/hget/pkg/fsutil"
	"github.com/samber/do"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewWorker(t *testing.T) {
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
		t.Run(fmt.Sprintf("Size: %d - Index: %d - Workers: %d", tc.downloadSize, tc.workerIndex, tc.workerCount), func(t *testing.T) {
			worker := NewWorker(tc.workerIndex, &Download{Size: tc.downloadSize, Workers: make([]*Worker, tc.workerCount)})

			assert.Equal(t, tc.workerIndex, worker.ID)
			assert.Equal(t, tc.startingPoint, worker.StartingPoint)
			assert.Equal(t, tc.endPoint, worker.EndPoint)
		})
	}
}

func TestWorker_filePath(t *testing.T) {
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
		t.Run(tc.expected, func(t *testing.T) {
			do.ProvideValue[*config.Config](nil, &config.Config{ProgramFolder: "/home/user/.hget"})
			worker := &Worker{ID: tc.workerID, download: &Download{ID: tc.downloadID}}

			assert.Equal(t, "/home/user/.hget/downloads/"+tc.expected, worker.filePath())
		})
	}
}

func TestWorker_fileSize(t *testing.T) {
	fs := &afero.Afero{Fs: afero.NewMemMapFs()}
	do.ProvideValue[*config.Config](nil, &config.Config{ProgramFolder: "/home/user/.hget"})
	do.ProvideValue[*afero.Afero](nil, fs)

	worker := &Worker{ID: 0, download: &Download{ID: "fdc134c5f503b1bd-go1.17.2.src.tar.gz"}}

	_ = fs.WriteFile("/home/user/.hget/downloads/fdc134c5f503b1bd-go1.17.2.src.tar.gz/worker.00000", make([]byte, 51233), 0755)
	assert.Equal(t, int64(51233), worker.fileSize())
}
