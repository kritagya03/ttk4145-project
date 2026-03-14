package network

import (
	"Network-go/network/bcast"
	"fmt"
	"time"

	. "github.com/kritagya03/ttk4145-project/internal/models"
)

type typeTaggedJSON struct {
	Type    string
	Payload []byte
}

type WorldviewType interface {
	MasterWorldview | SlaveWorldview
}

func isValidNetworkID(networkID int, elevatorCount int) bool {
	return networkID > 0 && networkID <= elevatorCount
}

// TODO: This is reused from master.go
func resetTimer(timer *time.Timer) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(HeartbeatTimeout)
	// fmt.Println("network_server.go resetTimer - Timer has been reset.")
}

// TODO: Add NewSlaveConnection

func Server(masterNetworkCommands <-chan MasterWorldview,
	slaveNetworkCommands <-chan SlaveWorldview,
	masterNetworkEvents chan<- interface{},
	slaveNetworkEvents chan<- MasterWorldview,
	networkPort int,
	networkID int,
	elevatorCount int) {

	// TODO: also used in main.go
	const channelBufferSize int = 16

	// TODO: change terrible variable names
	masterWorldivewTransmit := make(chan MasterWorldview, channelBufferSize)
	masterWorldviewReceive := make(chan MasterWorldview, channelBufferSize)
	slaveWorldviewTransmit := make(chan SlaveWorldview, channelBufferSize)
	slaveWorldviewReceive := make(chan SlaveWorldview, channelBufferSize)

	go bcast.Transmitter(networkPort, masterWorldivewTransmit, slaveWorldviewTransmit)
	go bcast.Receiver(networkPort, masterWorldviewReceive, slaveWorldviewReceive)

	masterHeartbeatTimer := time.NewTimer(HeartbeatTimeout)
	masterHeartbeatTimer.Stop()
	masterIsTimedOut := true

	slaveHeartbeatTimers := make([]*time.Timer, elevatorCount)
	slaveIsTimedOutList := make([]bool, elevatorCount)
	for i := range elevatorCount {
		slaveHeartbeatTimers[i] = time.NewTimer(HeartbeatTimeout)
		slaveHeartbeatTimers[i].Stop()
		slaveIsTimedOutList[i] = true
	}

	slaveTimeouts := make(chan int, elevatorCount)

	for slaveIndex := range elevatorCount {
		go func(i int, timer *time.Timer) {
			for {
				<-timer.C
				slaveTimeouts <- slaveIndex + 1 // TODO: maybe implement NetworkID type?
			}
		}(slaveIndex, slaveHeartbeatTimers[slaveIndex])
	}

	for {
		select {

		case worldview := <-masterWorldviewReceive:
			if masterIsTimedOut && worldview.NetworkID != networkID {
				fmt.Println("network_server.go case masterWorldviewReceive. Received MasterWorldview while master was previously timed out, sending MasterWorldview to masterNetworkEvents channel.")
				masterNetworkEvents <- NewMasterConnection(0)
				masterIsTimedOut = false
			}
			// fmt.Printf("network_server.go case masterWorldviewReceive. Sending MasterWorldview to slaveNetworkEvents channel. worldview = %v\n", worldview)
			masterNetworkEvents <- worldview // Combining multiple masters after network partition
			slaveNetworkEvents <- worldview
			resetTimer(masterHeartbeatTimer)

		case worldview := <-slaveWorldviewReceive:
			if !isValidNetworkID(worldview.NetworkID, elevatorCount) {
				fmt.Printf("network_server.go case slaveWorldviewReceive. Received SlaveWorldview with invalid NetworkID: %d\n", worldview.NetworkID)
				continue
			}
			if slaveIsTimedOutList[worldview.NetworkID-1] {
				fmt.Printf("network_server.go case slaveWorldviewReceive. Received SlaveWorldview from NetworkID %d while slave was previously timed out, sending NewSlaveConnection to masterNetworkEvents channel to trigger appropriate handling.\n", worldview.NetworkID)
				masterNetworkEvents <- NewSlaveConnection{NetworkID: worldview.NetworkID}
				// masterNetworkEvents <- worldview // TODO: should the SlaveWorldview be sent to masterNetworkEvents?
				slaveIsTimedOutList[worldview.NetworkID-1] = false
			}
			// fmt.Printf("network_server.go case slaveWorldviewReceive. Sending SlaveWorldview to masterNetworkEvents channel. worldview = %v\n", worldview)
			masterNetworkEvents <- worldview
			resetTimer(slaveHeartbeatTimers[worldview.NetworkID-1])
			// fmt.Printf("network_server.go case slaveWorldviewReceive. Reset slave heartbeat timer for NetworkID %d\n", worldview.NetworkID)
		case masterCommand := <-masterNetworkCommands:
			// fmt.Println("network_server.go case masterNetworkCommands.")
			masterWorldivewTransmit <- masterCommand
			slaveNetworkEvents <- masterCommand
		case slaveCommand := <-slaveNetworkCommands:
			// fmt.Println("network_server.go case slaveNetworkCommands.")
			slaveWorldviewTransmit <- slaveCommand
			masterNetworkEvents <- slaveCommand
		case <-masterHeartbeatTimer.C:
			fmt.Println("network_server.go case masterHeartbeatTimer.C. Master heartbeat timeout.")
			masterIsTimedOut = true
			masterNetworkEvents <- MasterTimeout(0)
			// Take appropriate action for master timeout
		case slaveNetworkID := <-slaveTimeouts:
			fmt.Printf("network_server.go case slaveTimeouts. Slave %d heartbeat timeout.\n", slaveNetworkID)
			if slaveNetworkID != networkID {
				masterNetworkEvents <- SlaveTimeout{NetworkID: slaveNetworkID}
				slaveIsTimedOutList[slaveNetworkID-1] = true
			}
		}
	}
}
