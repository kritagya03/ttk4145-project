package network

import (
	"encoding/json"
	"fmt"
	"reflect"
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
	fmt.Println("network_server.go resetTimer - Timer has been reset.")
}

// TODO: Add NewSlaveConnection

func Server(
	broadcastEvents <-chan []byte,
	masterNetworkCommands <-chan MasterWorldview,
	slaveNetworkCommands <-chan SlaveWorldview,
	broadcastCommands chan<- []byte,
	masterNetworkEvents chan<- interface{},
	slaveNetworkEvents chan<- MasterWorldview,
	networkID int,
	elevatorCount int) {

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
		case broadcastEvent := <-broadcastEvents:
			fmt.Println("network_server.go case broadcastEvents. Received broadcast event, attempting to decode worldview.")
			worldview := packetToWorldview(broadcastEvent)
			if worldview == nil {
				fmt.Println("network_server.go case broadcastEvents. Received invalid packet, skipping.")
				continue
			}
			switch worldview := worldview.(type) {
			case MasterWorldview:
				if masterIsTimedOut && worldview.NetworkID != networkID {
					fmt.Println("network_server.go case broadcastEvents. Received MasterWorldview while master was previously timed out, sending MasterWorldview to masterNetworkEvents channel.")
					masterNetworkEvents <- NewMasterConnection(0)
					masterIsTimedOut = false
				}
				fmt.Printf("network_server.go case broadcastEvents. Sending MasterWorldview to slaveNetworkEvents channel. worldview = %v\n", worldview)
				masterNetworkEvents <- worldview // Combining multiple masters after network partition
				slaveNetworkEvents <- worldview
				resetTimer(masterHeartbeatTimer)
			case SlaveWorldview:
				if !isValidNetworkID(worldview.NetworkID, elevatorCount) {
					fmt.Printf("network_server.go case broadcastEvents. Received SlaveWorldview with invalid NetworkID: %d\n", worldview.NetworkID)
					continue
				}
				if slaveIsTimedOutList[worldview.NetworkID-1] {
					fmt.Printf("network_server.go case broadcastEvents. Received SlaveWorldview from NetworkID %d while slave was previously timed out, sending NewSlaveConnection to masterNetworkEvents channel to trigger appropriate handling.\n", worldview.NetworkID)
					masterNetworkEvents <- NewSlaveConnection{NetworkID: worldview.NetworkID}
					// masterNetworkEvents <- worldview // TODO: should the SlaveWorldview be sent to masterNetworkEvents?
					slaveIsTimedOutList[worldview.NetworkID-1] = false
				}
				fmt.Printf("network_server.go case broadcastEvents. Sending SlaveWorldview to masterNetworkEvents channel. worldview = %v\n", worldview)
				masterNetworkEvents <- worldview
				resetTimer(slaveHeartbeatTimers[worldview.NetworkID-1])
				fmt.Printf("network_server.go case broadcastEvents. Reset slave heartbeat timer for NetworkID %d\n", worldview.NetworkID)
			default:
				fmt.Printf("network_server.go case broadcastEvents. Received unknown worldview type: %T\n", worldview)
			}
		case masterCommand := <-masterNetworkCommands:
			fmt.Println("network_server.go case masterNetworkCommands.")
			packet := worldviewToPacket(masterCommand)
			broadcastCommands <- packet
		case slaveCommand := <-slaveNetworkCommands:
			fmt.Println("network_server.go case slaveNetworkCommands.")
			packet := worldviewToPacket(slaveCommand)
			broadcastCommands <- packet
		case <-masterHeartbeatTimer.C:
			fmt.Println("network_server.go case masterHeartbeatTimer.C. Master heartbeat timeout.")
			masterIsTimedOut = true
			masterNetworkEvents <- MasterTimeout(0)
			// Take appropriate action for master timeout
		case slaveNetworkID := <-slaveTimeouts:
			fmt.Printf("network_server.go case slaveTimeouts. Slave %d heartbeat timeout.\n", slaveNetworkID)
			masterNetworkEvents <- SlaveTimeout{NetworkID: slaveNetworkID}
			slaveIsTimedOutList[slaveNetworkID-1] = true
		}
	}
}

// Don't panic incase of receiving packets from unintented senders.
func packetToWorldview(packet []byte) interface{} {
	var typeTagged typeTaggedJSON
	if errorPacket := json.Unmarshal(packet, &typeTagged); errorPacket != nil {
		fmt.Printf("Failed to decode packet to typeTaggedJSON: %v\n", errorPacket)
		return nil
	}

	switch typeTagged.Type {
	case reflect.TypeFor[MasterWorldview]().String():
		var worldview MasterWorldview
		if errorPayload := json.Unmarshal(typeTagged.Payload, &worldview); errorPayload != nil {
			fmt.Printf("Failed to decode payload to MasterWorldview: %v\n", errorPayload)
			return nil
		}
		return worldview

	case reflect.TypeFor[SlaveWorldview]().String():
		var worldview SlaveWorldview
		if errorPayload := json.Unmarshal(typeTagged.Payload, &worldview); errorPayload != nil {
			fmt.Printf("Failed to decode payload to SlaveWorldview: %v\n", errorPayload)
			return nil
		}
		return worldview

	default:
		fmt.Printf("Unknown worldview type: %s\n", typeTagged.Type)
		return nil
	}
}

func worldviewToPacket[worldviewType WorldviewType](worldview worldviewType) []byte {
	typeName := reflect.TypeFor[worldviewType]().String()
	fmt.Printf("network_server.go worldviewToPacket. typeName = %v\n", typeName)

	jsonData, errorEncodingWorldview := json.Marshal(worldview)
	if errorEncodingWorldview != nil {
		panic(fmt.Sprintf(
			"Failed to encode worldview to JSON (Type: %v, Payload: %v): %v",
			typeName, jsonData, errorEncodingWorldview))
	}

	packet, err := json.Marshal(typeTaggedJSON{
		Type:    typeName,
		Payload: jsonData,
	})
	if err != nil {
		panic(fmt.Sprintf(
			"Failed to encode worldview to typeTaggedJSON (Type: %v, Payload: %v)",
			typeName, jsonData))
	}

	if len(packet) > NetworkBufferSize {
		panic(fmt.Sprintf(
			"Packet too large (length: %d, max: %d)",
			len(packet), NetworkBufferSize))
	}

	return packet
}
