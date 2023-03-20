package eventhandler

import (
	connmanagement "example.com/stackserver/connectionmanagement"
	"example.com/stackserver/eventqueue"
)

type EventType int

const (
	Push EventType = iota
	Pop
)

var pushQueue eventqueue.Queue
var popQueue eventqueue.Queue

type EventHandler struct{}

type Event struct {
	EventType         EventType
	ConnectionWrapper *connmanagement.TcpConnection
}

func (eventHandler *EventHandler) WaitForPopEvent(notificationChan chan bool, connectionWrapper *connmanagement.TcpConnection) {
	pushQueue.Enqueue(notificationChan, connectionWrapper)
	popQueue.Enqueue(notificationChan, connectionWrapper)
}

func (eventHandler *EventHandler) WaitForPushEvent(notificationChan chan bool, connectionWrapper *connmanagement.TcpConnection) {
	pushQueue.Enqueue(notificationChan, connectionWrapper)
	popQueue.Enqueue(notificationChan, connectionWrapper)
}

func (eventHandler *EventHandler) HandleEvent(event Event) {
	switch event.EventType {
	case Push:
		if !popQueue.IsEmpty() {
			queueEntry := popQueue.Dequeue()
			for queueEntry.ConnectionWrapper.Closed() && !popQueue.IsEmpty() {
				queueEntry = popQueue.Dequeue()
			}

			if !queueEntry.ConnectionWrapper.Closed() {
				queueEntry.Channel <- true
			}
		}
	case Pop:
		if !pushQueue.IsEmpty() {
			queueEntry := pushQueue.Dequeue()
			for queueEntry.ConnectionWrapper.Closed() && !pushQueue.IsEmpty() {
				queueEntry = pushQueue.Dequeue()
			}

			if !queueEntry.ConnectionWrapper.Closed() {
				queueEntry.Channel <- true
			}
		}
	}
}
