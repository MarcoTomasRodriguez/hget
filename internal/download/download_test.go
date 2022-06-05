package download

import (
	"github.com/MarcoTomasRodriguez/hget/internal/config"
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

	do.ProvideValue[*config.Config](do.DefaultInjector, cfg)
	do.ProvideValue[afero.Fs](do.DefaultInjector, afero.NewMemMapFs())
}

func (s *DownloadSuite) TestDownloadFilePath() {
	d := &Download{Name: downloadName}
	s.Equal(filepath.Join(downloadFolder, downloadName), d.DownloadFilePath())
}

func (s *DownloadSuite) TestFolderPath() {
	d := &Download{ID: downloadID}
	s.Equal(d.FolderPath(), filepath.Join(programFolder, "downloads", downloadID))
}

func (s *DownloadSuite) TestFilePath() {
	d := &Download{ID: downloadID}
	s.Equal(d.FilePath(), filepath.Join(programFolder, "downloads", downloadID, "download.toml"))
}

func TestDownloadSuite(t *testing.T) {
	suite.Run(t, new(DownloadSuite))
}
