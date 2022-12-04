package download

import (
	"context"
	"fmt"
	"github.com/MarcoTomasRodriguez/hget/pkg/httputil"
	"github.com/jarcoal/httpmock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
	"math/rand"
	"net/http"
	"os"
	"testing"
)

type ManagerSuite struct {
	suite.Suite
}

func (s *ManagerSuite) SetupSuite() {
	httpmock.Activate()
}

func (s *ManagerSuite) TearDownTest() {
	httpmock.Reset()
}

func (s *ManagerSuite) TearDownSuite() {
	httpmock.DeactivateAndReset()
}

func (s *ManagerSuite) TestManager_GetDownloadById() {
	fs := afero.NewMemMapFs()
	afs := afero.Afero{Fs: fs}

	golangSampleYml, _ := yaml.Marshal(golangSample)
	javaSampleYml, _ := yaml.Marshal(javaSample)

	_ = afs.WriteFile("downloads/"+golangSample.Id+"/download.yml", golangSampleYml, os.ModePerm)
	_ = afs.WriteFile("downloads/"+javaSample.Id+"/download.yml", javaSampleYml, os.ModePerm)

	m := NewManager(fs)
	download, err := m.GetDownloadById(golangSample.Id)

	s.NoError(err)
	s.Equal(golangSample, download)
}

func (s *ManagerSuite) TestManager_GetDownloadByUrl() {
	fs := afero.NewMemMapFs()
	afs := afero.Afero{Fs: fs}

	golangSampleYml, _ := yaml.Marshal(golangSample)
	javaSampleYml, _ := yaml.Marshal(javaSample)

	_ = afs.WriteFile("downloads/"+golangSample.Id+"/download.yml", golangSampleYml, os.ModePerm)
	_ = afs.WriteFile("downloads/"+javaSample.Id+"/download.yml", javaSampleYml, os.ModePerm)

	m := NewManager(fs)
	download, err := m.GetDownloadById(golangSample.Id)

	s.NoError(err)
	s.Equal(golangSample, download)
}

func (s *ManagerSuite) TestManager_ListDownloads() {
	fs := afero.NewMemMapFs()
	afs := afero.Afero{Fs: fs}

	golangSampleYml, _ := yaml.Marshal(golangSample)
	javaSampleYml, _ := yaml.Marshal(javaSample)

	_ = afs.WriteFile("downloads/"+golangSample.Id+"/download.yml", golangSampleYml, os.ModePerm)
	_ = afs.WriteFile("downloads/"+javaSample.Id+"/download.yml", javaSampleYml, os.ModePerm)

	m := NewManager(fs)
	downloads, err := m.ListDownloads()

	s.NoError(err)
	s.Equal(javaSample, downloads[0])
	s.Equal(golangSample, downloads[1])
}

func (s *ManagerSuite) TestManager_DeleteDownload() {
	fs := afero.NewMemMapFs()
	afs := afero.Afero{Fs: fs}

	golangSampleYml, _ := yaml.Marshal(golangSample)
	javaSampleYml, _ := yaml.Marshal(javaSample)

	_ = afs.WriteFile("downloads/"+golangSample.Id+"/download.yml", golangSampleYml, os.ModePerm)
	_ = afs.WriteFile("downloads/"+javaSample.Id+"/download.yml", javaSampleYml, os.ModePerm)

	m := NewManager(fs)
	err := m.DeleteDownloadById(golangSample.Id)

	s.NoError(err)

	exists, err := afs.DirExists(golangSample.Name)
	s.NoError(err)
	s.False(exists)
}

func (s *ManagerSuite) TestManager_StartDownload() {
	ctx := context.TODO()
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	content := make([]byte, javaSample.Size)
	rand.Read(content)

	httputil.RegisterResponder(javaSample.URL, content, http.Header{"Accept-Ranges": []string{"bytes"}})

	manager := NewManager(fs)
	download, _ := NewDownload(javaSample.URL, 4)

	err := manager.StartDownload(download, ctx)
	s.NoError(err)

	fileContent, _ := fs.ReadFile(fmt.Sprintf("downloads/%s/%s", download.Id, javaSample.Name))
	s.Equal(content, fileContent)
}

func TestManagerSuite(t *testing.T) {
	suite.Run(t, new(ManagerSuite))
}
