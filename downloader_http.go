package krawler

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

// ErrDownloadTimeout indicates the download failed because of timeout
var ErrDownloadTimeout = errors.New("Download timeout")

// ErrDownloaderShuttingDown indicates the downloader is currently shutting down
// and no new task is allow to be scheduled
var ErrDownloaderShuttingDown = errors.New("The downloaded is currently shutting down")

// HTTPDownloader implements a simple http downloader
type HTTPDownloader struct {
	logger         *log.Logger
	userAgent      string
	timeout        time.Duration
	maxRetryTimes  int
	followRedirect bool
	concurrency    int
	running        chan int
	start          bool
	shuttingDown   bool
}

// NewHTTPDownloader returns a HTTP Downloader objects
func NewHTTPDownloader(config *Config) *HTTPDownloader {
	d := new(HTTPDownloader)
	d.logger = config.Logger
	d.timeout = config.RequestTimeout
	d.maxRetryTimes = config.RequestMaxRetryTimes
	d.userAgent = config.RequestUserAgent
	d.followRedirect = config.RequestFollowRedirect
	d.SetConcurrency(config.RequestConcurrency)
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
		result.Err = fmt.Errorf("Create request instance failed, %v", err)
		doDownloadResultChannel <- result
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
		result.Err = fmt.Errorf("Request failed, %v", err)
		doDownloadResultChannel <- result
		return
	}

	result.StatusCode = response.StatusCode
	result.Cookies = response.Cookies()
	result.Headers = response.Header
	result.Content, err = ioutil.ReadAll(response.Body)
	if err != nil {
		result.Err = fmt.Errorf("Read body failed, %v", err)
	}

	doDownloadResultChannel <- result
}

func (d *HTTPDownloader) handleDownloadResult(task *Task, doDownloadResultChannel chan *DownloadResult, resultChannel chan *DownloadResult) {
	defer d.finishTask()

	select {
	case result := <-doDownloadResultChannel:
		resultChannel <- result
	case <-time.After(d.timeout):
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