package main

import (
	"flag"
	"fmt"
	"github.com/MarcoTomasRodriguez/hget/download"
	"github.com/MarcoTomasRodriguez/hget/logger"
	"github.com/MarcoTomasRodriguez/hget/utils"
	"os"
	"runtime"
)

func printUsage() {
	logger.Info(`Usage:
| hget [-n Threads] [URL]
| hget tasks
| hget resume [TaskName | URL]
`)
	os.Exit(2)
}

func tasksCommand() {
	tasks, err := download.GetAllTasks()
	utils.FatalCheck(err)

	logger.Info("Currently on going download:\n")
	for _, task := range tasks {
		fmt.Println(task)
	}
}

func resumeCommand(args []string, conn int) {
	if len(args) < 2 {
		logger.Error("TaskName or URL is required.\n")
		printUsage()
		os.Exit(1)
	}

	taskName := args[1]
	if utils.IsURL(taskName) {
		URL, err := utils.ResolveURL(taskName)
		utils.FatalCheck(err)
		taskName = utils.FilenameWithHash(URL)
	}

	task, err := download.ReadTask(taskName)
	utils.FatalCheck(err)

	download.Download(task.URL, task, conn)
}

func downloadCommand(args []string, conn int) {
	URL, err := utils.ResolveURL(args[0])
	utils.FatalCheck(err)

	if utils.ExistDir(utils.FolderOf(URL)) {
		logger.Warn("Downloading task already exists. Deleting it first.\n")
		err := os.RemoveAll(utils.FolderOf(URL))
		utils.FatalCheck(err)
	}

	download.Download(URL, nil, conn)
}

func main() {
	// Flags
	parallelismPtr := flag.Int("n", runtime.NumCPU(), "number of threads")

	// Help
	flag.Usage = printUsage

	// Parse
	flag.Parse()

	// Args
	args := flag.Args()

	if len(args) < 1 {
		logger.Error("URL is required.\n")
		printUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "tasks":
		tasksCommand()
		break
	case "resume":
		resumeCommand(args, *parallelismPtr)
		break
	default:
		downloadCommand(args, *parallelismPtr)
		break
	}
}
