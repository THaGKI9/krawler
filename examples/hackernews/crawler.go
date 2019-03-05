package main

import "github.com/thagki9/krawler"

func main() {
	engine := krawler.NewEngine()
	engine.AddProcessor(RSSFeedParser, "hackernews")
	engine.AddTask(&krawler.Task{
		URL:           "https://news.ycombinator.com/rss",
		Method:        "GET",
		ProcessorName: "hackernews",
	})
	engine.Start()
}
