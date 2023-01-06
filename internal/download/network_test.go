package download_test

import (
	"bytes"
	"context"
	"github.com/MarcoTomasRodriguez/hget/internal/download"
	"github.com/MarcoTomasRodriguez/hget/pkg/httputil"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"net/http"
	"testing"
)

var golangResource = download.Resource{
	Filename:     "go1.19.1.src.tar.gz",
	URL:          "https://go.dev/dl/go1.19.1.src.tar.gz",
	Size:         1300,
	AcceptRanges: false,
}

var javaResource = download.Resource{
	Filename:     "jre-8u351-macosx-x64.dmg",
	URL:          "https://java.com/download/jre/jre-8u351-macosx-x64.dmg",
	Size:         2583,
	AcceptRanges: true,
}

type NetworkSuite struct {
	suite.Suite
}

func (s *NetworkSuite) SetupSuite() {
	httpmock.Activate()
}

func (s *NetworkSuite) TearDownTest() {
	httpmock.Reset()
}

func (s *NetworkSuite) TearDownSuite() {
	httpmock.DeactivateAndReset()
}

func (s *NetworkSuite) TestNetwork_DownloadResource() {
	network := download.NewNetwork()

	body := make([]byte, javaSample.Size)
	rand.Read(body)
	httputil.RegisterResponder(javaSample.URL, body, http.Header{"Accept-Ranges": []string{"bytes"}})

	buffer := new(bytes.Buffer)
	segment := javaSample.Segments[1]

	err := network.DownloadResource(javaSample.URL, segment.Start, segment.End, buffer, context.TODO())
	s.NoError(err)
	s.Equal(body[segment.Start:segment.End+1], buffer.Bytes())
}

func (s *NetworkSuite) TestNetwork_DownloadResource_ShouldDoNothingIfAlreadyFinished() {
	network := download.NewNetwork()

	body := make([]byte, javaSample.Size)
	rand.Read(body)
	httputil.RegisterResponder(javaSample.URL, body, http.Header{"Accept-Ranges": []string{"bytes"}})

	buffer := new(bytes.Buffer)
	segment := javaSample.Segments[1]

	err := network.DownloadResource(javaSample.URL, segment.End, segment.End, buffer, context.TODO())
	s.NoError(err)
	s.Equal(0, buffer.Len())
}

func (s *NetworkSuite) TestNetwork_DownloadResource_ShouldFailIfPositionExceedsRange() {
	network := download.NewNetwork()

	body := make([]byte, javaSample.Size)
	rand.Read(body)
	httputil.RegisterResponder(javaSample.URL, body, http.Header{"Accept-Ranges": []string{"bytes"}})

	buffer := new(bytes.Buffer)
	segment := javaSample.Segments[1]

	err := network.DownloadResource(javaSample.URL, segment.End+1, segment.End, buffer, context.TODO())
	s.ErrorIs(err, download.SegmentOverflowErr)
	s.Equal(0, buffer.Len())
}

func (s *NetworkSuite) TestNetwork_DownloadResource_ShouldFailIfServerNotAvailable() {
	network := download.NewNetwork()

	buffer := new(bytes.Buffer)
	segment := javaSample.Segments[1]

	err := network.DownloadResource("th1ss1t3sh0uldn0tex1st.test/path/to/file.txt", segment.Start, segment.End, buffer, context.TODO())
	s.Error(err)
	s.Equal(0, buffer.Len())
}

func TestNetworkSuite(t *testing.T) {
	suite.Run(t, new(NetworkSuite))
}
