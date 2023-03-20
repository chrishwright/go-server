package connection

import (
	"errors"
	"fmt"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

type TcpConnectionManager struct {
	maxConnections        int
	numCurrentConnections int
	connections           map[uuid.UUID]TcpConnection
	mu                    sync.Mutex
}

func (tcpConnectionManager *TcpConnectionManager) AddConnection(tcpConnection net.Conn) (*TcpConnection, error) {
	tcpConnectionManager.mu.Lock()
	defer tcpConnectionManager.mu.Unlock()

	if tcpConnectionManager.numCurrentConnections >= tcpConnectionManager.maxConnections {
		if !tcpConnectionManager.closeOldestConnection() {
			message := fmt.Sprintf("There are too many connections on the server: %d", tcpConnectionManager.numCurrentConnections)
			return nil, errors.New(message)
		}
	}

	newConnection := createNewTcpConnection(tcpConnection)

	tcpConnectionManager.connections[newConnection.id] = *newConnection

	tcpConnectionManager.numCurrentConnections++

	return newConnection, nil
}

func (tcpConnectionManager *TcpConnectionManager) CloseConnection(connectionWrapper *TcpConnection) error {
	tcpConnectionManager.mu.Lock()
	defer tcpConnectionManager.mu.Unlock()

	connectionId := connectionWrapper.id
	if _, ok := tcpConnectionManager.connections[connectionId]; !ok {
		fmt.Printf("Connection with ID: %s has already been closed and removed.\n", connectionId)
		return nil
	}

	err := connectionWrapper.connection.Close()
	if err != nil {
		fmt.Printf("There was an error closing the connection with ID: %s and error:%s\n", connectionId.String(), err)
	}

	connectionWrapper.connClosed = true
	delete(tcpConnectionManager.connections, connectionId)
	tcpConnectionManager.numCurrentConnections--

	return nil
}

func (tcpConnectionManager *TcpConnectionManager) closeOldestConnection() bool {
	connections := tcpConnectionManager.connections

	connectionIdKeys := make([]uuid.UUID, 0, len(connections))
	for key := range connections {
		connectionIdKeys = append(connectionIdKeys, key)
	}

	sort.SliceStable(connectionIdKeys, func(i, j int) bool {
		firstConnectionCreated := connections[connectionIdKeys[i]].created
		secondConnectionCreated := connections[connectionIdKeys[j]].created
		return firstConnectionCreated.After(secondConnectionCreated)
	})

	oldestConnectionTime := connections[connectionIdKeys[len(connectionIdKeys)-1]].created
	oldestTcpConnectionWrapper := connections[connectionIdKeys[len(connectionIdKeys)-1]]

	tenSecondsAgo := time.Now().Add(-10 * time.Second)
	if oldestConnectionTime.Before(tenSecondsAgo) {
		oldestTcpConnectionWrapper.connection.Close()
		delete(tcpConnectionManager.connections, oldestTcpConnectionWrapper.id)
		tcpConnectionManager.numCurrentConnections--
		return true
	}

	return false
}

func NewTcpConnectionManager(maxConnections int) *TcpConnectionManager {
	return &TcpConnectionManager{
		maxConnections: maxConnections,
		connections:    make(map[uuid.UUID]TcpConnection),
	}
}

type TcpConnection struct {
	connection net.Conn
	id         uuid.UUID
	created    time.Time
	connClosed bool
}

func createNewTcpConnection(connection net.Conn) *TcpConnection {
	return &TcpConnection{
		id:         uuid.New(),
		connection: connection,
		created:    time.Now(),
	}
}

func (connection *TcpConnection) GetConnection() net.Conn {
	return connection.connection
}

func (connection *TcpConnection) Closed() bool {
	return connection.connClosed
}

func (connection *TcpConnection) GetId() string {
	return connection.id.String()
}
