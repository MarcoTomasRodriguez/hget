package download

import (
	"github.com/MarcoTomasRodriguez/hget/logger"
	"github.com/MarcoTomasRodriguez/hget/utils"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

// Download downloads the file from the url considering the state using parallelism.
func Download(url string, state *State, conn int, skipTLS bool) {
	var err error

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// set up parallel
	var files = make([]string, 0)
	var parts = make([]Part, 0)
	var isInterrupted = false

	doneChan := make(chan bool, conn)
	fileChan := make(chan string, conn)
	errorChan := make(chan error, 1)
	stateChan := make(chan Part, 1)
	interruptChan := make(chan bool, conn)

	var downloader *HttpDownloader
	if state == nil {
		downloader = NewHttpDownloader(url, conn, skipTLS)
	} else {
		downloader = &HttpDownloader{
			Url:         state.Url,
			FileName:    filepath.Base(state.Url),
			Parallelism: int64(len(state.Parts)),
			Parts:       state.Parts,
			Resumable:   true,
		}
	}

	go downloader.Do(doneChan, fileChan, errorChan, interruptChan, stateChan)

	for {
		select {
		case <-signalChan:
			// send par number of interrupt for each routine
			isInterrupted = true
			for i := 0; i < conn; i++ { interruptChan <- true }
		case file := <-fileChan:
			files = append(files, file)
		case err := <-errorChan:
			logger.Panic(err)
		case part := <-stateChan:
			parts = append(parts, part)
		case <-doneChan:
			if isInterrupted {
				if downloader.Resumable {
					logger.Info("Interrupted, saving state... \n")
					s := &State{Url: url, Parts: parts}
					err := s.Save()
					if err != nil { logger.Info("%v\n", err) }
					return
				} else {
					logger.Warn("Interrupted, but downloading url is not resumable, silently die\n")
					return
				}
			} else {
				err = JoinFile(files, filepath.Base(url))
				utils.FatalCheck(err)
				err = os.RemoveAll(utils.FolderOf(url))
				utils.FatalCheck(err)
				return
			}
		}
	}
}