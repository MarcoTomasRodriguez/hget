package download

import (
	"bytes"
	"context"
	"github.com/MarcoTomasRodriguez/hget/pkg/httputil"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"net/http"
	"testing"
)

type SegmentSuite struct {
	suite.Suite
}

func (s *SegmentSuite) SetupSuite() {
	httpmock.Activate()
}

func (s *SegmentSuite) TearDownTest() {
	httpmock.Reset()
}

func (s *SegmentSuite) TearDownSuite() {
	httpmock.DeactivateAndReset()
}

func (s *SegmentSuite) TestSegment_Filename_ShouldPrependAZeroWhenIdHasOneDigit() {
	segment := &Segment{Id: 2}
	s.Equal("segment.02", segment.Filename())
}

func (s *SegmentSuite) TestSegment_Filename_ShouldDisplayFullIdWhenItHasTwoDigits() {
	segment := &Segment{Id: 13}
	s.Equal("segment.13", segment.Filename())
}

func (s *SegmentSuite) TestSegment_Download() {
	body := make([]byte, javaSample.Size)
	rand.Read(body)

	httputil.RegisterResponder(javaSample.URL, body, http.Header{"Accept-Ranges": []string{"bytes"}})

	buffer := new(bytes.Buffer)
	segment := javaSample.Segments[1]

	err := segment.Download(javaSample.URL, segment.Start, buffer, context.TODO())

	s.NoError(err)
	s.Equal(body[segment.Start:segment.End+1], buffer.Bytes())
}

func (s *SegmentSuite) TestSegment_Download_ShouldDoNothingIfPositionIsEnd() {
	body := make([]byte, javaSample.Size)
	rand.Read(body)

	httputil.RegisterResponder(javaSample.URL, body, http.Header{"Accept-Ranges": []string{"bytes"}})

	buffer := new(bytes.Buffer)
	segment := javaSample.Segments[1]

	err := segment.Download(javaSample.URL, segment.End, buffer, context.TODO())

	s.NoError(err)
	s.Equal(0, buffer.Len())
}

func (s *SegmentSuite) TestSegment_Download_ShouldFailIfPositionExceedsRange() {
	body := make([]byte, javaSample.Size)
	rand.Read(body)

	httputil.RegisterResponder(javaSample.URL, body, http.Header{"Accept-Ranges": []string{"bytes"}})

	buffer := new(bytes.Buffer)
	segment := javaSample.Segments[1]

	err := segment.Download(javaSample.URL, segment.End+1, buffer, context.TODO())

	s.ErrorIs(err, SegmentOverflowError{})
	s.Equal(0, buffer.Len())
}

func (s *SegmentSuite) TestSegment_Download_ShouldFailIfServerNotAvailable() {
	buffer := new(bytes.Buffer)
	segment := javaSample.Segments[1]

	err := segment.Download("th1ss1t3sh0uldn0tex1st.test/path/to/file.txt", segment.Start, buffer, context.TODO())

	s.IsType(NetworkError(""), err)
	s.Equal(0, buffer.Len())
}
func TestSegmentSuite(t *testing.T) {
	suite.Run(t, new(SegmentSuite))
}
