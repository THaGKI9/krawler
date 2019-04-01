package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/thagki9/krawler"
)

func main() {
	config := krawler.GetDefaultConfig()
	config.Logger.Level = log.DebugLevel

	engine := krawler.GetEngine()
	engine.Initialize(config)
	engine.InstallQueue(krawler.NewLocalQueue())
	engine.InstallDownloader(krawler.NewHTTPDownloader(config))
	engine.InstallProcessor(RSSFeedParser, "hackernews")
	engine.AddTask(&krawler.Task{
		URL:              "https://news.ycombinator.com/rss",
		Method:           "GET",
		ProcessorName:    "hackernews",
		AllowDuplication: true,
	})
	engine.Start()
}
