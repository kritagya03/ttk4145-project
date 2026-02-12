package main

import (
	"fmt"
	"time"

	. "github.com/kritagya03/ttk4145-project/internal/models"
	"github.com/kritagya03/ttk4145-project/internal/network"
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
	// floorCount := getEnvInt("FLOOR_COUNT")
	// elevatorCount := getEnvInt("ELEVATOR_COUNT")
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

	const floorCount int = 4
	const elevatorCount int = 3
	const buttonTypeCount int = 2 + elevatorCount // the number 2 is hall down and hall up, elevatorCount because one list of cab calls per elevator

	networkPort := 30045

	broadcastEvents := make(chan []byte, 16) //maybe remove buffer when emptying in server
	watchdogNetworkCommands := make(chan string, 16)
	masterNetworkCommands := make(chan MasterWorldview, 16)
	slaveNetworkCommands := make(chan SlaveWorldview, 16)
	broadcastCommands := make(chan []byte, 16)
	masterNetworkEvents := make(chan SlaveWorldview, 16)
	slaveNetworkEvents := make(chan MasterWorldview, 16)

	go network.PacketSender(broadcastCommands, networkPort)
	go network.PacketReceiver(broadcastEvents, networkPort)
	go network.Server(broadcastEvents,
		watchdogNetworkCommands,
		masterNetworkCommands,
		slaveNetworkCommands,
		broadcastCommands,
		masterNetworkEvents,
		slaveNetworkEvents)

	fmt.Println("Finished setting up goroutines.")

	for {
		time.Sleep(time.Second)

		// Test masterNetworkCommands
		// var arrayMaster [4][5]CallState = [4][5]CallState{{0, 0, 0, 0, 0}, {0, CallStateOrder, CallStateNone, CallStateCompleted, 0}, {0, 0, 3, 0, 0}, {0, 0, 0, 0, 0}}
		// callsMaster := Calls{Calls: arrayMaster}
		// m := MasterWorldview{Calls: callsMaster}
		// masterNetworkCommands <- m
		// fmt.Println("Sendt to masterNetworkCommands.")
		

		// Test slaveNetworkCommands
		networkID := 1
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
}
