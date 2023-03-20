package connection

import (
	"fmt"
	"io"
	"time"
)

func PollForClosedConnections(connectionWrapper *TcpConnection, connectionManager *TcpConnectionManager) {
	connection := connectionWrapper.GetConnection()
	for !connectionWrapper.connClosed {
		oneByte := make([]byte, 1)
		connection.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
		if _, err := connection.Read(oneByte); err != nil && err == io.EOF {
			fmt.Printf("Connection closing inside thread for ID: %s\n", connectionWrapper.GetId())
			connectionManager.CloseConnection(connectionWrapper)
		}
	}
}
