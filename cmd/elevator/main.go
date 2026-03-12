package main

import (
	"Driver-go/elevio"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/kritagya03/ttk4145-project/internal/master"
	. "github.com/kritagya03/ttk4145-project/internal/models"
	"github.com/kritagya03/ttk4145-project/internal/network"
	"github.com/kritagya03/ttk4145-project/internal/slave"
)

func areFlagsValid(networkID int, floorCount int, elevatorCount int, networkPort int, hardwarePort int) bool {
	return networkID >= 1 &&
		floorCount >= 1 &&
		elevatorCount >= 1 &&
		networkPort >= 1024 && networkPort <= 65535 &&
		hardwarePort >= 1024 && hardwarePort <= 65535
}

func main() {
	var networkID, floorCount, elevatorCount, networkPort, hardwarePort int

	flag.IntVar(&networkID, "network-id", -1, "Elevator network ID. Must larger or equal to 1")
	flag.IntVar(&floorCount, "floor-count", 4, "Number of floors")
	flag.IntVar(&elevatorCount, "elevator-count", 3, "Number of elevators")
	flag.IntVar(&networkPort, "network-port", 30045, "Network port")
	flag.IntVar(&hardwarePort, "hardware-port", 15657, "Hardware port")

	flag.Parse()

	if !areFlagsValid(networkID, floorCount, elevatorCount, networkPort, hardwarePort) {
		log.Fatal("All arguments are required and must be valid: -network-id, -floor-count, -elevator-count, -network-port, -hardware-port")
	}

	fmt.Printf("Arguments are networkID=%d, floorCount=%d, elevatorCount=%d, networkPort=%d, hardwarePort=%d\n", networkID, floorCount, elevatorCount, networkPort, hardwarePort)

	buttonTypeCount := 2 + elevatorCount

	// Make sure to check if networkID is valid (between 1 and elevatorCount)

	// const networkPort int = 30045
	// const floorCount int = 4
	// const elevatorCount int = 3
	// const networkID int = 1
	// const buttonTypeCount int = 2 + elevatorCount // the number 2 is hall down and hall up, elevatorCount because one list of cab calls per elevator

	const channelBufferSize int = 16

	// broadcastEvents := make(chan []byte, channelBufferSize) //maybe remove buffer when emptying in server
	// watchdogNetworkCommands := make(chan string, channelBufferSize)
	masterNetworkCommands := make(chan MasterWorldview, channelBufferSize)
	slaveNetworkCommands := make(chan SlaveWorldview, channelBufferSize)
	// broadcastCommands := make(chan []byte, channelBufferSize)
	masterNetworkEvents := make(chan interface{}, channelBufferSize)
	slaveNetworkEvents := make(chan MasterWorldview, channelBufferSize)

	// go network.Transmitter(broadcastCommands, networkPort)
	// go network.Receiver(broadcastEvents, networkPort)
	go network.Server(masterNetworkCommands,
		slaveNetworkCommands,
		masterNetworkEvents,
		slaveNetworkEvents,
		networkPort,
		networkID,
		elevatorCount)
	go master.Server(masterNetworkEvents, masterNetworkCommands, networkID, floorCount, buttonTypeCount, elevatorCount)
	go slave.Server(slaveNetworkEvents, slaveNetworkCommands, networkID, floorCount, buttonTypeCount, hardwarePort)

	fmt.Println("Finished setting up goroutines.")

	// for {
	// 	time.Sleep(time.Second)
	// 	testSendWorldview(networkID, floorCount, buttonTypeCount, elevatorCount, masterNetworkCommands, slaveNetworkCommands)
	// }
	// testSendSlaveWorldview(networkID, floorCount, buttonTypeCount, elevatorCount, masterNetworkCommands, slaveNetworkCommands)
	// // time.Sleep(2 * time.Second)
	// masterWorldview := testGetDefaultMasterWorldview(networkID, floorCount, buttonTypeCount)
	// masterWorldview.Calls.Matrix[3][1] = 1
	// masterWorldview.Calls.Matrix[2][3] = 1
	// masterWorldview.Calls.Matrix[3][4] = 3
	// testSendMasterWorldview(masterWorldview, masterNetworkCommands)
	// masterWorldview.Calls.Matrix[1][1] = CallStateOrder
	// masterWorldview.Calls.Matrix[2][0] = CallStateOrder
	// testSendMasterWorldview(masterWorldview, masterNetworkCommands)

	time.Sleep(time.Hour) // ! TODO: CHANGE TO WAIT FOR SOMETHING
}

func testSendSlaveWorldview(networkID int, floorCount int, buttonTypeCount int, elevatorCount int, masterNetworkCommands chan<- MasterWorldview, slaveNetworkCommands chan<- SlaveWorldview) {
	callsSlave := make([][]CallState, floorCount)
	for i := range callsSlave {
		callsSlave[i] = make([]CallState, buttonTypeCount)
	}
	slaveWorld := SlaveWorldview{
		NetworkID:        networkID,
		Behaviour:        BehaviourIdle,
		Direction:        elevio.MD_Stop,
		FloorLastVisited: 0,
		Calls:            CallsMatrix{Matrix: callsSlave},
	}
	slaveNetworkCommands <- slaveWorld
	fmt.Println("Sendt to slaveNetworkCommands.")
}

func testGetDefaultMasterWorldview(networkID int, floorCount int, buttonTypeCount int) MasterWorldview {
	calls := make([][]CallState, floorCount)
	for i := range calls {
		calls[i] = make([]CallState, buttonTypeCount)
	}
	return MasterWorldview{
		NetworkID: networkID,
		Calls:     CallsMatrix{Matrix: calls},
	}
}

func testSendMasterWorldview(masterWorldview MasterWorldview, masterNetworkCommands chan<- MasterWorldview) {
	masterNetworkCommands <- masterWorldview
	fmt.Println("Sendt to masterNetworkCommands.")
}
