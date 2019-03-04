package krawler

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Task defines the structure of a task
type Task struct {
	URL           string
	Method        string
	Headers       http.Header
	Body          io.Reader
	ProcessorName string
}

// HashCode returns a unique identity to the task
func (t *Task) HashCode() string {
	return fmt.Sprintf("%s|%s|%s", t.Method, t.URL, t.ProcessorName)
}

// String returns name of the task
func (t *Task) String() string {
	return fmt.Sprintf("%s %s", strings.ToUpper(t.Method), t.URL)
}
