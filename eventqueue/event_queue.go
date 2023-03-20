package eventqueue

import (
	"sync"

	connmanagement "example.com/stackserver/connectionmanagement"
)

type QueueEntry struct {
	Channel           chan bool
	ConnectionWrapper *connmanagement.TcpConnection
}

type Queue struct {
	entries []QueueEntry
	mutex   sync.Mutex
}

func (queue *Queue) IsEmpty() bool {
	return len(queue.entries) == 0
}

func (queue *Queue) Enqueue(notificationChan chan bool, connectionWrapper *connmanagement.TcpConnection) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	entry := QueueEntry{
		Channel:           notificationChan,
		ConnectionWrapper: connectionWrapper,
	}
	queue.entries = append(queue.entries, entry)
}

func (queue *Queue) Dequeue() QueueEntry {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	response := queue.entries[0]
	queue.entries = queue.entries[1:]
	return response
}

func (queue *Queue) Peek() QueueEntry {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	return queue.entries[0]
}
