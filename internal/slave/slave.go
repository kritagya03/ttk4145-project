package slave

import (
	"Driver-go/elevio"
	"fmt"
	"time"

	. "github.com/kritagya03/ttk4145-project/internal/models"
)

// 1. Implement initialization (e.g. initialize between floors)
// 2. Implement marking call orders
// 3. Implement marking call completed
// 4. Implement arriving at floor
// 5. Implement door timeout
// 6. Implement checking if has requests above, here, below
// 7. Implement choosing direction
// 8. Implement if should stop
// 9. Implement turning on and off lights
// 10. Implement setting motor
// 11. Implement setting floor indicator
// 12. Implement acceptance tests
// 13. Implement door obstruction
// 14. Implement motor stuck

func Server(
	slaveNetworkEvents <-chan MasterWorldview,
	slaveNetworkCommands chan<- SlaveWorldview,
	networkID int,
	floorCount int,
	buttonTypeCount int,
	hardwarePort int) {

	slaveWorldview := getDefaultSlaveWorldview(networkID, floorCount, buttonTypeCount)

	// setAllLights(slaveWorldview)

	heartbeatTicker := time.NewTicker(HeartbeatInterval)
	defer heartbeatTicker.Stop()
	doorOpenTimeout := time.NewTimer(DoorOpenTimeoutDuration)
	doorOpenTimeout.Stop()

	buttonEventReceiver := make(chan elevio.ButtonEvent)
	floorEventReceiver := make(chan int)
	obstructionEventReceiver := make(chan bool)
	stopEventReceiver := make(chan bool)

	hardwareAddress := fmt.Sprintf("localhost:%d", hardwarePort)
	elevio.Init(hardwareAddress, floorCount)

	go elevio.PollButtons(buttonEventReceiver)
	go elevio.PollFloorSensor(floorEventReceiver)
	go elevio.PollObstructionSwitch(obstructionEventReceiver)
	go elevio.PollStopButton(stopEventReceiver)

	dontCloseDoor := elevio.GetObstruction()

	if !dontCloseDoor {
		resetTimer(doorOpenTimeout, DoorOpenTimeoutDuration)
	}

	if floor := elevio.GetFloor(); floor == -1 {
		fmt.Println("Elevator started between two floors.")
		onInitializeBetweenFloors(&slaveWorldview)
	} else {
		fmt.Printf("Elevator started at a floor==%d\n", floor)
		slaveWorldview.FloorLastVisited = floor
	}

	for {
		select {
		case masterWorldview := <-slaveNetworkEvents:
			// fmt.Printf("slave.go case slaveNetworkEvents. Received MasterWorldview: %+v\n", masterWorldview)

			slaveWorldview = getNewSlaveWorldview(slaveWorldview, masterWorldview)

			elevator := &slaveWorldview

			if slaveWorldview.Behaviour == BehaviourIdle {
				directionBehaviourPair := chooseDirectionBehaviour(*elevator)
				elevator.Direction = directionBehaviourPair.Direction
				elevator.Behaviour = directionBehaviourPair.Behaviour
				// fmt.Println("slaveNetworkEvents - BehaviourIdle: Chosen new direction:", directionBehaviourPair)

				switch directionBehaviourPair.Behaviour {
				case BehaviourDoorOpen:
					elevio.SetDoorOpenLamp(true)
					resetTimer(doorOpenTimeout, DoorOpenTimeoutDuration)
					*elevator = clearServedCallsAtCurrentFloor(*elevator)
					// fmt.Println("slaveNetworkEvents - BehaviourIdle - New behaviour is BehaviourDoorOpen: Opened door, wanting to start the doorOpenTimeout timer, maybe remove order(s) at floor if possible.")
				case BehaviourMoving:
					elevio.SetMotorDirection(elevator.Direction)
					// fmt.Println("slaveNetworkEvents - BehaviourIdle - New behaviour is EB_Moving: set the motor direction: ", elevator.Direction)
				case BehaviourIdle:
					// Do nothing
					// fmt.Println("slaveNetworkEvents - BehaviourIdle - New behaviour is BehaviourIdle: Doing nothing")
				}
			}

			// fmt.Println("slaveNetworkEvents: Wanting to update all lights.")
			setAllLights(*elevator)
			// fmt.Println("\nNew state:")
			// elevator.Print()

		case event := <-buttonEventReceiver:
			fmt.Println("Call button pressed:", event)
			onRequestButtonPress(event, &slaveWorldview, doorOpenTimeout, networkID)

		case floor := <-floorEventReceiver:
			fmt.Println("Floor entered:", floor)
			onFloorArrival(floor, &slaveWorldview, doorOpenTimeout)
			obstructionHandler(dontCloseDoor, &slaveWorldview, doorOpenTimeout)
		case isObstructed := <-obstructionEventReceiver:
			fmt.Printf("\nobstructionEventReceiver: isObstructed==%v, setting dontCloseDoor to %v\n", isObstructed, dontCloseDoor)
			obstructionHandler(isObstructed, &slaveWorldview, doorOpenTimeout)
			dontCloseDoor = isObstructed

		case isStopped := <-stopEventReceiver:
			fmt.Println("Stop button event:", isStopped)
			// TODO: Maybe add stop button functionality
			// The below gets overwritten by setAllLights
			// if isStopped {
			// 	for floor := range floorCount {
			// 		for buttonType := range elevio.ButtonType(3) {
			// 			elevio.SetButtonLamp(buttonType, floor, false)
			// 		}
			// 	}
			// }

		case <-heartbeatTicker.C:
			if slaveWorldview.FloorLastVisited < 0 || slaveWorldview.FloorLastVisited >= floorCount {
				fmt.Printf("Invalid floorLastVisited %d, not sending slave heartbeat.\n", slaveWorldview.FloorLastVisited)
				continue
			}
			// fmt.Println("Slave heartbeat. Current SlaveWorldview:", slaveWorldview)
			slaveNetworkCommands <- slaveWorldview

		case <-doorOpenTimeout.C:
			fmt.Println("Door Open Timeout.")
			onDoorOpenTimeout(&slaveWorldview, doorOpenTimeout)
		}
	}
}

func obstructionHandler(isObstructed bool, elevator *SlaveWorldview, doorOpenTimeout *time.Timer) {
	fmt.Printf("\nobstructionHandler: isObstructed==%v\n\n", isObstructed)
	fmt.Println("Door obstruction detected:", isObstructed)
	// TODO: Implementation depends on specific requirements, often pauses the door timer
	if isObstructed && elevator.Behaviour == BehaviourDoorOpen {
		elevio.SetDoorOpenLamp(true)
		if !doorOpenTimeout.Stop() {
			select {
			case <-doorOpenTimeout.C:
			default:
			}
		}
	} else if !isObstructed && elevator.Behaviour == BehaviourDoorOpen {
		resetTimer(doorOpenTimeout, DoorOpenTimeoutDuration)
	}

}

// TODO: This is reused from network_server.go, master.go
func resetTimer(timer *time.Timer, duration time.Duration) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(duration)
}

func getDefaultSlaveWorldview(networkID int, floorCount int, buttonTypeCount int) SlaveWorldview {
	calls := make([][]CallState, floorCount)
	for floor := range floorCount {
		calls[floor] = make([]CallState, buttonTypeCount)
		for buttonType := range buttonTypeCount {
			calls[floor][buttonType] = CallStateNone
		}
	}
	return SlaveWorldview{
		NetworkID:        networkID,
		Behaviour:        BehaviourDoorOpen,
		Direction:        elevio.MD_Stop,
		FloorLastVisited: -1,
		Calls:            CallsMatrix{Matrix: calls},
	}
}

func buttonTypeToMatrixIndex(buttonType elevio.ButtonType, networkID int) int {
	switch buttonType {
	case elevio.BT_HallUp:
		return 0
	case elevio.BT_HallDown:
		return 1
	case elevio.BT_Cab:
		hallButtonTypeTypeCount := 2                   // TODO: maybe not hardcode
		return hallButtonTypeTypeCount + networkID - 1 // 0-indexed
	default:
		panic(fmt.Sprintf("buttonTypeToMatrixIndex: invalid buttonType %v: ", buttonType))
	}
}

func matrixIndexToButtonType(matrixIndex int) elevio.ButtonType {
	switch matrixIndex {
	case int(elevio.BT_HallUp), int(elevio.BT_HallDown):
		return elevio.ButtonType(matrixIndex)
	default:
		return elevio.BT_Cab
	}
}

func onRequestButtonPress(buttonEvent elevio.ButtonEvent, elevator *SlaveWorldview, doorOpenTimeout *time.Timer, networkID int) {
	fmt.Printf("\n\n%s(%d, %d)\n", "onRequestButtonPress", buttonEvent.Floor, buttonEvent.Button)
	// e.Print()

	switch elevator.Behaviour {
	case BehaviourDoorOpen:
		if shouldClearImmediately(*elevator, buttonEvent.Floor, buttonEvent.Button) {
			fmt.Println("onRequestButtonPress - BehaviourDoorOpen - shouldClearImmediately=true: Wanting to start the timer again.")
			resetTimer(doorOpenTimeout, DoorOpenTimeoutDuration)
		} else {
			matrixIndex := buttonTypeToMatrixIndex(buttonEvent.Button, networkID)
			elevator.Calls.Matrix[buttonEvent.Floor][matrixIndex] = CallStateOrder
			fmt.Println("onRequestButtonPress - BehaviourDoorOpen - shouldClearImmediately=false: Added order.")
		}

	case BehaviourMoving:
		matrixIndex := buttonTypeToMatrixIndex(buttonEvent.Button, networkID)
		elevator.Calls.Matrix[buttonEvent.Floor][matrixIndex] = CallStateOrder
		fmt.Println("onRequestButtonPress - EB_Moving: Added order.")

	case BehaviourIdle:
		matrixIndex := buttonTypeToMatrixIndex(buttonEvent.Button, networkID)
		elevator.Calls.Matrix[buttonEvent.Floor][matrixIndex] = CallStateOrder
		fmt.Println("onRequestButtonPress - BehaviourIdle: Added order.")
		// directionBehaviourPair := chooseDirectionBehaviour(*elevator)
		// elevator.Direction = directionBehaviourPair.Direction
		// elevator.Behaviour = directionBehaviourPair.Behaviour
		// fmt.Println("onRequestButtonPress - BehaviourIdle: Chosen new direction:", directionBehaviourPair)

		// switch directionBehaviourPair.Behaviour {
		// case BehaviourDoorOpen:
		// 	elevio.SetDoorOpenLamp(true)
		// 	resetTimer(doorOpenTimeout, DoorOpenTimeoutDuration)
		// 	*elevator = clearServedCallsAtCurrentFloor(*elevator)
		// 	fmt.Println("onRequestButtonPress - BehaviourIdle - New behaviour is BehaviourDoorOpen: Opened door, wanting to start the timer, maybe remove order(s) at floor if possible.")
		// case BehaviourMoving:
		// 	elevio.SetMotorDirection(elevator.Direction)
		// 	fmt.Println("onRequestButtonPress - BehaviourIdle - New behaviour is EB_Moving: set the motor direction: ", elevator.Direction)
		// case BehaviourIdle:
		// 	// Do nothing
		// 	fmt.Println("onRequestButtonPress - BehaviourIdle - New behaviour is BehaviourIdle: Doing nothing")
		// }
	}

	// fmt.Println("onRequestButtonPress: Wanting to update all lights.")
	// setAllLights(*elevator)
	// // fmt.Println("\nNew state:")
	// // elevator.Print()
}

// setAllLights sets the button lamps based on requests
func setAllLights(elevator SlaveWorldview) {
	matrix := elevator.Calls.Matrix

	matrixIndices := []int{
		buttonTypeToMatrixIndex(elevio.BT_HallUp, elevator.NetworkID),
		buttonTypeToMatrixIndex(elevio.BT_HallDown, elevator.NetworkID),
		buttonTypeToMatrixIndex(elevio.BT_Cab, elevator.NetworkID),
	}

	for floor := range matrix {
		for _, matrixIndex := range matrixIndices {
			callState := matrix[floor][matrixIndex]
			isCallAssignedToAnyone := int(callState) > 0 // TODO: maybe not hardcode
			// turnOn := isCallAssignedToAnyone
			turnOn := isCallAssignedToAnyone // || callState == CallStateCompleted // TODO
			buttonType := matrixIndexToButtonType(matrixIndex)
			elevio.SetButtonLamp(buttonType, floor, turnOn)
		}
	}
	// fmt.Println("setAllLights: updated all lights")
}

// onInitializeBetweenFloors moves the elevator down if it starts between floors
func onInitializeBetweenFloors(elevator *SlaveWorldview) {
	fmt.Println("onInitializeBetweenFloors: wanting to move down.")
	elevator.Direction = elevio.MD_Down
	elevator.Behaviour = BehaviourMoving
	elevio.SetMotorDirection(elevio.MD_Down)

}

// onFloorArrival handles arriving at a floor
func onFloorArrival(newFloor int, elevator *SlaveWorldview, doorOpenTimeout *time.Timer) {
	fmt.Printf("\n\n%s(%d)\n", "onFloorArrival", newFloor)
	// elevator.Print()

	// fmt.Println("onFloorArrival: wanting to set floor indicator.")
	elevator.FloorLastVisited = newFloor
	elevio.SetFloorIndicator(newFloor)
	switch elevator.Behaviour {
	case BehaviourMoving:
		if shouldStop(*elevator) {
			fmt.Println("onFloorArrival - EB_Moving - shouldStop()==True: wanting to stop, open door, maybe clear order(s) at floor, reset doorOpenTimeout timer.")
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			*elevator = clearServedCallsAtCurrentFloor(*elevator)
			resetTimer(doorOpenTimeout, DoorOpenTimeoutDuration)
			// setAllLights(*elevator)
			elevator.Behaviour = BehaviourDoorOpen
		}
	default:
		// Can enter here if initializing elevator on a floor
		fmt.Printf("onFloorArrival - elevator.Behaviour==default(%v): Doing nothing.\n", elevator.Behaviour)
		// Should not happen if strictly following FSM, but safe to ignore
	}

	// fmt.Println("\nNew state:")
	// elevator.Print()
}

// onDoorOpenTimeout handles the door closing
func onDoorOpenTimeout(elevator *SlaveWorldview, doorOpenTimeout *time.Timer) {
	// fmt.Printf("\n\n%s()\n", "onDoorOpenTimeout")
	// elevator.Print()

	switch elevator.Behaviour {
	case BehaviourDoorOpen:
		// fmt.Println("onDoorOpenTimeout - BehaviourDoorOpen: choosing new direction.")
		directionBehaviourPair := chooseDirectionBehaviour(*elevator)
		elevator.Direction = directionBehaviourPair.Direction
		elevator.Behaviour = directionBehaviourPair.Behaviour

		switch elevator.Behaviour {
		case BehaviourDoorOpen:
			fmt.Println("onDoorOpenTimeout - BehaviourDoorOpen - new behaviour is BehaviourDoorOpen: wanting to reset timer, maybe clear order(s) at floor, updating lights.")
			resetTimer(doorOpenTimeout, DoorOpenTimeoutDuration)
			*elevator = clearServedCallsAtCurrentFloor(*elevator)
			// setAllLights(*elevator)
		case BehaviourMoving, BehaviourIdle:
			fmt.Printf("onDoorOpenTimeout - BehaviourDoorOpen - new behaviour is %v: opening door, setting direction %v.\n", elevator.Behaviour, elevator.Direction)
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevator.Direction)
		}

	default:
		// If initializing the elevator on a floor, this happens
		fmt.Printf("onDoorOpenTimeout - elevator.Behaviour==default(%v): Doing nothing.\n", elevator.Behaviour)
		// Should not happen
	}

	// fmt.Println("\nNew state:")
	// elevator.Print()
}

func hasCallsAbove(elevator SlaveWorldview) bool {
	matrix := elevator.Calls.Matrix
	floorCount := len(matrix)
	for floor := elevator.FloorLastVisited + 1; floor < floorCount; floor++ {
		for buttonType := 0; buttonType < len(matrix[floor]); buttonType++ {
			if matrix[floor][buttonType] == CallState(elevator.NetworkID) {
				return true
			}
		}
	}
	return false
}

func hasCallsBelow(elevator SlaveWorldview) bool {
	matrix := elevator.Calls.Matrix
	for floor := 0; floor < elevator.FloorLastVisited; floor++ {
		for buttonType := 0; buttonType < len(matrix[floor]); buttonType++ {
			if matrix[floor][buttonType] == CallState(elevator.NetworkID) {
				return true
			}
		}
	}
	return false
}

func hasCallsHere(elevator SlaveWorldview) bool {
	matrix := elevator.Calls.Matrix
	for _, callState := range matrix[elevator.FloorLastVisited] {
		isCallAssigned := callState == CallState(elevator.NetworkID)
		if isCallAssigned {
			return true
		}
	}
	return false
}

type directionBehaviourPair struct {
	Direction elevio.MotorDirection
	Behaviour ElevatorBehaviour
}

func chooseDirectionBehaviour(elevator SlaveWorldview) directionBehaviourPair {
	switch elevator.Direction {
	case elevio.MD_Up:
		if hasCallsAbove(elevator) {
			return directionBehaviourPair{elevio.MD_Up, BehaviourMoving}
		} else if hasCallsHere(elevator) {
			return directionBehaviourPair{elevio.MD_Down, BehaviourDoorOpen} // Intention of going down because previous if statement verified that there are no requests above
		} else if hasCallsBelow(elevator) {
			return directionBehaviourPair{elevio.MD_Down, BehaviourMoving}
		} else {
			return directionBehaviourPair{elevio.MD_Stop, BehaviourIdle}
		}
	case elevio.MD_Down:
		if hasCallsBelow(elevator) {
			return directionBehaviourPair{elevio.MD_Down, BehaviourMoving}
		} else if hasCallsHere(elevator) {
			return directionBehaviourPair{elevio.MD_Up, BehaviourDoorOpen} // Intention of going up because previous if statement verified that there are no requests below
		} else if hasCallsAbove(elevator) {
			return directionBehaviourPair{elevio.MD_Up, BehaviourMoving}
		} else {
			return directionBehaviourPair{elevio.MD_Stop, BehaviourIdle}
		}
	case elevio.MD_Stop:
		if hasCallsHere(elevator) {
			return directionBehaviourPair{elevio.MD_Stop, BehaviourDoorOpen}
		} else if hasCallsAbove(elevator) {
			return directionBehaviourPair{elevio.MD_Up, BehaviourMoving}
		} else if hasCallsBelow(elevator) {
			return directionBehaviourPair{elevio.MD_Down, BehaviourMoving}
		} else {
			return directionBehaviourPair{elevio.MD_Stop, BehaviourIdle}
		}
	default:
		return directionBehaviourPair{elevio.MD_Stop, BehaviourIdle}
	}
}

// TODO: also used in master.go
func isCallAssigned(callState CallState, elevatorCount int) bool {
	if int(callState) > 0 && int(callState) <= elevatorCount {
		return true
	}
	return false
}

func isCallAssignedToElevator(callState CallState, networkID int) bool {
	if int(callState) == networkID {
		return true
	}
	return false
}

// shouldStop implements requests_shouldStop
func shouldStop(elevator SlaveWorldview) bool {
	buttonHallDownIndex := buttonTypeToMatrixIndex(elevio.BT_HallDown, elevator.NetworkID)
	buttonHallUpIndex := buttonTypeToMatrixIndex(elevio.BT_HallUp, elevator.NetworkID)
	buttonCabIndex := buttonTypeToMatrixIndex(elevio.BT_Cab, elevator.NetworkID)

	switch elevator.Direction {
	case elevio.MD_Down:
		return isCallAssignedToElevator(elevator.Calls.Matrix[elevator.FloorLastVisited][buttonHallDownIndex], elevator.NetworkID) ||
			isCallAssignedToElevator(elevator.Calls.Matrix[elevator.FloorLastVisited][buttonCabIndex], elevator.NetworkID) ||
			!hasCallsBelow(elevator)
	case elevio.MD_Up:
		return isCallAssignedToElevator(elevator.Calls.Matrix[elevator.FloorLastVisited][buttonHallUpIndex], elevator.NetworkID) ||
			isCallAssignedToElevator(elevator.Calls.Matrix[elevator.FloorLastVisited][buttonCabIndex], elevator.NetworkID) ||
			!hasCallsAbove(elevator)
	case elevio.MD_Stop:
		fallthrough
	default:
		return true
	}
}

// shouldClearImmediately implements requests_shouldClearImmediately
func shouldClearImmediately(elevator SlaveWorldview, buttonFloor int, buttonType elevio.ButtonType) bool {
	return elevator.FloorLastVisited == buttonFloor &&
		((elevator.Direction == elevio.MD_Up && buttonType == elevio.BT_HallUp) ||
			(elevator.Direction == elevio.MD_Down && buttonType == elevio.BT_HallDown) ||
			elevator.Direction == elevio.MD_Stop ||
			buttonType == elevio.BT_Cab)
}

func clearIfAssigned(callState CallState, networkID int) CallState {
	if isCallAssignedToElevator(callState, networkID) {
		return CallStateCompleted
	}
	return callState
}

func clearServedCallsAtCurrentFloor(elevator SlaveWorldview) SlaveWorldview {
	floor := elevator.FloorLastVisited
	matrix := elevator.Calls.Matrix
	networkID := elevator.NetworkID

	cabIndex := buttonTypeToMatrixIndex(elevio.BT_Cab, networkID)
	hallUpIndex := buttonTypeToMatrixIndex(elevio.BT_HallUp, networkID)
	hallDownIndex := buttonTypeToMatrixIndex(elevio.BT_HallDown, networkID)

	matrix[floor][cabIndex] = CallStateCompleted // TODO: commented this one out, should it be kept?

	switch elevator.Direction {
	case elevio.MD_Up:
		if !hasCallsAbove(elevator) && !isCallAssignedToElevator(matrix[floor][hallUpIndex], networkID) {
			matrix[floor][hallDownIndex] = clearIfAssigned(matrix[floor][hallDownIndex], networkID)
		}
		matrix[floor][hallUpIndex] = clearIfAssigned(matrix[floor][hallUpIndex], networkID)

	case elevio.MD_Down:
		if !hasCallsBelow(elevator) && !isCallAssignedToElevator(matrix[floor][hallDownIndex], networkID) {
			matrix[floor][hallUpIndex] = clearIfAssigned(matrix[floor][hallUpIndex], networkID)
		}
		matrix[floor][hallDownIndex] = clearIfAssigned(matrix[floor][hallDownIndex], networkID)

	case elevio.MD_Stop:
		fallthrough
	default:
		matrix[floor][hallUpIndex] = clearIfAssigned(matrix[floor][hallUpIndex], networkID)
		matrix[floor][hallDownIndex] = clearIfAssigned(matrix[floor][hallDownIndex], networkID)
	}

	return elevator
}

// TODO: assumes master matrix and slave matrix are of the same dimensions
// ! TODO: many bugs are caused by this function
func getNewSlaveWorldview(slaveWorldview SlaveWorldview, masterWorldview MasterWorldview) SlaveWorldview {
	slaveMatrix := slaveWorldview.Calls.Matrix
	masterMatrix := masterWorldview.Calls.Matrix
	for floor := range slaveMatrix {
		for buttonType := range slaveMatrix[floor] {
			masterCallState := masterMatrix[floor][buttonType]
			slaveCallState := slaveMatrix[floor][buttonType]
			isMasterCallAssignedToAnyone := int(masterCallState) > 0 // TODO: maybe not hardcode
			isSlaveCallAssignedToAnyone := int(slaveCallState) > 0   // TODO: maybe not hardcode

			// TODO: this also implements that slave can go from (assigned to self) to (none) if master says so, but the master should never say so unless some other elevator does the order or the master has lost the order (NEVER LOSE ORDERS)

			if slaveCallState != masterCallState {
				fmt.Println("\n\nslave.go - getNewSlaveWorldview: Detected mismatch in call state for floor ", floor, " button type ", buttonType, ": masterCallState=", masterCallState, " slaveCallState=", slaveCallState, " isMasterCallAssignedToAnyone=", isMasterCallAssignedToAnyone, " isSlaveCallAssignedToAnyone=", isSlaveCallAssignedToAnyone, "\n\n\n")

			// if slaveCallState != CallStateNone {
			// 	fmt.Printf("\n\n!!!!!!!!!!slave.go - getNewSlaveWorldview: For floor %d, button type %d, masterCallState=%v, slaveCallState=%v, isSlaveCallAssignedToAnyone=%v\n\n\n", floor, buttonType, masterCallState, slaveCallState, isSlaveCallAssignedToAnyone)
			// } else if masterCallState != CallStateNone {
			// 	fmt.Printf("\n\n!!!!!!!!!!slave.go - getNewSlaveWorldview: For floor %d, button type %d, masterCallState=%v, slaveCallState=%v, isSlaveCallAssignedToAnyone=%v\n\n\n", floor, buttonType, masterCallState, slaveCallState, isSlaveCallAssignedToAnyone)
			// }

			if masterCallState == CallStateNone {
				if isSlaveCallAssignedToAnyone || slaveCallState == CallStateCompleted {
					slaveMatrix[floor][buttonType] = masterCallState
					// fmt.Printf("\n\nslave.go - getNewSlaveWorldview: Updating call state for floor %d, button type %d\n, from %v to %v\n\n", floor, buttonType, slaveCallState, masterCallState)
				}
			} else if isMasterCallAssignedToAnyone {
				if slaveCallState == CallStateNone || slaveCallState == CallStateOrder || isSlaveCallAssignedToAnyone {
					slaveMatrix[floor][buttonType] = masterCallState
					// fmt.Printf("\n\nslave.go - getNewSlaveWorldview: Updating call state for floor %d, button type %d\n, from %v to %v\n\n", floor, buttonType, slaveCallState, masterCallState)
				}
			} // else {
			// 	// master has the call, but it is assigned to somebody else
			// 	// if we previously believed it belonged to us (or were in completed state), drop it
			// 	if slaveCallState == CallState(slaveWorldview.NetworkID) || slaveCallState == CallStateCompleted {
			// 		slaveMatrix[floor][buttonType] = CallStateNone
			// 	}
			// }
		}
	}
	slaveWorldview.Calls.Matrix = slaveMatrix
	return slaveWorldview
}
