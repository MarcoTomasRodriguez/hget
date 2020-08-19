package download

import (
	"github.com/MarcoTomasRodriguez/hget/config"
	"github.com/MarcoTomasRodriguez/hget/logger"
	"github.com/fatih/color"
	"gopkg.in/cheggaaa/pb.v1"
	"io"
	"os"
	"sort"
)

// copyFile copies the content from a src file to an already opened destFile
func copyFile(src string, destFile *os.File) error {
	srcFile, err := os.OpenFile(src, os.O_RDONLY, 0600)
	if err != nil { return err }
	defer srcFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}

// JoinFile joins all the parts of a file and saves them in the output path.
func JoinFile(files []string, out string) error {
	var bar *pb.ProgressBar

	// sort with file name or we will join files with wrong order
	sort.Strings(files)

	logger.Info("Start joining \n")
	if config.DisplayProgressBar {
		bar = pb.StartNew(len(files)).Prefix(color.CyanString("Joining"))
	}

	outFile, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil { return err }
	defer outFile.Close()

	for _, file := range files {
		if err = copyFile(file, outFile); err != nil { return err }

		if bar != nil { bar.Increment() }
	}
	if bar != nil { bar.Finish() }

	return nil
}