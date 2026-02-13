package network

import (
	"encoding/json"
	"fmt"
	"reflect"

	. "github.com/kritagya03/ttk4145-project/internal/models"
)

type typeTaggedJSON struct {
	Type    string
	Payload []byte
}

type WorldviewType interface {
	MasterWorldview | SlaveWorldview
}

func Server(
	broadcastEvents <-chan []byte,
	masterNetworkCommands <-chan MasterWorldview,
	slaveNetworkCommands <-chan SlaveWorldview,
	broadcastCommands chan<- []byte,
	masterNetworkEvents chan<- SlaveWorldview,
	slaveNetworkEvents chan<- MasterWorldview) {

	for {
		select {
		case broadcastEvent := <-broadcastEvents:
			wordlview := packetToWorldview(broadcastEvent)
			switch worldview := wordlview.(type) {
			case MasterWorldview:
				fmt.Printf("network_server.go case broadcastEvents. Sending MasterWorldview to slaveNetworkEvents channel. worldview = %v\n", worldview)
				slaveNetworkEvents <- worldview
			case SlaveWorldview:
				fmt.Printf("network_server.go case broadcastEvents. Sending SlaveWorldview to masterNetworkEvents channel. worldview = %v\n", worldview)
				masterNetworkEvents <- worldview
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
		}
	}
}

func packetToWorldview(packet []byte) interface{} {
	var typeTagged typeTaggedJSON
	if errorPacket := json.Unmarshal(packet, &typeTagged); errorPacket != nil {
		panic(fmt.Sprintf("Failed to decode packet to typeTaggedJSON: %v", errorPacket))
	}

	switch typeTagged.Type {
	case reflect.TypeFor[MasterWorldview]().String():
		var worldview MasterWorldview
		if errorPayload := json.Unmarshal(typeTagged.Payload, &worldview); errorPayload != nil {
			panic(fmt.Sprintf("Failed to decode payload to MasterWorldview: %v", errorPayload))
		}
		return worldview

	case reflect.TypeFor[SlaveWorldview]().String():
		var worldview SlaveWorldview
		if errorPayload := json.Unmarshal(typeTagged.Payload, &worldview); errorPayload != nil {
			panic(fmt.Sprintf("Failed to decode payload to SlaveWorldview: %v", errorPayload))
		}
		return worldview

	default:
		panic(fmt.Sprintf("Unknown worldview type: %s", typeTagged.Type))
	}
}

func worldviewToPacket[worldviewType WorldviewType](worldview worldviewType) []byte {
	typeName := reflect.TypeFor[worldviewType]().String()
	fmt.Printf("network_server.go worldviewToPacket. typeName = %v\n", typeName)

	jsonData, errorEncodingWorldview := json.Marshal(worldview)
	if errorEncodingWorldview != nil {
		panic(fmt.Sprintf(
			"Failed to encode wordlview to JSON (Type: %v, Payload: %v): %v",
			typeName, jsonData, errorEncodingWorldview))
	}

	packet, err := json.Marshal(typeTaggedJSON{
		Type:    typeName,
		Payload: jsonData,
	})
	if err != nil {
		panic(fmt.Sprintf(
			"Failed to encode wordlview to typeTaggedJSON (Type: %v, Payload: %v)",
			typeName, jsonData))
	}

	bufferSize := 1024 // TODO: Remove hardcoded buffer size
	if len(packet) > bufferSize {
		panic(fmt.Sprintf(
			"Packet too large (length: %d, max: %d)",
			len(packet), bufferSize))
	}

	return packet
}
