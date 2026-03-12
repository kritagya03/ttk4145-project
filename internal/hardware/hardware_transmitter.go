package hardware

import (
	"fmt"
	"net"
)

func Transmitter(hardwareCommands <-chan []byte, destinationAddress string) {
	tcpAddress, errorResolve := net.ResolveTCPAddr("tcp", destinationAddress)
	if errorResolve != nil {
		panic(fmt.Sprintf("Error resolving TCP Address: %v", errorResolve))
	}

	transmitConnection, errorDial := net.DialTCP("tcp", nil, tcpAddress)
	if errorDial != nil {
		panic(fmt.Sprintf("Error dialing: %v", errorDial))
	}
	defer transmitConnection.Close()

	for {
		packetBuffer := <-hardwareCommands
		_, errorSending := transmitConnection.Write(packetBuffer)
		if errorSending != nil {
			panic(fmt.Sprintf("Error sending: %v", errorSending))
		}
	}
}