package krawler

import (
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
)

// Engine is the core of the crawler system.
type Engine struct {
	Config *Config

	downloader       Downloader
	queue            *Queue
	logger           *log.Logger
	processors       map[string]FuncProcessor
	shuttingDown     bool
	downloadingCount *int64
}

// NewEngine creates a engine instance from file configuration
func NewEngine(configPath string) *Engine {
	config, err := LoadConfigFromPath(configPath)
	if err != nil {
		panic(err)
	}

	return NewEngineFromConfig(config)
}

// NewEngineFromConfig creates a engine instance
func NewEngineFromConfig(config *Config) *Engine {
	engine := &Engine{
		queue:            NewQueue(config),
		processors:       make(map[string]FuncProcessor),
		logger:           config.Logger,
		downloadingCount: new(int64),

		Config: config,
	}
	return engine
}

// AddProcessor registers processor into engine
func (e *Engine) AddProcessor(processor FuncProcessor, aliases ...string) {
	for _, alias := range aliases {
		e.logger.Debugf("Added processor with alias `%s`", alias)
		e.processors[alias] = processor
	}
}

// SetDownloader sets up a downloader for the crawler
func (e *Engine) SetDownloader(downloader Downloader) {
	e.downloader = downloader
}

// AddTask adds task to the queue
func (e *Engine) AddTask(tasks ...*Task) {
	for _, task := range tasks {
		taskCopy := *task

		if _, exists := e.processors[task.ProcessorName]; !exists {
			e.logger.Warnf("Ignore task with processor missing. ProcessName=%s", task.ProcessorName)
			continue
		}

		success := e.queue.Enqueue(&taskCopy, !task.AllowDuplication, EnqueuePositionTail)
		if success {
			e.logger.Infof("Ignore duplicated task %s", taskCopy.Name())
		} else {
			task.Meta.EnqueueTime = time.Now()
		}
	}
}

// AddTaskFront adds task to the front of the queue
func (e *Engine) AddTaskFront(tasks ...*Task) {
	for _, task := range tasks {
		taskCopy := *task

		if _, exists := e.processors[task.ProcessorName]; !exists {
			e.logger.Warnf("Ignore task with processor missing. ProcessName=%s", task.ProcessorName)
			continue
		}

		success := e.queue.Enqueue(&taskCopy, !task.AllowDuplication, EnqueuePositionHead)
		if success {
			e.logger.Infof("Ignore duplicated task %s", taskCopy.Name())
		} else {
			task.Meta.EnqueueTime = time.Now()
		}
	}
}

// RetryTask will check if a task exceeds maximum retry times and reschedule it with highest priority
func (e *Engine) RetryTask(task *Task) {
	taskName := task.Name()

	if task.Meta.RetryTimes > 0 {
		e.logger.Errorf("Task %s has already retried for %d times", taskName, task.Meta.RetryTimes)
	}

	if task.Meta.RetryTimes >= e.Config.RequestMaxRetryTimes {
		e.logger.Errorf("Task %s is removed because it has exceeds maximum retry times", task.Name())
		return
	}

	task.Meta.RetryTimes++
	task.Meta.Retried = true
	e.queue.Enqueue(task, false, EnqueuePositionHead)
	e.logger.Debugf("Task %s has been reschedule for retrying", task.Name())
}

// RescheduleTask will put the task in the front of the queue and will not check duplication
func (e *Engine) RescheduleTask(task *Task) {
	e.queue.Enqueue(task, false, EnqueuePositionHead)
	e.logger.Debugf("Task %s has been reschedule for state persisting", task.Name())
}

func (e *Engine) handleDownloadTask(chResult chan *DownloadResult) {
	defer func() {
		atomic.AddInt64(e.downloadingCount, -1)
	}()
	result := <-chResult
	task := result.Task
	taskName := task.Name()

	if result.Err == ErrDownloaderShuttingDown {
		e.RescheduleTask(task)
		return
	} else if result.Err != nil {
		e.logger.Errorf("Download task %s failed because: %v", taskName, result.Err)
		e.RetryTask(task)
		return
	}

	processor := e.processors[task.ProcessorName]
	parseResult, err := processor(result)
	if err != nil {
		e.logger.Errorf("Process task %s failed because: %v", taskName, err)
		if !task.DontRetryIfProcessorFails {
			e.RetryTask(task)
		}
		return
	}

	e.logger.Infof("Retrieve %d items from %s", parseResult.Items.Len(), taskName)
	if len(parseResult.Tasks) > 0 {
		e.AddTask(parseResult.Tasks...)
	}
}

func (e *Engine) runTask(task *Task) {
	defer func() {
		err := recover()
		if err != nil {
			e.logger.Errorf("Recover from panic while running task %s, panic: %s", task.Name(), err)
			e.RetryTask(task)
		}
	}()

	e.logger.Debugf("Run task %s", task.Name())
	ch := make(chan *DownloadResult)
	atomic.AddInt64(e.downloadingCount, 1)
	e.downloader.Download(task, ch)
	go e.handleDownloadTask(ch)
}

func (e *Engine) work(complete chan bool) {
	for !e.shuttingDown {
		// Pick a task
		task := e.queue.Pop()
		if task == nil {
			downloadingCount := atomic.LoadInt64(e.downloadingCount)
			if downloadingCount > 0 {
				// TODO: wait for processor
				e.logger.Debug("There are no new tasks in the queue, wait for downloading to stop")
				time.Sleep(2 * time.Second)
				continue
			} else if downloadingCount == 0 {
				break
			}
		}

		e.runTask(task)
	}

	e.logger.Infof("No new tasks to be run. Crawler stops")
	complete <- true
}

// Start launches the crawler
func (e *Engine) Start() {
	if len(e.processors) == 0 {
		e.logger.Fatal("No processor has been configure")
	}

	if e.downloader == nil {
		e.logger.Warn("Downloader has not been set up. HTTPDownloader would be set up as default downloader")
		e.SetDownloader(NewHTTPDownloader(e.Config))
	}

	chComplete := make(chan bool)
	go e.work(chComplete)

	chSigInt := make(chan os.Signal)
	signal.Notify(chSigInt, os.Interrupt)

	select {
	case <-chSigInt:
		e.logger.Info("Receive Ctrl-C, start to shutdown")
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
