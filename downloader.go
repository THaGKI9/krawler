package krawler

import (
	"errors"
	"net/http"
)

// DownloadResult defines how download result should be organized
type DownloadResult struct {
	StatusCode int
	Content    []byte
	Headers    http.Header
	Cookies    []*http.Cookie

	Err  error
	Task *Task
}

// Downloader define a downloader
type Downloader interface {
	Download(task *Task, resultChannel chan *DownloadResult)
	Stop()
}

// ErrDownloadTimeout indicates the download failed because of timeout
var ErrDownloadTimeout = errors.New("Download timeout")

// ErrDownloaderShuttingDown indicates the downloader is currently shutting down
// and no new task is allow to be scheduled
var ErrDownloaderShuttingDown = errors.New("The downloaded is currently shutting down")
