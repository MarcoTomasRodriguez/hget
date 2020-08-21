package download

import (
	"github.com/MarcoTomasRodriguez/hget/config"
	"github.com/MarcoTomasRodriguez/hget/logger"
	"github.com/MarcoTomasRodriguez/hget/utils"
	"github.com/fatih/color"
	"gopkg.in/cheggaaa/pb.v1"
	"io"
	"os"
	"path/filepath"
	"sort"
)

// copyFile copies the content from a src file to an already opened destFile
func copyFile(src string, destFile *os.File) error {
	srcFile, err := os.OpenFile(src, os.O_RDONLY, 0600)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}

// JoinParts joins all the parts of a file and saves them in the output path.
func JoinParts(parts []string, outputPath string) error {
	var bar *pb.ProgressBar

	// sort with file name or we will join parts with wrong order
	sort.Strings(parts)

	if config.DisplayProgressBar {
		bar = pb.StartNew(len(parts)).Prefix(color.CyanString("Joining"))
	}

	outputFile, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	for _, file := range parts {
		if err = copyFile(file, outputFile); err != nil {
			return err
		}

		if bar != nil {
			bar.Increment()
		}
	}

	if bar != nil {
		bar.Finish()
	}

	return nil
}

// CalculateParts calculates & initialize the parts of a download. This parts will be joined later by JoinParts.
func CalculateParts(url string, parallelism int64, length int64) []Part {
	ret := make([]Part, parallelism)
	for current := int64(0); current < parallelism; current++ {
		from := (length / parallelism) * current
		to := length

		if current < parallelism-1 {
			to = (length/parallelism)*(current+1) - int64(1)
		}

		folder := utils.FolderOf(url)
		if err := utils.MkdirIfNotExist(folder); err != nil {
			logger.Info("%v", err)
			os.Exit(1)
		}

		// Padding 0 before path name as filename will be sorted as string
		path := filepath.Join(folder, utils.PartName(current, parallelism))

		ret[current] = Part{Path: path, RangeFrom: from, RangeTo: to}
	}

	return ret
}
