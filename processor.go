package krawler

import "container/list"

// FuncProcessor defines a function that read downloaded content and
// organize structed items and extract new tasks
type FuncProcessor = func(*DownloadResult) (*list.List, []*Task, error)
