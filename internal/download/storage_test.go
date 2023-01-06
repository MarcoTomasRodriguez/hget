// Note: The following tests correspond to a previous version, where mocking each part was not possible.
// The code was refactored for better maintainability and testability, so these tests are temporary and can be
// improved using mock implementations.
package download_test

import (
	"github.com/MarcoTomasRodriguez/hget/internal/download"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
	"os"
	"testing"
)

type StorageSuite struct {
	suite.Suite
	afs     afero.Afero
	storage download.Storage
}

func (s *StorageSuite) SetupTest() {
	s.afs = afero.Afero{Fs: afero.NewMemMapFs()}
	s.storage = download.NewStorage(s.afs.Fs)
}

func (s *StorageSuite) TestStorage_ReadDownloadSpec() {
	golangSampleYml, _ := yaml.Marshal(golangSample)
	javaSampleYml, _ := yaml.Marshal(javaSample)

	_ = s.afs.WriteFile(golangSample.Id+"/download.yml", golangSampleYml, os.ModePerm)
	_ = s.afs.WriteFile(javaSample.Id+"/download.yml", javaSampleYml, os.ModePerm)

	spec, err := s.storage.ReadDownloadSpec(golangSample.Id)
	s.NoError(err)
	s.Equal(golangSample, spec)
}

func (s *StorageSuite) TestStorage_ListDownloads() {
	golangSampleYml, _ := yaml.Marshal(golangSample)
	javaSampleYml, _ := yaml.Marshal(javaSample)

	_ = s.afs.WriteFile(golangSample.Id+"/download.yml", golangSampleYml, os.ModePerm)
	_ = s.afs.WriteFile(javaSample.Id+"/download.yml", javaSampleYml, os.ModePerm)

	specs, err := s.storage.ListDownloads()
	s.NoError(err)
	s.Equal(javaSample, specs[0])
	s.Equal(golangSample, specs[1])
}

func (s *StorageSuite) TestStorage_DeleteDownload() {
	golangSampleYml, _ := yaml.Marshal(golangSample)
	javaSampleYml, _ := yaml.Marshal(javaSample)

	_ = s.afs.WriteFile(golangSample.Id+"/download.yml", golangSampleYml, os.ModePerm)
	_ = s.afs.WriteFile(javaSample.Id+"/download.yml", javaSampleYml, os.ModePerm)

	err := s.storage.DeleteDownload(golangSample.Id)
	s.NoError(err)

	exists, err := s.afs.DirExists(golangSample.Name)
	s.NoError(err)
	s.False(exists)
}

func TestStorageSuite(t *testing.T) {
	suite.Run(t, new(StorageSuite))
}
