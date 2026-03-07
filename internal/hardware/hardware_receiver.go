package hardware

import (
	"fmt"
	"net"

	. "github.com/kritagya03/ttk4145-project/internal/models"
)

// ! Version 1
func Receiver(hardwareEvents chan<- []byte, receivingAddress string) {
	tcpAddress, errorResolve := net.ResolveTCPAddr("tcp", receivingAddress)

	if errorResolve != nil {
		panic(fmt.Sprintf("Error resolving TCP Address: %v", errorResolve))
	}

	listener, errorListen := net.ListenTCP("tcp", tcpAddress)
	if errorListen != nil {
		panic(fmt.Sprintf("Error listening: %v", errorListen))
	}
	defer listener.Close()

	for {
		receiveConnection, errorAccept := listener.AcceptTCP()
		if errorAccept != nil {
			fmt.Println("Error accepting connection:", errorAccept)
			continue
		}

		packetBuffer := make([]byte, NetworkBufferSize)

		for {
			packetByteCount, errorRead := receiveConnection.Read(packetBuffer)
			if errorRead != nil {
				fmt.Println("Error reading:", errorRead)
				break
			}
			fmt.Printf("Received message from %s: %s\n", receiveConnection.RemoteAddr().String(), string(packetBuffer[:packetByteCount]))
			hardwareEvents <- packetBuffer[:packetByteCount]
		}

		receiveConnection.Close()
	}

}

// ! Version 2
// func Receiver(hardwareEvents chan<- []byte, receivingAddress string) {
// 	receiveConnection, errorDial := net.Dial("tcp", receivingAddress)
// 	if errorDial != nil {
// 		panic(fmt.Sprintf("Error dialing connection: %v\n", errorDial))
// 	}

// 	packetBuffer := make([]byte, NetworkBufferSize)

// 	for {
// 		packetByteCount, errorRead := receiveConnection.Read(packetBuffer)
// 		if errorRead != nil {
// 			fmt.Println("Error reading:", errorRead)
// 			break
// 		}
// 		fmt.Printf("Received message from %s: %s\n", receiveConnection.RemoteAddr().String(), string(packetBuffer[:packetByteCount]))
// 		hardwareEvents <- packetBuffer[:packetByteCount]
// 	}

// 	defer receiveConnection.Close()

// }
