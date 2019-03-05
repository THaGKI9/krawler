package krawler

import (
	"os"
	"os/signal"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// Engine is the core of the crawler system.
type Engine struct {
	downloader       Downloader
	logger           *log.Logger
	queue            *Queue
	processors       map[string]FuncProcessor
	shuttingDown     bool
	downloadingCount int
}

// NewEngine creates a engine instance
func NewEngine() *Engine {
	engine := &Engine{
		logger:     log.New(),
		queue:      NewQueue(),
		processors: make(map[string]FuncProcessor),
	}

	return engine
}

// AddProcessor registers processor into engine
func (e *Engine) AddProcessor(processor FuncProcessor, aliases ...string) *Engine {
	for _, alias := range aliases {
		e.logger.Printf("[Info] Added processor with alias `%s`", alias)
		e.processors[alias] = processor
	}
	return e
}

// SetDownloader sets up a downloader for the crawler
func (e *Engine) SetDownloader(downloader Downloader) *Engine {
	e.downloader = downloader
	return e
}

// AddTask adds task to the queue
func (e *Engine) AddTask(tasks ...*Task) *Engine {
	for _, task := range tasks {
		if _, exists := e.processors[task.ProcessorName]; !exists {
			e.logger.Printf("[Warn] Ignore task with processor missing. ProcessName=%s\n", task.ProcessorName)
			continue
		}
		e.queue.Add(task)
	}
	return e
}

func (e *Engine) handleDownloadTask(chResult chan *DownloadResult) {
	defer func() {
		e.downloadingCount--
	}()
	result := <-chResult
	task := result.Task

	processor := e.processors[task.ProcessorName]
	items, tasks, err := processor(result)

	taskName := task.String()
	if err != nil {
		e.logger.Printf("[Error] Cannot process data from %s, error: %v\n", taskName, err)
		return
	}

	e.logger.Printf("[Info] Retrieve %d items from %s\n", items.Len(), taskName)
	if len(tasks) > 0 {
		e.AddTask(tasks...)
	}
}

func (e *Engine) work(complete chan bool) {
	for !e.shuttingDown {
		task := e.queue.Pop()
		if task == nil {
			if e.downloadingCount > 0 {
				// TODO: wait for processor
				e.logger.Println("[Info] There are no tasks in the queue, wait for downloading to stop.")
				time.Sleep(2 * time.Second)
				continue
			} else if e.downloadingCount == 0 {
				break
			}
		}

		e.logger.Println("[Info] Schedule task " + task.String())
		ch := make(chan *DownloadResult)
		e.downloadingCount++
		e.downloader.Download(task, ch)
		go e.handleDownloadTask(ch)
	}

	e.logger.Println("[Info] No new tasks to be scheduled. Crawler stops.")
	complete <- true
}

// Start launchs the crawler
func (e *Engine) Start() {
	if len(e.processors) == 0 {
		e.logger.Fatalln("[Error] No processor has been configure")
	}

	if e.downloader == nil {
		e.logger.Println("[Warn] Downloader has not been set up. HTTPDownloader would be set up as default downloader.")
		e.SetDownloader(NewHTTPDownloader())
	}

	chComplete := make(chan bool)
	go e.work(chComplete)

	chSigInt := make(chan os.Signal)
	signal.Notify(chSigInt, os.Interrupt)

	select {
	case <-chSigInt:
		log.Println("[Info] Receieve Ctrl-C, start to shutdown")
	case <-chComplete:
	}

	signal.Reset(os.Interrupt)
	e.shutdownElegantly()
}

func (e *Engine) shutdownElegantly() {
	e.shuttingDown = true
	wg := sync.WaitGroup{}

	wg.Add(1)

	go func(downloader Downloader) {
		downloader.Stop()
		wg.Done()
	}(e.downloader)

	wg.Wait()
}
