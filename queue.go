package krawler

import (
	"container/list"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// EnqueuePosition indicates where an item will be inserted
type EnqueuePosition int32

// EnqueuePosition constant definitions
const (
	_ EnqueuePosition = iota
	EnqueuePositionHead
	EnqueuePositionTail
)

// Hashable defines a interface
type Hashable interface {
	HashCode() string
}

// Queue holds all tasks. Follow FIFO rule.
type Queue struct {
	Config *Config

	logger  *log.Logger
	mutex   *sync.Mutex
	tasks   *list.List
	visited map[string]bool
}

// NewQueue creates a task queue
func NewQueue(config *Config) *Queue {
	return &Queue{
		Config: config,

		tasks:   list.New(),
		visited: make(map[string]bool),
		mutex:   &sync.Mutex{},
		logger:  config.Logger,
	}
}

// Enqueue add a item into the queue
func (q *Queue) Enqueue(item Hashable, checkDuplication bool, position EnqueuePosition) bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if checkDuplication && !q.checkDuplication(item) {
		return false
	}

	switch position {
	case EnqueuePositionHead:
		q.tasks.PushFront(item)
	case EnqueuePositionTail:
		q.tasks.PushBack(item)
	}

	return true
}

// AddFront checks the duplication of task and enqueue the task to the front most of the queue
func (q *Queue) AddFront(task *Task) bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if !task.AllowDuplication {
		if !q.checkDuplication(task) {
			q.logger.Infof("Ignore duplicated task, %s %s, ProcessorName=%s", task.Method, task.URL, task.ProcessorName)
			return false
		}
	}

	q.tasks.PushFront(task)
	task.Meta.EnqueueTime = time.Now()

	return true
}

// Pop returns a task in the front most and remove it from the queue
func (q *Queue) Pop() *Task {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.tasks.Len() == 0 {
		return nil
	}

	return q.tasks.Remove(q.tasks.Front()).(*Task)
}

// Len returns the length of the queue
func (q *Queue) Len() int {
	return q.tasks.Len()
}

func (q *Queue) checkDuplication(item Hashable) bool {
	hashCode := item.HashCode()
	if visited := q.visited[hashCode]; visited {
		return false
	}
	q.visited[hashCode] = true
	return true
}
