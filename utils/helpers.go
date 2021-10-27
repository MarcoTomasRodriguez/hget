package utils

import "errors"

var (
	// ErrDownloadNotExist ...
	ErrDownloadNotExist = errors.New("download does not exist")

	// ErrDownloadBroken ...
	ErrDownloadBroken = errors.New("download is broken")

	// ErrFilenameEmpty ...
	ErrFilenameEmpty = errors.New("filename is empty")

	// ErrFilenameTooLong ...
	ErrFilenameTooLong = errors.New("filename is too long")
)
