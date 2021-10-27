package utils

import (
	"crypto/tls"
	"net/http"
)

var HTTPClient = &http.Client{
	Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
}
