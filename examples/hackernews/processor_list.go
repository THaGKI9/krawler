package main

import (
	"container/list"
	"encoding/xml"

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
	Descripts   string `xml:"description"`
}

// RSSFeedParser implements Processor#Parse
func RSSFeedParser(downloadResult *krawler.DownloadResult) (*krawler.ParseResult, error) {
	if downloadResult.Err != nil {
		return nil, downloadResult.Err
	}

	var rss RSS
	err := xml.Unmarshal(downloadResult.Content, &rss)
	if err != nil {
		return nil, err
	}

	items := list.New()
	for _, item := range rss.Items {
		items.PushBack(item)
	}

	task := &krawler.Task{
		URL:              "https://news.ycombinator.com/rss",
		Method:           "GET",
		ProcessorName:    "hackernews",
		AllowDuplication: true,
	}
	return &krawler.ParseResult{Items: items, Tasks: []*krawler.Task{task}}, nil
}
