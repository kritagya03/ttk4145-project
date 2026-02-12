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

func Server(broadcastEvents <-chan []byte,
	watchdogNetworkCommands <-chan string,
	masterNetworkCommands <-chan MasterWorldview,
	slaveNetworkCommands <-chan SlaveWorldview,
	broadcastCommands chan<- []byte,
	masterNetworkEvents chan<- SlaveWorldview,
	slaveNetworkEvents chan<- MasterWorldview) {

	for {
		select {
		case broadcastEvent := <-broadcastEvents:
			fmt.Println("network_server.go case broadcastEvents. Received broadcastEvent")

			var packet typeTaggedJSON

			if err := json.Unmarshal(broadcastEvent, &packet); err != nil {
				fmt.Println("network_server.go case broadcastEvents. json.Unmarshal(broadcastEvent, &packet). err=", err)
				continue
			}

			masterWorldviewType := reflect.TypeOf((*MasterWorldview)(nil)).Elem().String()
			slaveWorldviewType := reflect.TypeOf((*SlaveWorldview)(nil)).Elem().String()

			switch packet.Type {
			case masterWorldviewType:
				var value MasterWorldview
				if err := json.Unmarshal(packet.Payload, &value); err != nil {
					fmt.Println("network_server.go case broadcastEvents. json.Unmarshal(packet.Payload, &value). err=", err)
					continue
				}
				fmt.Printf("network_server.go case broadcastEvents. Sending MasterWorldview to slaveNetworkEvents channel. value = %v\n", value)
				slaveNetworkEvents <- value

			case masterWorldviewType:
				var value SlaveWorldview
				if err := json.Unmarshal(packet.Payload, &value); err != nil {
					fmt.Println("network_server.go case broadcastEvents. json.Unmarshal(packet.Payload, &value). err=", err)
					continue
				}
				fmt.Printf("network_server.go case broadcastEvents. Sending SlaveWorldview to masterNetworkEvents channel. value = %v\n", value)
				masterNetworkEvents <- value

			default:
				fmt.Printf("network_server.go case broadcastEvents. packet.Type != typeNameMasterWorldview && packet.Type != typeNameSlaveWorldview. packet.Type=%v. typeNameMasterWorldview=%v, typeNameSlaveWorldview=%v\n", packet.Type, masterWorldviewType, slaveWorldviewType)
				// Ignore packets for other types
				continue
			}

		// case watchdogCommand := <-watchdogNetworkCommands:
		// 	break // Temporary

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

func worldviewToPacket[worldviewType WorldviewType](command worldviewType) []byte {
	typeName := reflect.TypeOf((*worldviewType)(nil)).Elem().String()
	fmt.Printf("network_server.go worldviewToPacket. typeName = %v\n", typeName)

	jsonData, err := json.Marshal(command)
	if err != nil {
		panic(fmt.Sprintf(
			"Failed to encode wordlview to JSON (Type: %v, Payload: %v)",
			typeName, jsonData))
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

	bufferSize := 1024
	if len(packet) > bufferSize {
		panic(fmt.Sprintf(
			"Packet too large (length: %d, max: %d)",
			len(packet), bufferSize))
	}

	return packet
}
