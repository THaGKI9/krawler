package krawler

import "container/list"

// FuncProcessor defines a function that read downloaded content and
// organize structured items and extract new tasks
type FuncProcessor = func(*DownloadResult) (*ParseResult, error)

// ParseResult defines what a processor should return including items extracting
// from downloaded content and new tasks to schedule
type ParseResult struct {
	Items *list.List
	Tasks []*Task
}
