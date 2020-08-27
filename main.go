package main

import (
	"flag"
	"fmt"
	"github.com/MarcoTomasRodriguez/hget/config"
	"github.com/MarcoTomasRodriguez/hget/download"
	"github.com/MarcoTomasRodriguez/hget/logger"
	"github.com/MarcoTomasRodriguez/hget/utils"
	"os"
	"path/filepath"
	"runtime"
)

func printUsage() {
	logger.Info(`Usage:
| hget [-n Threads] [URL]
| hget list
| hget resume [Task | URL]
| hget clear
| hget remove [Task | URL]
`)
	os.Exit(2)
}

func listCommand() {
	tasks, err := download.GetAllTasks()
	utils.FatalCheck(err)

	logger.Info("Saved tasks:\n")
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

	taskName, err := download.FindTask(args[1])
	utils.FatalCheck(err)

	task, err := download.ReadTask(taskName)
	utils.FatalCheck(err)

	download.Download(task.URL, task, conn)
}

func removeCommand(args []string) {
	if len(args) < 2 {
		logger.Error("TaskName or URL is required.\n")
		printUsage()
		os.Exit(1)
	}

	taskName, err := download.FindTask(args[1])
	utils.FatalCheck(err)

	err = download.RemoveTask(taskName)
	utils.FatalCheck(err)
}

func clearCommand() {
	err := download.RemoveAllTasks()
	utils.FatalCheck(err)
}

func downloadCommand(args []string, conn int) {
	URL, err := utils.ResolveURL(args[0])
	utils.FatalCheck(err)

	if utils.Exists(utils.FolderOf(URL)) {
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
		logger.Error("URL or command is required.\n")
		printUsage()
		os.Exit(1)
	}

	// Load config
	cfg, err := config.LoadConfig(filepath.Join(config.Config.ProgramFolder, config.Config.ConfigFilename))
	if err != nil {
		logger.Error("%v", err)
	}
	config.Config = cfg

	// Execute command
	switch args[0] {
	case "list":
		listCommand()
		break
	case "resume":
		resumeCommand(args, *parallelismPtr)
		break
	case "clear":
		clearCommand()
		break
	case "remove":
		removeCommand(args)
		break
	default:
		downloadCommand(args, *parallelismPtr)
		break
	}
}
