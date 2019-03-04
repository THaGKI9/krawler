package krawler

import (
	"io"
	"net/http"
)

// DownloadResult defines how download result should be organized
type DownloadResult struct {
	Err     error
	Task    *Task
	Content io.Reader
	Headers http.Header
}
