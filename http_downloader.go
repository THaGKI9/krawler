package krawler

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// ErrDownloadTimeout indicates the download failed because of timeout
var ErrDownloadTimeout = errors.New("Download timeout")

// ErrDownloaderShuttingDown indicates the downloader is currently shutting down
// and no new task is allow to be scheduled
var ErrDownloaderShuttingDown = errors.New("The downloaded is currently shutting down")

// HTTPDownloader implements a simple http downloader
type HTTPDownloader struct {
	// Request timeout in seconds, default: 5
	Timeout time.Duration

	concurrency  int
	running      chan int
	start        bool
	shuttingDown bool
}

// NewHTTPDownloader returns a HTTP Downloader objects
func NewHTTPDownloader() *HTTPDownloader {
	d := &HTTPDownloader{
		Timeout: 5 * time.Second,
	}

	d.SetConcurrency(5)
	return d
}

// SetConcurrency sets concurrency of the downloader before the download starts
func (d *HTTPDownloader) SetConcurrency(newConcurrency int) error {
	if d.start || len(d.running) > 0 {
		return errors.New("Cannot set up after download started")
	}

	d.running = make(chan int, newConcurrency)
	d.concurrency = newConcurrency
	return nil
}

func (d *HTTPDownloader) startTask() {
	d.running <- 1
}

func (d *HTTPDownloader) finishTask() {
	<-d.running
}

func (d *HTTPDownloader) doDownload(task *Task, doDownloadResultChannel chan *DownloadResult) {
	request, err := http.NewRequest(task.Method, task.URL, task.Body)
	if err != nil {
		doDownloadResultChannel <- &DownloadResult{
			Err:  fmt.Errorf("Create request instance failed, %v", err),
			Task: task,
		}
		return
	}
	request.Header = task.Headers

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		doDownloadResultChannel <- &DownloadResult{
			Err:  fmt.Errorf("Request failed, %v", err),
			Task: task,
		}
		return
	}

	doDownloadResultChannel <- &DownloadResult{
		Err:     nil,
		Task:    task,
		Content: response.Body,
		Headers: response.Header,
	}
}

func (d *HTTPDownloader) handleDownloadResult(task *Task, doDownloadResultChannel chan *DownloadResult, resultChannel chan *DownloadResult) {
	defer d.finishTask()

	select {
	case result := <-doDownloadResultChannel:
		resultChannel <- result
	case <-time.After(d.Timeout):
		resultChannel <- &DownloadResult{Task: task, Err: ErrDownloadTimeout}
	}
}

// Download read information from task and download content respectly
func (d *HTTPDownloader) Download(task *Task, resultChannel chan *DownloadResult) {
	if d.shuttingDown {
		resultChannel <- &DownloadResult{Task: task, Err: ErrDownloaderShuttingDown}
		return
	}
	d.start = true

	d.startTask()
	ch := make(chan *DownloadResult)
	go d.doDownload(task, ch)
	go d.handleDownloadResult(task, ch, resultChannel)
}

// Stop waits for workers to stop and return
func (d *HTTPDownloader) Stop() {
	d.shuttingDown = true
	for len(d.running) > 0 {
		time.Sleep(1 * time.Second)
	}
}
