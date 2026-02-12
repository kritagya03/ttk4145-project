package network

import (
	"fmt"
	"net"
	"reflect"
)

func PacketSender(broadcastCommands <-chan []byte, destinationPort int) {
	destinationAddress := fmt.Sprintf("255.255.255.255:%d", destinationPort)
	udpAddress, errorResolve := net.ResolveUDPAddr("udp4", destinationAddress)

	fmt.Printf("network_transmitter.go PacketSender. reflect.TypeOf((*[]byte)(nil)).Elem().String()=%v\n", reflect.TypeOf((*[]byte)(nil)).Elem().String())

	fmt.Println("PacketSender entered")

	if errorResolve != nil {
		fmt.Println("Error resolving UDP Address:", errorResolve)
	}

	sendConnection, errorDial := net.DialUDP("udp4", nil, udpAddress)
	if errorDial != nil {
		fmt.Println("Error dialing:", errorDial)
	}
	defer sendConnection.Close()

	for {
		packetBuffer := <-broadcastCommands
		_, errorSending := sendConnection.Write(packetBuffer)
		if errorSending != nil {
			fmt.Println("Error sending:", errorSending)
		}
	}
}
