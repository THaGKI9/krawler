package main

import (
	"container/list"
	"encoding/xml"
	"io/ioutil"

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
func RSSFeedParser(downloadResult *krawler.DownloadResult) (*list.List, []*krawler.Task, error) {
	var rss RSS
	content, err := ioutil.ReadAll(downloadResult.Content)
	if err != nil {
		return nil, nil, err
	}

	err = xml.Unmarshal(content, &rss)
	if err != nil {
		return nil, nil, err
	}

	items := list.New()
	for _, item := range rss.Items {
		items.PushBack(item)
	}

	task := &krawler.Task{
		URL:           "https://news.ycombinator.com/rss",
		Method:        "GET",
		ProcessorName: "hackernews",
	}
	return items, []*krawler.Task{task}, nil
}
