package httputil

import (
	"bytes"
	"github.com/jarcoal/httpmock"
	"io"
	"net/http"
	"regexp"
	"strconv"
)

type URLCannotBeResolvedError string

func (e URLCannotBeResolvedError) Error() string {
	return "url cannot be resolved: " + string(e)
}

var urlRegex = regexp.MustCompile("^([A-Za-z]+://)?[-A-Za-z\\d+&@#/%?=~_|!:,.;]+[-A-Za-z\\d+&@#/%=~_|]$")

// ResolveURL resolves the url adding the http scheme, preferring https over http.
func ResolveURL(rawURL string) (string, error) {
	match := urlRegex.FindStringSubmatch(rawURL)

	// Check if rawURL is empty.
	if len(match) == 0 {
		return "", URLCannotBeResolvedError("invalid url")
	}

	// If scheme is provided, attempt to execute a request.
	scheme := match[1]
	if scheme == "https://" || scheme == "http://" {
		if _, err := http.Get(rawURL); err != nil {
			return "", URLCannotBeResolvedError("server not available")
		}

		return rawURL, nil
	}

	if scheme != "" {
		return "", URLCannotBeResolvedError("invalid scheme")
	}

	// Resolve using https.
	if url, err := ResolveURL("https://" + rawURL); err == nil {
		return url, nil
	}

	// Resolve using http.
	if url, err := ResolveURL("http://" + rawURL); err == nil {
		return url, nil
	}

	return "", URLCannotBeResolvedError("cannot find a matching scheme")
}

func RegisterResponder(url string, body []byte, header http.Header) {
	println(len(body), url)

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
