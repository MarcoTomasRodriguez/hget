package utils

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/MarcoTomasRodriguez/hget/config"
	"github.com/MarcoTomasRodriguez/hget/logger"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	FilenameLengthLimit = 255
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

// RemoveHashFromFilename returns the basename + the hash of the url.
func FilenameWithHash(url string) string {
	base := filepath.Base(url)
	hash := HashOf(url)[:config.UseHashLength]
	if base == "." {
		logger.Panic(errors.New("there is no basename for the url"))
	}

	filename := hash + "-" + base
	if len(filename) > FilenameLengthLimit {
		logger.Panic(fmt.Errorf("the filename length should never exceed the limit of %d",
			FilenameLengthLimit-len(hash)+1))
	}

	return filename
}

// FilenameWithoutHash returns the basename of the url.
func FilenameWithoutHash(url string) string {
	filename := filepath.Base(url)
	if filename == "." {
		logger.Panic(errors.New("there is no basename for the url"))
	}

	if len(filename) > FilenameLengthLimit {
		logger.Panic(fmt.Errorf("the filename length should never exceed the limit of %d", FilenameLengthLimit))
	}

	return filename
}

// FolderOf gets the folder of a download safely.
func FolderOf(url string) string {
	safePath := filepath.Join(config.Home, config.ProgramFolder)
	fullQualifyPath, err := filepath.Abs(filepath.Join(config.Home, config.ProgramFolder, FilenameWithHash(url)))
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
func IsUrl(URL string) bool {
	_, err := url.Parse(URL)
	return err == nil
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

// PartName creates the part name with the part number formatted with as many leading zeros as needed.
// For example, PartName(0, 100) = "part.00", PartName(100, 100) = "part.99" and PartName(101, 101) = "part.100".
func PartName(part int64, parallelism int64) string {
	leadingZeros := int(math.Max(math.Log10(float64(parallelism-1))+1, 1))
	fmt.Println(leadingZeros)
	return fmt.Sprintf("part.%0"+strconv.Itoa(leadingZeros)+"d", part)
}
