package utils

import (
	"crypto/tls"
	"net/http"
)

// HTTPClient is a custom client with tls insecure skip verify enabled.
// TODO: Find a way to enable tls verify, and thus improve security, while allowing multi-threaded downloads.
var HTTPClient = &http.Client{
	Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
}
