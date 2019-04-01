package krawler

// FuncProcessor defines a function that read downloaded content and
// extract new tasks.
type FuncProcessor = func(*DownloadResult, *Engine) error
