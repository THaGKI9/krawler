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
	// Download receive a task, perform downloading and send the download result
	// to the channel.
	Download(*Task, chan<- *DownloadResult)

	// Shutdown Indicates the downloader to wait for downloading task and stop receiving
	// new tasks.
	Shutdown()
}

var (
	// ErrDownloadTimeout indicates the download failed because of timeout
	ErrDownloadTimeout = errors.New("download timeout")

	// ErrDownloaderShuttingDown indicates the downloader is currently shutting down
	// and no new task is allow to be scheduled
	ErrDownloaderShuttingDown = errors.New("the downloader is currently shutting down")
)
