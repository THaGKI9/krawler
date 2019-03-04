package main

import "github.com/thagki9/krawler"

func main() {
	processor := ListFeedPageProcessor{}
	engine := krawler.NewEngine()
	engine.AddProcessor(processor, "hackernews")
	engine.AddTask(&krawler.Task{
		URL:           "https://news.ycombinator.com/rss",
		Method:        "GET",
		ProcessorName: "hackernews",
	})
	engine.Start()
}
