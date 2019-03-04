package krawler

import "container/list"

// Processor define the interface of a processor
type Processor interface {
	// Parse read downloaded content and organize structed items and extract new tasks
	Parse(*DownloadResult) (*list.List, []*Task, error)
}
