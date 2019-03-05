package krawler

import "net/http"

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
