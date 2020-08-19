package config

import (
	"github.com/mattn/go-isatty"
	"os"
)

// Home is the $HOME of the system.
var Home = os.Getenv("HOME")

// ProgramFolder is the folder in which the program will store his information
// about the ongoing downloads. This path is relative to $HOME.
var ProgramFolder = ".hget/"

// StateFilename represents the state of a download. This file will be located
// in $HOME/ProgramFolder/Download
var StateFilename = "state.json"

// DisplayProgressBar enables/disables the display of the progress bar.
var DisplayProgressBar = isatty.IsTerminal(os.Stdout.Fd())