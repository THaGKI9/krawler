package krawler

// Downloader define a downloader
type Downloader interface {
	Download(task *Task, resultChannel chan *DownloadResult)
	Stop()
}
