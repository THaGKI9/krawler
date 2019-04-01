package main

import (
	"encoding/xml"

	log "github.com/sirupsen/logrus"

	"github.com/thagki9/krawler"
)

// RSS feed
type RSS struct {
	Items []NewsItem `xml:"channel>item"`
}

// NewsItem is a Hackernews news
type NewsItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	PublishDate string `xml:"pubDate"`
	Comment     string `xml:"comments"`
	Description string `xml:"description"`
}

// RSSFeedParser implements Processor#Parse
func RSSFeedParser(downloadResult *krawler.DownloadResult, engine *krawler.Engine) error {
	var rss RSS
	err := xml.Unmarshal(downloadResult.Content, &rss)
	if err != nil {
		return err
	}

	for _, item := range rss.Items {
		log.Infof("Retrieved item %v", item)
	}

	engine.AddTask(&krawler.Task{
		URL:              "https://news.ycombinator.com/rss",
		Method:           "GET",
		ProcessorName:    "hackernews",
		AllowDuplication: true,
	})
	return nil
}
