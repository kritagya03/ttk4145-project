package hardware

// import (
// 	"Driver-go/elevio"
// 	"fmt"

// 	. "github.com/kritagya03/ttk4145-project/internal/models"
// )

// // ! TODO: move typeTaggedJSON to a common package if it is used in multiple places, e.g., network_server.go and hardware_server.go
// type typeTaggedJSON struct {
// 	Type    string
// 	Payload []byte
// }

// // func Server[hardwareEventType HardwareEvent, hardwareCommandType HardwareCommand](
// func Server(
// 	hardwareEvents <-chan []byte,
// 	slaveHardwareCommands <-chan interface{},
// 	hardwareCommands chan<- []byte,
// 	slaveHardwareEvents chan<- interface{},
// 	floorCount int) {

// 	elevio.Init("localhost:15657", floorCount)

// 	buttonEventReceiver := make(chan elevio.ButtonEvent)
// 	floorEventReceiver := make(chan int)
// 	obstructionEventReceiver := make(chan bool)
// 	stopEventReceiver := make(chan bool)

// 	go elevio.PollButtons(buttonEventReceiver)
// 	go elevio.PollFloorSensor(floorEventReceiver)
// 	go elevio.PollObstructionSwitch(obstructionEventReceiver)
// 	go elevio.PollStopButton(stopEventReceiver)

// 	// On startup send the current floor to slave.go such that it can initialize (maybe inbetween floors). Send Initialization{Floor: XXX} first to slave.go

// 	for {
// 		select {
// 		// case hardwareEventPacket := <-hardwareEvents:
// 		// 	fmt.Printf("hardware_server.go received hardware event: %s\n", string(hardwareEventPacket))
// 		// 	convertedEvent := packetToHardwareEvent(hardwareEventPacket)
// 		// 	if convertedEvent != nil {
// 		// 		slaveHardwareEvents <- convertedEvent
// 		// 	}

// 		// case slaveHardwareCommand := <-slaveHardwareCommands:
// 		// 	fmt.Println("hardware_server.go case slaveHardwareCommands.")
// 		// 	packet := slaveHardwareCommandToPacket(slaveHardwareCommand)
// 		// 	hardwareCommands <- packet
// 		// }

// 		case buttonEvent := <-buttonEventReceiver:
// 			fmt.Printf("hardware_server.go received hardware event: %v\n", buttonEvent)
// 			slaveHardwareEvents <- buttonEvent

// 		case floorEvent := <-floorEventReceiver:
// 			fmt.Printf("hardware_server.go received hardware event: %v\n", floorEvent)
// 			slaveHardwareEvents <- floorEvent

// 		case obstructionEvent := <-obstructionEventReceiver:
// 			fmt.Printf("hardware_server.go received hardware event: %v\n", obstructionEvent)
// 			slaveHardwareEvents <- obstructionEvent

// 		case stopEvent := <-stopEventReceiver:
// 			fmt.Printf("hardware_server.go received hardware event: %v\n", stopEvent)
// 			slaveHardwareEvents <- stopEvent

// 		case slaveHardwareCommand := <-slaveHardwareCommands:
// 			fmt.Printf("hardware_server.go received slave hardware command: %v\n", slaveHardwareCommand)

// 		}

// 		// case button := <-buttonEventReceiver:
// 		// 	fmt.Printf("%+v\n", button)
// 		// 	elevio.SetButtonLamp(button.Button, button.Floor, true)

// 		// case floor := <-floorEventReceiver:
// 		// 	fmt.Printf("%+v\n", floor)
// 		// 	if floor == floorCount-1 {
// 		// 		motorDirection = elevio.MD_Down
// 		// 	} else if floor == 0 {
// 		// 		motorDirection = elevio.MD_Up
// 		// 	}
// 		// 	elevio.SetMotorDirection(motorDirection)

// 		// case obstruction := <-obstructionEventReceiver:
// 		// 	fmt.Printf("%+v\n", obstruction)
// 		// 	if obstruction {
// 		// 		elevio.SetMotorDirection(elevio.MD_Stop)
// 		// 	} else {
// 		// 		elevio.SetMotorDirection(motorDirection)
// 		// 	}

// 		// case stop := <-stopEventReceiver:
// 		// 	fmt.Printf("%+v\n", stop)
// 		// 	for floor := 0; floor < floorCount; floor++ {
// 		// 		for buttonType := elevio.ButtonType(0); buttonType < 3; buttonType++ {
// 		// 			elevio.SetButtonLamp(buttonType, floor, false)
// 		// 		}
// 		// 	}
// 		// }
// 	}
// }

// // func packetToHardwareEvent(packet []byte) interface{} {
// // 	packetType := packet[0]
// // 	callButtonType := 6

// // 	switch packetType {
// // 	case callButtonType:
// // 		// [4]byte{6, byte(callType), byte(floor), 0} // writing is like this
// // 		callTypeIndex := 1
// // 		floorIndex := 2
// // 		callType := CallType(packet[callTypeIndex])
// // 		floor := int(packet[floorIndex])

// // 		isPressed := toBool(packet[1])
// // 		previousCallButtons[floor][callType] = isPressed
// // 		if isPressed != previousCallButtons[floor][callType] && isPressed != false {
// // 			return CallButton {
// // 				Floor: floor
// // 				CallType: CallType(callType)
// // 			}
// // 		} else {
// // 			return nil
// // 		}

// // 	case reflect.TypeFor[FloorEnter]().String():
// // 		var floorEnter FloorEnter
// // 		if errorPayload := json.Unmarshal(typeTagged.Payload, &floorEnter); errorPayload != nil {
// // 			fmt.Printf("Failed to decode payload to FloorEnter: %v\n", errorPayload)
// // 			return nil
// // 		}
// // 		return floorEnter

// // 	case reflect.TypeFor[Stop]().String():
// // 		var stop Stop
// // 		if errorPayload := json.Unmarshal(typeTagged.Payload, &stop); errorPayload != nil {
// // 			fmt.Printf("Failed to decode payload to Stop: %v\n", errorPayload)
// // 			return nil
// // 		}
// // 		return stop

// // 	case reflect.TypeFor[DoorObstruction]().String():
// // 		var doorObstruction DoorObstruction
// // 		if errorPayload := json.Unmarshal(typeTagged.Payload, &doorObstruction); errorPayload != nil {
// // 			fmt.Printf("Failed to decode payload to DoorObstruction: %v\n", errorPayload)
// // 			return nil
// // 		}
// // 		return doorObstruction

// // 	case reflect.TypeFor[Initialization]().String():
// // 		var initialization Initialization
// // 		if errorPayload := json.Unmarshal(typeTagged.Payload, &initialization); errorPayload != nil {
// // 			fmt.Printf("Failed to decode payload to Initialization: %v\n", errorPayload)
// // 			return nil
// // 		}
// // 		return initialization

// // 	default:
// // 		fmt.Printf("Unknown hardware event type: %s\n", typeTagged.Type)
// // 		return nil
// // 	}
// // }

// // // func hardwareCommandToPacket[hardwareCommandType HardwareCommand](hardwareCommand hardwareCommandType) []byte {
// // func slaveHardwareCommandToPacket(slaveHardwareCommand interface{}) []byte {

// // 	// typeName := reflect.TypeOf(slaveHardwareCommand).String()

// // 	// switch typeName {
// // 	// case reflect.TypeFor[CallButton]().String():
// // 	// 	var callButton CallButton
// // 	// 	if errorPayload := json.Unmarshal(typeTagged.Payload, &callButton); errorPayload != nil {
// // 	// 		fmt.Printf("Failed to decode payload to CallButton: %v\n", errorPayload)
// // 	// 		return nil
// // 	// 	}
// // 	// 	return callButton

// // 	// case reflect.TypeFor[FloorEnter]().String():
// // 	// 	var floorEnter FloorEnter
// // 	// 	if errorPayload := json.Unmarshal(typeTagged.Payload, &floorEnter); errorPayload != nil {
// // 	// 		fmt.Printf("Failed to decode payload to FloorEnter: %v\n", errorPayload)
// // 	// 		return nil
// // 	// 	}
// // 	// 	return floorEnter

// // 	// case reflect.TypeFor[Stop]().String():
// // 	// 	var stop Stop
// // 	// 	if errorPayload := json.Unmarshal(typeTagged.Payload, &stop); errorPayload != nil {
// // 	// 		fmt.Printf("Failed to decode payload to Stop: %v\n", errorPayload)
// // 	// 		return nil
// // 	// 	}
// // 	// 	return stop

// // 	// case reflect.TypeFor[DoorObstruction]().String():
// // 	// 	var doorObstruction DoorObstruction
// // 	// 	if errorPayload := json.Unmarshal(typeTagged.Payload, &doorObstruction); errorPayload != nil {
// // 	// 		fmt.Printf("Failed to decode payload to DoorObstruction: %v\n", errorPayload)
// // 	// 		return nil
// // 	// 	}
// // 	// 	return doorObstruction

// // 	// case reflect.TypeFor[Initialization]().String():
// // 	// 	var initialization Initialization
// // 	// 	if errorPayload := json.Unmarshal(typeTagged.Payload, &initialization); errorPayload != nil {
// // 	// 		fmt.Printf("Failed to decode payload to Initialization: %v\n", errorPayload)
// // 	// 		return nil
// // 	// 	}
// // 	// 	return initialization

// // 	// default:
// // 	// 	fmt.Printf("Unknown hardware event type: %s\n", typeTagged.Type)
// // 	// 	return nil
// // 	// }

// // 	return []byte{0} // ! TEMP

// // 	// typeName := reflect.TypeOf(slaveHardwareCommand).String()
// // 	// fmt.Printf("hardware_server.go slaveHardwareCommandToPacket. typeName = %v\n", typeName)

// // 	// jsonData, errorEncodingWorldview := json.Marshal(slaveHardwareCommand)
// // 	// if errorEncodingWorldview != nil {
// // 	// 	panic(fmt.Sprintf(
// // 	// 		"Failed to encode hardware command to JSON (Type: %v, Payload: %v): %v",
// // 	// 		typeName, jsonData, errorEncodingWorldview))
// // 	// }

// // 	// packet, err := json.Marshal(typeTaggedJSON{
// // 	// 	Type:    typeName,
// // 	// 	Payload: jsonData,
// // 	// })
// // 	// if err != nil {
// // 	// 	panic(fmt.Sprintf(
// // 	// 		"Failed to encode hardware command to typeTaggedJSON (Type: %v, Payload: %v)",
// // 	// 		typeName, jsonData))
// // 	// }

// // 	// if len(packet) > NetworkBufferSize {
// // 	// 	panic(fmt.Sprintf(
// // 	// 		"Packet too large (length: %d, max: %d)",
// // 	// 		len(packet), NetworkBufferSize))
// // 	// }

// // 	// return packet
// // }
