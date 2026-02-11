package main

import (
	"fmt"
	"time"

	"github.com/kritagya03/ttk4145-project/internal/network"
)

func main() {

	networkPort := 30045

	broadcastCommands := make(chan []byte, 16)
	broadcastEvents := make(chan []byte, 16) //maybe remove buffer when emptying in server

	go network.PacketSender(broadcastCommands, networkPort)
	go network.PacketReceiver(broadcastEvents, networkPort)

	fmt.Println("Finished setting up goroutines.")

	for {
		time.Sleep(time.Second)
		broadcastCommands <- []byte("Kritt Er KuL")
		fmt.Println("Sendt to broadcastCommands.")
	}
}
