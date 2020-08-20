package download

import (
	"github.com/MarcoTomasRodriguez/hget/config"
	"github.com/MarcoTomasRodriguez/hget/logger"
	"github.com/MarcoTomasRodriguez/hget/utils"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

// Download downloads the file from the url considering the state of the task using parallelism.
func Download(url string, task *Task, parallelism int, skipTLS bool) {
	var err error

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// set up parallel
	var files = make([]string, 0)
	var parts = make([]Part, 0)
	var isInterrupted = false

	doneChan := make(chan bool, parallelism)
	fileChan := make(chan string, parallelism)
	errorChan := make(chan error, 1)
	taskChan := make(chan Part, 1)
	interruptChan := make(chan bool, parallelism)

	var downloader *HttpDownloader
	if task == nil {
		downloader = NewHttpDownloader(url, int64(parallelism), skipTLS)
	} else {
		downloader = &HttpDownloader{
			Url:         task.Url,
			FileName:    filepath.Base(task.Url),
			Parallelism: int64(len(task.Parts)),
			Parts:       task.Parts,
			Resumable:   true,
		}
	}

	go downloader.Do(doneChan, fileChan, errorChan, interruptChan, taskChan)

	for {
		select {
		case <-signalChan:
			// send par number of interrupt for each routine
			isInterrupted = true
			for i := 0; i < parallelism; i++ { interruptChan <- true }
		case file := <-fileChan:
			files = append(files, file)
		case err := <-errorChan:
			logger.Panic(err)
		case part := <-taskChan:
			parts = append(parts, part)
		case <-doneChan:
			if isInterrupted {
				if downloader.Resumable {
					logger.Info("Interrupted, saving task... \n")
					s := &Task{Url: url, Parts: parts}
					err := s.SaveTask()
					if err != nil { logger.Info("%v\n", err) }
					return
				} else {
					logger.Warn("Interrupted, but downloading url is not resumable, silently die\n")
					return
				}
			} else {
				var outputName string

				if config.SaveWithHash {
					outputName = utils.FilenameWithHash(url)
				} else {
					outputName = utils.FilenameWithoutHash(url)
				}

				err = JoinFile(files, outputName)
				utils.FatalCheck(err)

				err = os.RemoveAll(utils.FolderOf(url))
				utils.FatalCheck(err)

				return
			}
		}
	}
}