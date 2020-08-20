package main

import (
	"flag"
	"fmt"
	"github.com/MarcoTomasRodriguez/hget/download"
	"github.com/MarcoTomasRodriguez/hget/logger"
	"github.com/MarcoTomasRodriguez/hget/utils"
	"os"
	"path/filepath"
	"runtime"
)

func printUsage() {
	logger.Info(`Usage:
hget [URL] [-n connection] [-skip-tls true]
hget tasks
hget resume [TaskName]
`)
}

func tasksCommand() {
	tasks, err := download.GetAllTasks()
	utils.FatalCheck(err)

	logger.Info("Currently on going download:\n")
	for _, task := range tasks { fmt.Println(task) }
}

func resumeCommand(args []string, conn *int, skipTLS *bool) {
	if len(args) < 2 {
		logger.Error("downloading taskName name is required\n")
		printUsage()
		os.Exit(1)
	}

	var taskName string
	if utils.IsUrl(args[1]) {
		taskName = filepath.Base(args[1])
	} else {
		taskName = args[1]
	}

	task, err := download.ReadTask(taskName)
	utils.FatalCheck(err)

	download.Download(task.Url, task, *conn, *skipTLS)
}

func downloadCommand(args []string, conn *int, skipTLS *bool) {
	url := args[0]

	if utils.ExistDir(utils.FolderOf(url)) {
		logger.Warn("Downloading task already exist, remove first \n")
		err := os.RemoveAll(utils.FolderOf(url))
		utils.FatalCheck(err)
	}

	download.Download(url, nil, *conn, *skipTLS)
}

func main() {
	conn    := flag.Int("n", runtime.NumCPU(), "connection")
	skipTLS := flag.Bool("skip-tls", true, "skip verify certificate for https")

	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		logger.Error("url is required\n")
		printUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "tasks": tasksCommand(); break
	case "resume": resumeCommand(args, conn, skipTLS); break
	default: downloadCommand(args, conn, skipTLS); break
	}
}