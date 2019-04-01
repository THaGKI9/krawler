package krawler

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// HTTPDownloader implements a simple http downloader
type HTTPDownloader struct {
	userAgent      string
	timeout        time.Duration
	followRedirect bool
	concurrency    int
	running        chan int
	shuttingDown   bool
}

// NewHTTPDownloader returns a HTTP Downloader objects
func NewHTTPDownloader(config *Config) *HTTPDownloader {
	d := new(HTTPDownloader)
	d.timeout = config.Request.Timeout
	d.userAgent = config.Request.UserAgent
	d.followRedirect = config.Request.FollowRedirect
	d.setConcurrency(config.Request.Concurrency)
	return d
}

func (d *HTTPDownloader) setConcurrency(newConcurrency int) {
	d.running = make(chan int, newConcurrency)
	d.concurrency = newConcurrency
}

func (d *HTTPDownloader) startTask() {
	d.running <- 1
}

func (d *HTTPDownloader) finishTask() {
	<-d.running
}

func (d *HTTPDownloader) doDownload(task *Task, chDoResult chan *DownloadResult) {
	task.Meta.DownloadStartTime = time.Now()
	task.Meta.DownloadFinishTime = time.Time{}
	defer func(finishTime *time.Time) {
		*finishTime = time.Now()
	}(&task.Meta.DownloadFinishTime)

	result := &DownloadResult{
		Err:  nil,
		Task: task,
	}

	var body io.Reader
	if task.Body != nil {
		body = bytes.NewReader(task.Body)
	}
	request, err := http.NewRequest(task.Method, task.URL, body)
	if err != nil {
		result.Err = fmt.Errorf("create request instance failed, reason: %v", err)
		chDoResult <- result
		return
	}
	request.Header = make(http.Header)
	request.Header["User-Agent"] = []string{d.userAgent}
	for field, value := range task.Headers {
		request.Header[field] = value
	}
	for _, cookie := range task.Cookies {
		request.AddCookie(cookie)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		result.Err = fmt.Errorf("request failed, reason: %v", err)
		chDoResult <- result
		return
	}

	result.StatusCode = response.StatusCode
	result.Cookies = response.Cookies()
	result.Headers = response.Header
	result.Content, err = ioutil.ReadAll(response.Body)
	if err != nil {
		result.Err = fmt.Errorf("read body failed, reason: %v", err)
	}

	chDoResult <- result
}

// Download read information from task and download content in respect to the task
func (d *HTTPDownloader) Download(task *Task, chResult chan<- *DownloadResult) {
	if d.shuttingDown {
		chResult <- &DownloadResult{Task: task, Err: ErrDownloaderShuttingDown}
		return
	}

	d.startTask()
	ch := make(chan *DownloadResult)
	go d.doDownload(task, ch)

	go func() {
		select {
		case result := <-ch:
			chResult <- result
		case <-time.After(d.timeout):
			chResult <- &DownloadResult{Task: task, Err: ErrDownloadTimeout}
		}
		d.finishTask()
	}()
}

// Shutdown waits for workers to stop and return
func (d *HTTPDownloader) Shutdown() {
	d.shuttingDown = true
	for len(d.running) > 0 {
		time.Sleep(1 * time.Second)
	}
}
