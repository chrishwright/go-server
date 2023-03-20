package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	connmanagement "example.com/stackserver/connectionmanagement"
	eventhandler "example.com/stackserver/eventhandler"
	stk "example.com/stackserver/stack"
)

const portNumber = 8080

var stack stk.Stack

var popMutex sync.Mutex = sync.Mutex{}
var pushMutex sync.Mutex = sync.Mutex{}

var eventHandler eventhandler.EventHandler
var connectionManager connmanagement.TcpConnectionManager = *connmanagement.NewTcpConnectionManager(100)

func main() {
	fmt.Println("Starting server on port: ", portNumber)

	server, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", portNumber))
	if err != nil {
		fmt.Printf("Error listening on port %d with error: %s", portNumber, err.Error())
		os.Exit(1)
	}
	defer server.Close()

	for {
		connection, err := server.Accept()
		if err != nil {
			fmt.Println("Error accepting the connection request: ", err.Error())
			continue
		}

		connectionWrapper, err := connectionManager.AddConnection(connection)

		if err != nil {
			// Too many connections, sending busy byte.
			fmt.Println(err.Error())
			connection.Write([]byte{0xFF})
			connection.Close()
			continue
		}

		go handleConnection(connectionWrapper)
	}
}

func handleConnection(connectionWrapper *connmanagement.TcpConnection) {
	header := make([]byte, 1)
	connection := connectionWrapper.GetConnection()

	bytesRead, err := connection.Read(header)
	if err != nil || bytesRead <= 0 {
		fmt.Println("Error retrieving header from connection", err.Error())
		closeConnection(connectionWrapper)
		return
	}

	payloadSize := header[0]

	if isPush(payloadSize) {
		fmt.Printf("Received push with connection ID: %s at: %s\n", connectionWrapper.GetId(), time.Now())
		push(int(payloadSize), connectionWrapper)
	} else {
		fmt.Printf("Received pop with connection ID: %s at: %s\n", connectionWrapper.GetId(), time.Now())
		pop(connectionWrapper)
	}

	closeConnection(connectionWrapper)
}

func isPush(payloadSize byte) bool {
	return payloadSize&0x80 == 0
}

func push(payloadSize int, connectionWrapper *connmanagement.TcpConnection) {
	payload, err := readPayload(payloadSize, connectionWrapper)
	if err != nil {
		fmt.Println(err.Error())
		closeConnection(connectionWrapper)
		return
	}

	go connmanagement.PollForClosedConnections(connectionWrapper, &connectionManager)

	if stack.IsFull() {
		eventWaitingChan := make(chan bool, 1)
		eventHandler.WaitForPopEvent(eventWaitingChan, connectionWrapper)
		<-eventWaitingChan
	}

	pushMutex.Lock()
	defer pushMutex.Unlock()

	if connectionWrapper.Closed() {
		return
	}

	stack.PushToStack(payload)

	writePayload([]byte{0}, connectionWrapper)

	if err != nil {
		fmt.Printf("Connection closed with error in push after byte write: %s\n", err)
		closeConnection(connectionWrapper)
		return
	}

	event := createNewEvent(eventhandler.Push, connectionWrapper)
	eventHandler.HandleEvent(event)
}

func pop(connectionWrapper *connmanagement.TcpConnection) {
	go connmanagement.PollForClosedConnections(connectionWrapper, &connectionManager)

	if stack.IsEmpty() {
		eventWaitingChan := make(chan bool, 1)
		eventHandler.WaitForPushEvent(eventWaitingChan, connectionWrapper)
		<-eventWaitingChan
	}

	popMutex.Lock()
	defer popMutex.Unlock()

	if connectionWrapper.Closed() {
		return
	}

	bytesPopped := stack.PopFromStack()

	error := writePayload(bytesPopped, connectionWrapper)

	if error != nil {
		fmt.Println(error.Error())
		closeConnection(connectionWrapper)
		return
	}

	event := createNewEvent(eventhandler.Pop, connectionWrapper)
	eventHandler.HandleEvent(event)
}

func readPayload(payloadSize int, connectionWrapper *connmanagement.TcpConnection) ([]byte, error) {
	connection := connectionWrapper.GetConnection()
	payload := make([]byte, payloadSize)

	reader := bufio.NewReader(connection)
	bytesRead, err := io.ReadFull(reader, payload[:int(payloadSize)])
	if err != nil || bytesRead < payloadSize {
		message := fmt.Sprintf("Connection closed with error in push after payload read: %s\n", err)
		return nil, errors.New(message)
	}
	return payload, nil
}

func writePayload(payload []byte, connectionWrapper *connmanagement.TcpConnection) error {
	connection := connectionWrapper.GetConnection()
	payloadLength := len(payload)

	if len(payload) > 1 {
		bytesWritten, err := connection.Write([]byte{byte(payloadLength)})
		if bytesWritten <= 0 || err != nil {
			message := fmt.Sprintf("There was an error writing the payload size: %s\n", err.Error())
			return errors.New(message)
		}
	}

	bytesWritten, err := connection.Write(payload)

	if bytesWritten <= 0 || err != nil {
		message := fmt.Sprintf("There was an error writing the payload: %s\n", err.Error())
		return errors.New(message)
	}

	return nil
}

func createNewEvent(eventType eventhandler.EventType, connectionWrapper *connmanagement.TcpConnection) eventhandler.Event {
	return eventhandler.Event{
		EventType:         eventType,
		ConnectionWrapper: connectionWrapper,
	}
}

func closeConnection(connection *connmanagement.TcpConnection) {
	connectionManager.CloseConnection(connection)
}
