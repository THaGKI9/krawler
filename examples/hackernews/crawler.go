package main

import (
	"github.com/thagki9/krawler"
)

func main() {
	engine := krawler.GetEngine()
	engine.Initialize("./krawler.yaml")
	engine.InstallQueue(krawler.NewLocalQueue())
	engine.InstallProcessor(RSSFeedParser, "hackernews")
	engine.AddTask(&krawler.Task{
		URL:              "https://news.ycombinator.com/rss",
		Method:           "GET",
		ProcessorName:    "hackernews",
		AllowDuplication: true,
	})
	engine.Start()
}
