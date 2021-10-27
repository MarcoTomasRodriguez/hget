package utils

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/MarcoTomasRodriguez/hget/config"
)

const (
	B = 1 << (iota * 10)
	KB
	MB
	GB
	TB
)

const (
	iecBaseQuantity     = 1024
	filenameLengthLimit = 255
)

// ConsoleCancelableContext ...
func ConsoleCancelableContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		select {
		case <-sig:
			fmt.Println("sig")
			cancel()
		case <-ctx.Done():
		}
	}()

	return ctx
}

// HashFilename generates a hashed filename using the first N bytes of the rawURL checksum and the filename.
func HashFilename(url string, filename string) string {
	// Obtain the first N bytes from the rawURL checksum.
	urlChecksum := fmt.Sprintf("%x", sha256.Sum256([]byte(url)))[:config.Config.Download.UrlChecksumLength]

	// Concatenate checksum and basename.
	return urlChecksum + "-" + filename
}

// CheckFilenameValidity checks if a filename is valid. If it is not, returns an error.
func CheckFilenameValidity(filename string) error {
	// Check if filename is empty (they might be malicious).
	if filename == "" || filename == "." || filename == ".." || filename == "/" {
		return ErrFilenameEmpty
	}

	// Check filename length.
	if len(filename) > filenameLengthLimit {
		return ErrFilenameTooLong
	}

	return nil
}

// ResolveURL resolves the rawURL adding the http prefix, preferring https over http.
func ResolveURL(rawURL string) (string, error) {
	// Parse the raw rawURL.
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// Check if rawURL is empty.
	if parsedURL.String() == "" {
		return "", errors.New("rawURL is empty")
	}

	// Check if a scheme is provided.
	if parsedURL.Scheme == "https" || parsedURL.Scheme == "http" {
		return parsedURL.String(), nil
	}

	// Resolve using https.
	parsedURL.Scheme = "https"
	if res, err := http.Get(parsedURL.String()); err == nil && res != nil {
		return parsedURL.String(), nil
	}

	// Resolve using http.
	parsedURL.Scheme = "http"
	if res, err := http.Get(parsedURL.String()); err == nil && res != nil {
		return parsedURL.String(), nil
	}

	return "", errors.New("cannot resolve raw url using https or http")
}

// ReadableMemorySize returns a prettier form of some memory size expressed in bytes.
// Note: Do not exceed the float64 limit, as the result will overflow.
func ReadableMemorySize(bytes uint64) string {
	if bytes < iecBaseQuantity {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := uint64(iecBaseQuantity), 0
	for n := bytes / iecBaseQuantity; n >= iecBaseQuantity; n /= iecBaseQuantity {
		div *= iecBaseQuantity
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
