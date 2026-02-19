package main

import (
	"fmt"
	"time"

	"github.com/kritagya03/ttk4145-project/internal/master"
	. "github.com/kritagya03/ttk4145-project/internal/models"
	"github.com/kritagya03/ttk4145-project/internal/network"
	"github.com/kritagya03/ttk4145-project/internal/slave"
)

// func getEnvInt(key string) int {
// 	valueStr := os.Getenv(key)
// 	value, err := strconv.Atoi(valueStr)
// 	if err != nil {
// 		log.Fatalf("Invalid value for %s: %v", key, err)
// 	}
// 	return value
// }

func main() {
	// networkPort := getEnvInt("NETWORK_PORT")
	// floorCount := getEnvInt("FLOOR_COUNT")
	// elevatorCount := getEnvInt("ELEVATOR_COUNT")
	// networkID := getEnvInt("NETWORK_ID")
	// buttonTypeCount := 2 + elevatorCount
	// calls := make([][]tmp.CallState, floorCount)
	// for i := range calls {
	// 	calls[i] = make([]tmp.CallState, buttonTypeCount)
	// }
	// world := MasterWorldview{
	// 	Calls: Calls{
	// 		Calls: calls,
	// 	},
	// }

	// Make sure to check if networkID is valid (between 1 and elevatorCount)

	const networkPort int = 30045
	const floorCount int = 4
	const elevatorCount int = 3
	const networkID int = 1
	const buttonTypeCount int = 2 + elevatorCount // the number 2 is hall down and hall up, elevatorCount because one list of cab calls per elevator

	const channelBufferSize int = 16

	broadcastEvents := make(chan []byte, channelBufferSize) //maybe remove buffer when emptying in server
	// watchdogNetworkCommands := make(chan string, channelBufferSize)
	masterNetworkCommands := make(chan MasterWorldview, channelBufferSize)
	slaveNetworkCommands := make(chan SlaveWorldview, channelBufferSize)
	broadcastCommands := make(chan []byte, channelBufferSize)
	masterNetworkEvents := make(chan interface{}, channelBufferSize)
	slaveNetworkEvents := make(chan MasterWorldview, channelBufferSize)

	go network.Transmitter(broadcastCommands, networkPort)
	go network.Receiver(broadcastEvents, networkPort)
	go network.Server(broadcastEvents,
		masterNetworkCommands,
		slaveNetworkCommands,
		broadcastCommands,
		masterNetworkEvents,
		slaveNetworkEvents,
		elevatorCount)
	go master.Server(masterNetworkEvents, masterNetworkCommands, networkID, floorCount, buttonTypeCount, elevatorCount)
	go slave.Server(slaveNetworkEvents, slaveNetworkCommands)

	fmt.Println("Finished setting up goroutines.")

	// for {
	// 	time.Sleep(time.Second)
	// 	testSendWorldview(networkID, floorCount, buttonTypeCount, elevatorCount, masterNetworkCommands, slaveNetworkCommands)
	// }
	// testSendSlaveWorldview(networkID, floorCount, buttonTypeCount, elevatorCount, masterNetworkCommands, slaveNetworkCommands)
	// testSendMasterWorldview(floorCount, buttonTypeCount, masterNetworkCommands)
	// time.Sleep(2 * time.Second)
	// testSendMasterWorldview(floorCount, buttonTypeCount, masterNetworkCommands)
	time.Sleep(2 * time.Second)
}

func testSendSlaveWorldview(networkID int, floorCount int, buttonTypeCount int, elevatorCount int, masterNetworkCommands chan<- MasterWorldview, slaveNetworkCommands chan<- SlaveWorldview) {
	behaviour := BehaviourIdle
	direction := DirectionStop
	floorLastVisited := 0

	callsSlave := make([][]CallState, floorCount)
	for i := range callsSlave {
		callsSlave[i] = make([]CallState, buttonTypeCount)
	}
	slaveWorld := SlaveWorldview{
		NetworkID:        networkID,
		Behaviour:        behaviour,
		Direction:        direction,
		FloorLastVisited: floorLastVisited,
		Calls:            CallsMatrix{Matrix: callsSlave},
	}
	slaveNetworkCommands <- slaveWorld
	fmt.Println("Sendt to slaveNetworkCommands.")
}

func testSendMasterWorldview(floorCount int, buttonTypeCount int, masterNetworkCommands chan<- MasterWorldview) {
	callsMaster := make([][]CallState, floorCount)
	for i := range callsMaster {
		callsMaster[i] = make([]CallState, buttonTypeCount)
	}
	masterWorld := MasterWorldview{
		Calls: CallsMatrix{Matrix: callsMaster},
	}
	masterNetworkCommands <- masterWorld
	fmt.Println("Sendt to masterNetworkCommands.")
}
