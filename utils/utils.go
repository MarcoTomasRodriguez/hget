package utils

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/MarcoTomasRodriguez/hget/config"
	"github.com/MarcoTomasRodriguez/hget/logger"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	filenameLengthLimit = 255
	Byte                = 1
	KiloByte            = 1024 * Byte
	MegaByte            = 1024 * KiloByte
	GigaByte            = 1024 * MegaByte
	TeraByte            = 1024 * GigaByte
)

// FatalCheck prints & panics if there's an error.
func FatalCheck(err error) {
	if err != nil {
		logger.Panic(err)
	}
}

// MkdirIfNotExist makes a directory with perm 0700 if not exists.
func MkdirIfNotExist(folder string) error {
	if _, err := os.Stat(folder); err != nil {
		if err = os.MkdirAll(folder, 0700); err != nil {
			return err
		}
	}
	return nil
}

// ExistDir checks whether directory exists or not.
func ExistDir(folder string) bool {
	_, err := os.Stat(folder)
	return err == nil
}

// HashOf returns the sha256 sum256 hash of the parameter.
func HashOf(str string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(str)))
}

// FilenameWithHash returns the hash-basename from the url.
func FilenameWithHash(url string) string {
	base := filepath.Base(url)
	hash := HashOf(url)[:config.Config.UseHashLength]
	if base == "." {
		logger.Panic(errors.New("there is no basename for the url"))
	}

	filename := hash + "-" + base
	if len(filename) > filenameLengthLimit {
		logger.Panic(fmt.Errorf("the filename length should never exceed the limit of %d",
			filenameLengthLimit-len(hash)+1))
	}

	return filename
}

// FilenameWithoutHash returns the basename of the url.
func FilenameWithoutHash(url string) string {
	filename := filepath.Base(url)
	if filename == "." {
		logger.Panic(errors.New("there is no basename for the url"))
	}

	if len(filename) > filenameLengthLimit {
		logger.Panic(fmt.Errorf("the filename length should never exceed the limit of %d", filenameLengthLimit))
	}

	return filename
}

// FolderOf gets the folder of a download safely.
func FolderOf(url string) string {
	safePath := filepath.Join(config.Config.Home, config.Config.ProgramFolder)
	fullQualifyPath, err := filepath.Abs(filepath.Join(config.Config.Home, config.Config.ProgramFolder, FilenameWithHash(url)))
	FatalCheck(err)

	// must ensure full qualify path is CHILD of safe path
	// to prevent directory traversal attack
	// using Rel function to get relative between parent and child
	// if relative join base == child, then child path MUST BE real child
	relative, err := filepath.Rel(safePath, fullQualifyPath)
	FatalCheck(err)

	if strings.Contains(relative, "..") {
		FatalCheck(errors.New("you might be a victim of directory traversal path attack"))
		return "" // redundant but needed for the compiler to work
	}

	return fullQualifyPath
}

// IsURL checks whether a url is valid or not.
func IsURL(URL string) bool {
	_, err := url.Parse(URL)
	return err == nil
}

// ResolveURL resolves the url adding the http scheme going for https-first if needed.
func ResolveURL(URL string) (string, error) {
	// Check if the URL is valid
	if !IsURL(URL) {
		return "", errors.New("the URL provided is not valid")
	}

	// Check if the scheme is provided
	if strings.HasPrefix(URL, "https://") || strings.HasPrefix(URL, "http://") {
		return URL, nil
	}

	// Resolve https
	httpsURL := "https://" + URL
	if res, _ := http.Get(httpsURL); res != nil {
		return httpsURL, nil
	}

	// Resolve http
	httpURL := "http://" + URL
	if res, _ := http.Get(httpURL); res != nil {
		return httpURL, nil
	}

	return "", errors.New("cannot resolve url to HTTPS or HTTP")
}

// ReadableMemorySize returns a prettier form of some memory size.
func ReadableMemorySize(bytes int64) string {
	b := float64(bytes)
	if bytes < MegaByte {
		return fmt.Sprintf("%.1f KB", b/KiloByte)
	} else if bytes < GigaByte {
		return fmt.Sprintf("%.1f MB", b/MegaByte)
	} else if bytes < TeraByte {
		return fmt.Sprintf("%.1f GB", b/GigaByte)
	} else {
		return fmt.Sprintf("%.1f TB", b/TeraByte)
	}
}

// MakePartName creates the part name with the part number formatted with as many leading zeros as needed.
// For example, MakePartName(0, 100) = "part.00", MakePartName(100, 100) = "part.99" and MakePartName(101, 101) = "part.100".
func MakePartName(part int64, parallelism int64) string {
	leadingZeros := int(math.Max(math.Log10(float64(parallelism-1))+1, 1))
	return fmt.Sprintf("part.%0"+strconv.Itoa(leadingZeros)+"d", part)
}
