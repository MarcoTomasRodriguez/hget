package httputil

import (
	"bytes"
	"fmt"
	"github.com/jarcoal/httpmock"
	"io"
	"net/http"
	"regexp"
	"strconv"
)

var (
	InvalidUrlErr         = fmt.Errorf("invalid url")
	ServerNotAvailableErr = fmt.Errorf("server not available")
)

var urlRegex = regexp.MustCompile("^(https?://)?[-A-Za-z\\d+&@#/%?=~_|!,.;]+[-A-Za-z\\d+&@#/%=~_|]$")

// ResolveURL resolves the url adding the http scheme, preferring https over http.
func ResolveURL(rawURL string) (string, error) {
	urlParts := urlRegex.FindStringSubmatch(rawURL)

	// Check if rawURL matches the regex.
	if len(urlParts) < 1 {
		return "", InvalidUrlErr
	}

	// If scheme is provided, attempt to execute a request.
	if urlParts[1] != "" {
		if _, err := http.Get(rawURL); err != nil {
			return "", ServerNotAvailableErr
		}

		return rawURL, nil
	}

	// Prefer https over http.
	if url, err := ResolveURL("https://" + rawURL); err == nil {
		return url, nil
	}

	// If not available over https, try with http.
	return ResolveURL("http://" + rawURL)
}

// RegisterResponder registers a mock HTTP GET responder with support for ranges.
func RegisterResponder(url string, body []byte, header http.Header) {
	httpmock.RegisterResponder("GET", url, func(request *http.Request) (*http.Response, error) {
		start := 0
		end := len(body)

		rangeHeader := request.Header.Get("Range")
		if rangeHeader != "" {
			regex, _ := regexp.Compile("^(bytes=(\\d+)-(\\d+))$")

			rangeHeaderParsed := regex.FindStringSubmatch(rangeHeader)[2:4]
			start, _ = strconv.Atoi(rangeHeaderParsed[0])
			end, _ = strconv.Atoi(rangeHeaderParsed[1])

			// Bytes range is inclusive, whereas slicing is exclusive.
			// Thus, add one to the end range if it isn't greater than the body length.
			if end < len(body) {
				end++
			}
		}

		body := io.NopCloser(bytes.NewReader(body[start:end]))

		return &http.Response{
			ContentLength: int64(end - start),
			Body:          body,
			Header:        header,
		}, nil
	})
}
