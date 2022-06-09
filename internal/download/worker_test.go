package download

import (
	"github.com/MarcoTomasRodriguez/hget/internal/config"
	"github.com/samber/do"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"testing"
)

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
