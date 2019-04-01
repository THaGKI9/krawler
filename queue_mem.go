package krawler

import (
	"container/list"
	"sync"
)

// LocalQueue holds a series of items. Follow FIFO rule.
type LocalQueue struct {
	mutex   *sync.Mutex
	items   *list.List
	visited map[string]bool
}

// NewLocalQueue creates a queue which basic storage is a double linked list
func NewLocalQueue() *LocalQueue {
	queue := &LocalQueue{
		items:   list.New(),
		visited: make(map[string]bool),
		mutex:   &sync.Mutex{},
	}
	return queue
}

// Transfer the items in the list into persisted storage
func (q *LocalQueue) Shutdown() {
}

// Enqueue add a item into the queue
func (q *LocalQueue) Enqueue(item Hashable, allowDuplication bool, position EnqueuePosition) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if !allowDuplication && !q.checkDuplication(item) {
		return ErrQueueItemDuplicated
	}

	switch position {
	case EnqueuePositionHead:
		q.items.PushFront(item)
	case EnqueuePositionTail:
		q.items.PushBack(item)
	}

	return nil
}

// Pop returns a item in the front most and remove it from the queue
func (q *LocalQueue) Pop() (Hashable, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.items.Len() == 0 {
		return nil, nil
	}

	return q.items.Remove(q.items.Front()).(Hashable), nil
}

// Len returns the length of the queue
func (q *LocalQueue) Len() (int64, error) {
	return int64(q.items.Len()), nil
}

func (q *LocalQueue) checkDuplication(item Hashable) bool {
	hashCode := item.HashCode()
	if visited := q.visited[hashCode]; visited {
		return false
	}
	q.visited[hashCode] = true
	return true
}
