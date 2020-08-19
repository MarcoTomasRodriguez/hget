package utils

import (
	"errors"
	"fmt"
	"github.com/MarcoTomasRodriguez/hget/config"
	"github.com/MarcoTomasRodriguez/hget/logger"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// FatalCheck prints & panics if there's an error.
func FatalCheck(err error) {
	if err != nil { logger.Panic(err) }
}

// StringifyIpsV4 converts all the ipv4 ips to string.
func StringifyIpsV4(ips []net.IP) []string {
	var ret = make([]string, 0)
	for _, ip := range ips {
		if ip.To4() != nil {
			ret = append(ret, ip.String())
		}
	}
	return ret
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

// FolderOf gets the folder of a download safely.
func FolderOf(url string) string {
	safePath := filepath.Join(os.Getenv("HOME"), config.ProgramFolder)
	fullQualifyPath, err := filepath.Abs(filepath.Join(os.Getenv("HOME"), config.ProgramFolder, filepath.Base(url)))
	FatalCheck(err)

	// must ensure full qualify path is CHILD of safe path
	// to prevent directory traversal attack
	// using Rel function to get relative between parent and child
	// if relative join base == child, then child path MUST BE real child
	relative, err := filepath.Rel(safePath, fullQualifyPath)
	FatalCheck(err)

	if strings.Contains(relative, "..") {
		FatalCheck(errors.New("you may be a victim of directory traversal path attack\n"))
		return "" // redundant but needed for the compiler to work
	} else {
		return fullQualifyPath
	}
}

// IsUrl checks whether a url is valid or not.
func IsUrl(s string) bool {
	_, err := url.Parse(s)
	return err == nil
}

// ReadableMemorySize returns a prettier form of some memory size.
func ReadableMemorySize(bytes int64) string {
	megabytes := float64(bytes) / (1024 * 1024)
	if megabytes < 1024 {
		return fmt.Sprintf("%.1f MB", megabytes)
	} else {
		return fmt.Sprintf("%.1f GB", megabytes)
	}
}