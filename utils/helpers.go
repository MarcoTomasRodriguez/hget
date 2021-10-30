package utils

import "errors"

var (
	// ErrDownloadNotExist is an error thrown when trying to fetch an inexistent download.
	ErrDownloadNotExist = errors.New("download does not exist")

	// ErrDownloadBroken is an error thrown when trying to fetch a broken download.
	ErrDownloadBroken = errors.New("download is broken")

	// ErrFilenameEmpty is an error thrown when a cleaned download filename is empty.
	ErrFilenameEmpty = errors.New("filename is empty")

	// ErrFilenameTooLong is an error thrown when a download filename is too long.
	ErrFilenameTooLong = errors.New("filename is too long")
)
