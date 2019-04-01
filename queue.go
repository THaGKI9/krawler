package krawler

import (
	"fmt"
)

// EnqueuePosition indicates where an item will be inserted
type EnqueuePosition int32

// EnqueuePosition constant definitions
const (
	_ EnqueuePosition = iota
	EnqueuePositionHead
	EnqueuePositionTail
)

var (
	// ErrQueueItemDuplicated indicates that a specific has been added to the queue
	ErrQueueItemDuplicated = fmt.Errorf("item is duplicated")
)

// Hashable defines a interface
type Hashable interface {
	HashCode() string
}

// Queue defines a queue interface
type Queue interface {
	// Shutdown asks the queue to prepare for shutting down. In this stage, the queue
	// should do something for shutting down, like persisting.
	Shutdown()

	// Enqueue adds an item into specific position of the queue and
	// check duplication of the task if asked.
	Enqueue(item Hashable, allowDuplication bool, position EnqueuePosition) error

	// Pop removes and returns an item from the front-most of the queue.
	Pop() (Hashable, error)

	// Len returns the amount of items in the queue.
	Len() (int64, error)
}
