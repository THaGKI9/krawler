package krawler

import (
	"container/list"
	"sync"
)

// LocalQueue holds a series of tasks. Follow FIFO rule.
type LocalQueue struct {
	mutex   *sync.Mutex
	tasks   *list.List
	visited map[string]bool
}

// NewLocalQueue creates a queue which basic storage is a double linked list
func NewLocalQueue() *LocalQueue {
	queue := &LocalQueue{
		tasks:   list.New(),
		visited: make(map[string]bool),
		mutex:   &sync.Mutex{},
	}
	return queue
}

// Transfer the tasks in the list into persisted storage
func (q *LocalQueue) Shutdown() {
}

// Enqueue add a task into the queue
func (q *LocalQueue) Enqueue(task *Task, allowDuplication bool, position EnqueuePosition) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if !allowDuplication && !q.checkDuplication(task) {
		return ErrQueueTaskDuplicated
	}

	switch position {
	case EnqueuePositionHead:
		q.tasks.PushFront(task)
	case EnqueuePositionTail:
		q.tasks.PushBack(task)
	}

	return nil
}

// Pop returns a task in the front most and remove it from the queue
func (q *LocalQueue) Pop() (*Task, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.tasks.Len() == 0 {
		return nil, nil
	}

	return q.tasks.Remove(q.tasks.Front()).(*Task), nil
}

// Len returns the length of the queue
func (q *LocalQueue) Len() (int64, error) {
	return int64(q.tasks.Len()), nil
}

func (q *LocalQueue) checkDuplication(task *Task) bool {
	hashCode := task.HashCode()
	if visited := q.visited[hashCode]; visited {
		return false
	}
	q.visited[hashCode] = true
	return true
}
