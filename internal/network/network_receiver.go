package network

import (
	"fmt"
	"net"
	. "github.com/kritagya03/ttk4145-project/internal/models"
)

func Receiver(broadcastEvents chan<- []byte, receivingPort int) {
	receivingAddress := fmt.Sprintf(":%d", receivingPort)
	udpAddress, errorResolve := net.ResolveUDPAddr("udp4", receivingAddress)
	if errorResolve != nil {
		panic(fmt.Sprintf("Error resolving UDP Address: %v", errorResolve))
	}

	receiveConnection, errorListen := net.ListenUDP("udp4", udpAddress)
	if errorListen != nil {
		panic(fmt.Sprintf("Error listening: %v", errorListen))
	}
	defer receiveConnection.Close()

	packetBuffer := make([]byte, NetworkBufferSize)

	for {
		packetByteCount, receivedAddress, readError := receiveConnection.ReadFromUDP(packetBuffer)
		if readError != nil {
			fmt.Println("Error reading:", readError)
			continue
		}
		fmt.Printf("Received message from %s: %s\n", receivedAddress.String(), string(packetBuffer[:packetByteCount]))
		broadcastEvents <- packetBuffer[:packetByteCount]
	}
}
