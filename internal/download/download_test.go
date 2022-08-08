package download

import (
	"fmt"
	"github.com/MarcoTomasRodriguez/hget/internal/config"
	"github.com/MarcoTomasRodriguez/hget/pkg/fsutil"
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

func (s *DownloadSuite) SetupTest() {
	cfg := &config.Config{}
	cfg.ProgramFolder = programFolder
	cfg.Download.Folder = downloadFolder

	do.ProvideValue[*config.Config](nil, cfg)
	do.ProvideValue[*afero.Afero](nil, &afero.Afero{Fs: afero.NewMemMapFs()})
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
