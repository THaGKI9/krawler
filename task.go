package krawler

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Task defines the structure of a task.
type Task struct {
	// URL is the target URL of this task.
	URL string

	// Method is referred to HTTP method used to do the request.
	Method string

	// Headers defines the HTTP request headers.
	Headers http.Header

	// Cookies is a slice of cookies that sent along with the request.
	Cookies []*http.Cookie

	// Body is the body of a http request.
	Body []byte

	// ProcessorName defines which processor will be used to process this task.
	ProcessorName string

	// AllowDuplication indicates queue not to check duplication of this task.
	AllowDuplication bool

	// DontRetryIfProcessorFails indicates whether the task should be tried if processor fails to process it.
	// Retry invoked by processor also count into retry times.
	// If true, task would not be retried if processor failed.
	DontRetryIfProcessorFails bool

	// Meta is the meta information of a task.
	Meta Meta
}

// HashCode returns a unique identity to the task.
func (t *Task) HashCode() string {
	return fmt.Sprintf("%s|%s|%s", t.Method, t.URL, t.ProcessorName)
}

// Name returns name of the task
func (t *Task) Name() string {
	return fmt.Sprintf("%s %s", strings.ToUpper(t.Method), t.URL)
}

// Meta defines a struct that records meta information about a task.
type Meta struct {
	EnqueueTime        time.Time
	DownloadStartTime  time.Time
	DownloadFinishTime time.Time
	RetryTimes         int
	Retried            bool
}
