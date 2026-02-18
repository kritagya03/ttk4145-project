package network

import (
	"fmt"
	"net"
)

func Transmitter(broadcastCommands <-chan []byte, destinationPort int) {
	destinationAddress := fmt.Sprintf("255.255.255.255:%d", destinationPort)
	udpAddress, errorResolve := net.ResolveUDPAddr("udp4", destinationAddress)
	if errorResolve != nil {
		panic(fmt.Sprintf("Error resolving UDP Address: %v", errorResolve))
	}

	transmitConnection, errorDial := net.DialUDP("udp4", nil, udpAddress)
	if errorDial != nil {
		panic(fmt.Sprintf("Error dialing: %v", errorDial))
	}
	defer transmitConnection.Close()

	for {
		packetBuffer := <-broadcastCommands
		_, errorSending := transmitConnection.Write(packetBuffer)
		if errorSending != nil {
			panic(fmt.Sprintf("Error sending: %v", errorSending))
		}
	}
}
