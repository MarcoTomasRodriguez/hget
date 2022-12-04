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

func (s *SegmentSuite) TestSegmentdownload() {
	body := make([]byte, golangSample.Size)
	rand.Read(body)

	httputil.RegisterResponder(golangSample.URL, body, http.Header{"Accept-Ranges": []string{"bytes"}})

	buffer := new(bytes.Buffer)
	segment := Segment{Id: 1, Start: 1234, End: 4321}

	err := segment.Download(golangSample.URL, 1234, buffer, context.TODO())
	s.NoError(err)

	s.Equal(body[segment.Start:segment.End+1], buffer.Bytes())
}

func TestSegmentSuite(t *testing.T) {
	suite.Run(t, new(SegmentSuite))
}
