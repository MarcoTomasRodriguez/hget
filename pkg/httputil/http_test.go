package httputil

import (
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

const testUrl = "th1ss1t3sh0uldn0tex1st.test/path/to/file.txt"

type HttpUtilSuite struct {
	suite.Suite
}

func (s *HttpUtilSuite) SetupSuite() {
	httpmock.Activate()
}

func (s *HttpUtilSuite) TearDownTest() {
	httpmock.Reset()
}

func (s *HttpUtilSuite) TearDownSuite() {
	httpmock.DeactivateAndReset()
}

func (s *HttpUtilSuite) TestResolveURL_WithHttpsScheme() {
	RegisterResponder("https://"+testUrl, []byte{}, http.Header{})
	RegisterResponder("http://"+testUrl, []byte{}, http.Header{})
	url, err := ResolveURL("https://" + testUrl)

	s.NoError(err)
	s.Equal("https://"+testUrl, url)
}

func (s *HttpUtilSuite) TestResolveURL_WithHttpScheme() {
	RegisterResponder("https://"+testUrl, []byte{}, http.Header{})
	RegisterResponder("http://"+testUrl, []byte{}, http.Header{})
	url, err := ResolveURL("http://" + testUrl)

	s.NoError(err)
	s.Equal("http://"+testUrl, url)
}

func (s *HttpUtilSuite) TestResolveURL_WithoutScheme_ResolveHttps() {
	RegisterResponder("https://"+testUrl, []byte{}, http.Header{})
	RegisterResponder("http://"+testUrl, []byte{}, http.Header{})
	url, err := ResolveURL(testUrl)

	s.NoError(err)
	s.Equal("https://"+testUrl, url)
}

func (s *HttpUtilSuite) TestResolveURL_WithoutScheme_ResolveHttp() {
	RegisterResponder("http://"+testUrl, []byte{}, http.Header{})
	url, err := ResolveURL(testUrl)

	s.NoError(err)
	s.Equal("http://"+testUrl, url)
}

func (s *HttpUtilSuite) TestResolveURL_ServerNotAvailable() {
	url, err := ResolveURL("https://" + testUrl)

	s.ErrorIs(err, URLCannotBeResolvedError("server not available"))
	s.Empty(url)
}

func (s *HttpUtilSuite) TestResolveURL_InvalidScheme() {
	url, err := ResolveURL("ftp://" + testUrl)

	s.ErrorIs(err, URLCannotBeResolvedError("invalid scheme"))
	s.Empty(url)
}

func (s *HttpUtilSuite) TestResolveURL_InvalidUrl() {
	url, err := ResolveURL("")

	s.ErrorIs(err, URLCannotBeResolvedError("invalid url"))
	s.Empty(url)
}

func (s *HttpUtilSuite) TestResolveURL_CannotFindMatchingScheme() {
	url, err := ResolveURL(testUrl)

	s.ErrorIs(err, URLCannotBeResolvedError("cannot find a matching scheme"))
	s.Empty(url)
}

func TestHttpUtilSuite(t *testing.T) {
	suite.Run(t, new(HttpUtilSuite))
}
