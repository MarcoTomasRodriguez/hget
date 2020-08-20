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

// TaskFilename represents the state of a download. This file will be located
// in $HOME/ProgramFolder/Download
var TaskFilename = "task.json"

// DisplayProgressBar enables/disables the display of the progress bar.
var DisplayProgressBar = isatty.IsTerminal(os.Stdout.Fd())

// UseHashLength sets the length of the hash used to prevent collisions.
// Note that this can never be more than 32
var UseHashLength = 16

// SaveWithHash enables/disables the collision protection using a hash
// while moving the file from inside the program to outside.
var SaveWithHash = false

// CopyNBytes sets the bytes to copy in a row from the response body.
var CopyNBytes = int64(250)