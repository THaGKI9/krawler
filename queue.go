package krawler

import (
	"container/list"
	"log"
	"os"
	"sync"
)

// Queue holds all tasks. Follow FIFO rule.
type Queue struct {
	logger  *log.Logger
	mutex   *sync.Mutex
	tasks   *list.List
	visited map[string]bool
}

// NewQueue creates a task queue
func NewQueue() *Queue {
	return &Queue{
		tasks:   list.New(),
		visited: make(map[string]bool),
		mutex:   &sync.Mutex{},
		logger:  log.New(os.Stdout, "[KrawlerQueue]", log.LstdFlags),
	}
}

// Add checks the duplication of task and enqueue the task
func (q *Queue) Add(task *Task) (int, bool) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	hashCode := task.HashCode()
	if visited := q.visited[hashCode]; visited {
		q.logger.Printf("[Info] Ignore duplicated task, %s %s, ProcessorName=%s\n", task.Method, task.URL, task.ProcessorName)
		return q.tasks.Len(), false
	}

	q.visited[hashCode] = true
	q.tasks.PushBack(task)

	return q.tasks.Len(), true
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
