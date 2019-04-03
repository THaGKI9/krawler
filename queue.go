package krawler

import (
	"fmt"
)

// EnqueuePosition indicates where a task will be inserted
type EnqueuePosition int32

// EnqueuePosition constant definitions
const (
	_ EnqueuePosition = iota
	EnqueuePositionHead
	EnqueuePositionTail
)

var (
	// ErrQueueTaskDuplicated indicates that a specific task has already been added to the queue
	ErrQueueTaskDuplicated = fmt.Errorf("task is duplicated")
)

// Queue defines a queue interface
type Queue interface {
	// Shutdown asks the queue to prepare for shutting down. In this stage, the queue
	// should do something for shutting down, like persisting.
	Shutdown()

	// Enqueue adds a task into specific position of the queue and
	// check duplication of the task if asked.
	Enqueue(item *Task, allowDuplication bool, position EnqueuePosition) error

	// Pop removes and returns a task from the front-most of the queue.
	Pop() (*Task, error)

	// Len returns the amount of tasks in the queue.
	Len() (int64, error)
}
