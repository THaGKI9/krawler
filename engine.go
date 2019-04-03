package krawler

import (
	"os"
	"os/signal"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
)

// Engine is the core of the crawler system.
type Engine struct {
	Config *Config

	downloader       Downloader
	queue            Queue
	processors       map[string]FuncProcessor
	shuttingDown     bool
	downloadingCount *int64
}

var defaultEngine *Engine

func GetEngine() *Engine {
	if defaultEngine == nil {
		defaultEngine = &Engine{}
	}

	return defaultEngine
}

// Initialize the engine with given config
func (e *Engine) Initialize(config *Config) {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})

	e.processors = make(map[string]FuncProcessor)
	e.downloadingCount = new(int64)
	e.Config = config
	config.checkConfig()
}

// InstallQueue installs a task queue onto the engine
func (e *Engine) InstallQueue(queue Queue) {
	if e.queue != nil {
		log.Fatal("a queue has already been added!")
	}

	e.queue = queue
}

// InstallProcessor registers processor into engine
func (e *Engine) InstallProcessor(processor FuncProcessor, aliases ...string) {
	for _, alias := range aliases {
		log.Debugf("Added processor with alias `%s`", alias)
		if _, exists := e.processors[alias]; exists {
			log.Fatalf("A processor with alias `%s` has already been added.", alias)
		}
		e.processors[alias] = processor
	}
}

// InstallDownloader sets up a downloader for the crawler
func (e *Engine) InstallDownloader(downloader Downloader) {
	if e.downloader != nil {
		log.Fatal("a downloader has already been added!")
	}

	e.downloader = downloader
}

// AddTask adds task to the queue
func (e *Engine) AddTask(tasks ...*Task) {
	for _, task := range tasks {
		// copy the task so that any manipulation to the task won't affect task in the queue
		taskCopy := *task

		if _, exists := e.processors[task.ProcessorName]; !exists {
			log.Warnf("Ignore task with processor missing. ProcessName=%s", task.ProcessorName)
			continue
		}

		err := e.queue.Enqueue(&taskCopy, task.AllowDuplication, EnqueuePositionTail)
		if err == ErrQueueTaskDuplicated {
			log.Infof("Ignore duplicated task %s", taskCopy.Name())
		} else if err != nil {
			log.Errorf("Fail to add task to queue, reason: %v", err)
		} else {
			task.Meta.EnqueueTime = time.Now()
		}
	}
}

// RetryTask will check if a task exceeds maximum retry times and reschedule it with highest priority
func (e *Engine) RetryTask(task *Task) {
	taskName := task.Name()

	if task.Meta.RetryTimes >= e.Config.Request.MaxRetryTimes {
		log.Errorf("Task %s is removed because it has exceeds maximum retry times", taskName)
		return
	}

	task.Meta.RetryTimes++

	err := e.queue.Enqueue(task, true, EnqueuePositionTail)
	if err != nil {
		log.Errorf("Fail to reschedule a task %s for retrying, reason: %v", taskName, err)
	} else {
		task.Meta.EnqueueTime = time.Now()
		log.Debugf("Task %s has been reschedule for retrying", taskName)
	}
}

// RescheduleTask will put the task in the front of the queue and will not check duplication
func (e *Engine) RescheduleTask(task *Task) {
	err := e.queue.Enqueue(task, true, EnqueuePositionHead)
	if err != nil {
		log.Errorf("Fail to reschedule task %s for state persisting and task may lost! Reason: %v", task.Name(), err)
	} else {
		log.Debugf("Task %s has been reschedule for state persisting", task.Name())
	}
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
		log.Errorf("Download task %s failed, reason: %v", taskName, result.Err)
		e.RetryTask(task)
		return
	}

	processor := e.processors[task.ProcessorName]
	err := processor(result, e)
	if err != nil {
		log.Errorf("Process task %s failed, reason: %v", taskName, err)
		if !task.DontRetryIfProcessorFails {
			e.RetryTask(task)
		}
		return
	}
}

func (e *Engine) runTask(task *Task) {
	defer func() {
		err := recover()
		if err != nil {
			log.Errorf("Recover from panic while running task %s, panic: %s", task.Name(), err)
			e.RetryTask(task)
		}
	}()

	log.Debugf("Run task %s", task.Name())
	ch := make(chan *DownloadResult)
	atomic.AddInt64(e.downloadingCount, 1)
	e.downloader.Download(task, ch)
	go e.handleDownloadTask(ch)
}

func (e *Engine) work(complete chan<- bool) {
	log.Info("Engine starts to work")

	for !e.shuttingDown {
		// Pick a task
		task, err := e.queue.Pop()
		if err != nil {
			// TODO: better fallback policy
			log.Errorf("Fail to retrieve a task from the queue, reason: %v", err)
			task = nil
		}

		if task == nil {
			downloadingCount := atomic.LoadInt64(e.downloadingCount)
			if downloadingCount > 0 {
				// TODO: wait for processor
				log.Debug("There are no new tasks in the queue, wait for downloading to stop")
				time.Sleep(2 * time.Second)
				continue
			} else if downloadingCount == 0 {
				break
			}
		}

		e.runTask(task.(*Task))
	}

	log.Infof("No new tasks to be run. Crawler stops")
	complete <- true
}

// Start launches the crawler
func (e *Engine) Start() {
	if e.downloader == nil {
		log.Fatal("No downloader has been installed")
	}
	log.Debugf("Use downloader: %s", reflect.TypeOf(e.downloader).Elem().Name())

	if e.queue == nil {
		log.Fatal("No queue has been installed")
	}
	log.Debugf("Use queue: %s", reflect.TypeOf(e.queue).Elem().Name())

	if len(e.processors) == 0 {
		log.Fatal("No processor has been installed")
	}

	chComplete := make(chan bool)
	go e.work(chComplete)

	chSigInt := make(chan os.Signal)
	signal.Notify(chSigInt, os.Interrupt)

	select {
	case <-chSigInt:
		log.Info("Receive Ctrl-C, start to shutdown")
	case <-chComplete:
	}

	signal.Reset(os.Interrupt)
	e.shutdownElegantly()
}

func (e *Engine) shutdownElegantly() {
	e.shuttingDown = true
	wg := sync.WaitGroup{}

	wg.Add(2)

	go func() {
		e.downloader.Shutdown()
		wg.Done()
	}()

	go func() {
		e.queue.Shutdown()
		wg.Done()
	}()

	wg.Wait()
}
