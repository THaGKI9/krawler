package krawler

import (
	"fmt"
	"github.com/go-redis/redis"
	json "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
)

// RedisQueue is a queue that store the task in redis. Follow FIFO rule.
type RedisQueue struct {
	id    string
	redis *redis.Client

	redisKeyCounter           string
	redisKeyItemPrefix        string
	redisKeyQueue             string
	redisKeyDuplicationPrefix string
}

// redisScriptPush implements a lua script to push a task into the queue.
// KEYS = counter, queue, taskPrefix
// ARGV = enqueuePosition, task
var redisScriptPush = redis.NewScript(fmt.Sprintf(`
local id = redis.call('INCR', KEYS[1])
redis.call('SET', KEYS[3] .. id, ARGV[1])

if ARGV[1] == '%d' then
	redis.call('LPUSH', ARGV[2])
else if ARGV[1] == '%d' then
	redis.call('RPUSH', ARGV[2])
else
	redis.call('RPUSH', ARGV[2])
end`, EnqueuePositionHead, EnqueuePositionTail))

// redisScriptPop implements a lua script to pop a task from the queue.
// KEYS = queueKey, taskKeyPrefix
var redisScriptPop = redis.NewScript(`
local id = redis.call('LPOP', KEYS[1])
if id == false then
	return nil
end

local taskId = KEYS[2] .. id
local value = redis.call('GET', taskId)
redis.call('DEL', taskId)

return value`)

// NewRedisQueue creates a redis queue
func NewRedisQueue(id string, redisOptions *redis.Options) *RedisQueue {
	queue := &RedisQueue{
		id:    id,
		redis: redis.NewClient(redisOptions),

		redisKeyCounter:           fmt.Sprintf("{krawler:%s}:counter", id),
		redisKeyQueue:             fmt.Sprintf("{krawler:%s}:queue", id),
		redisKeyDuplicationPrefix: fmt.Sprintf("{krawler:%s}:dup:", id),
		redisKeyItemPrefix:        fmt.Sprintf("{krawler:%s}:task:", id),
	}

	return queue
}

// Transfer the tasks in the list into persisted storage
func (q *RedisQueue) Shutdown() {
	err := q.redis.Close()
	if err != nil {
		log.Errorf("Fail to close redis connection, reason: %v", err)
	}
}

// Enqueue add a task into the queue
func (q *RedisQueue) Enqueue(task *Task, allowDuplication bool, position EnqueuePosition) error {
	if !allowDuplication {
		hashCode := task.HashCode()
		key := q.redisKeyDuplicationPrefix + hashCode
		ok, err := q.redis.SetNX(key, q.id, 0).Result()

		if err != nil {
			return fmt.Errorf("failed to check duplication for %s, reason: %v", hashCode, err)
		}

		if !ok {
			return ErrQueueTaskDuplicated
		}
	}

	taskInBytes, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("fail to marshal a task, reason: %v", err)
	}

	_, err = redisScriptPush.Run(q.redis, []string{q.redisKeyCounter, q.redisKeyQueue, q.redisKeyItemPrefix}, position, taskInBytes).Result()
	if err != redis.Nil {
		return fmt.Errorf("fail to enqueue a task, reason: %v", err)
	}

	return nil
}

// Pop returns a task in the front most and remove it from the queue
func (q *RedisQueue) Pop() (*Task, error) {
	rawTask, err := redisScriptPop.Run(q.redis, []string{q.redisKeyQueue, q.redisKeyItemPrefix}).Result()
	if err != redis.Nil {
		return nil, fmt.Errorf("fail to pop a task from the queue, reason: %v", err)
	}

	task := new(Task)
	err = json.Unmarshal([]byte(rawTask.(string)), task)

	return task, nil
}

// Len returns the length of the queue
func (q *RedisQueue) Len() (int64, error) {
	length, err := q.redis.LLen(q.redisKeyQueue).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get length of queue, reason: %v", err)
	}

	return length, nil
}
