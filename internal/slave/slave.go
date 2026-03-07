package slave

import (
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
	slaveHardwareEvents <-chan interface{},
	slaveNetworkCommands chan<- SlaveWorldview,
	slaveHardwareCommands chan<- interface{},
	networkID int,
	floorCount int,
	buttonTypeCount int) {

	slaveWorldview := getDefaultSlaveWorldview(networkID, floorCount, buttonTypeCount)
	// Init door open timeout
	// Init between floors
	heartbeatTicker := time.NewTicker(HeartbeatInterval)
	defer heartbeatTicker.Stop()
	doorOpenTimeout := time.NewTimer(DoorOpenTimeoutDuration)
	doorOpenTimeout.Stop()

	for {
		select {
		case masterEvent := <-slaveNetworkEvents:
			fmt.Printf("slave.go case slaveNetworkEvents. Received MasterWorldview: %+v\n", masterEvent)
			// ! Implement

		case message := <-slaveHardwareEvents:
			switch event := message.(type) {
			case CallButton:
				callButtonEvent := event
				fmt.Println("Call button pressed:", event)
				OnRequestButtonPress(callButtonEvent, &slaveWorldview, doorOpenTimeout, slaveHardwareCommands, networkID)

			case FloorEnter:
				fmt.Println("Floor entered:", event)
				OnFloorArrival(event.Floor, &slaveWorldview, doorOpenTimeout, slaveHardwareCommands)

			case Stop:
				fmt.Println("Stop initiated:", event)
				for floor := range floorCount {
					for callType := range CallType(3) {
						slaveHardwareCommands <- ButtonLamp{
							CallType: callType,
							Floor:    floor,
							TurnOn:   false,
						}
					}
				}

			case DoorObstruction:
				elevator := slaveWorldview
				fmt.Println("Door obstruction detected:", event)
				// TODO: Implementation depends on specific requirements, often pauses the door timer
				if event.IsObstructed && elevator.Behaviour == BehaviourDoorOpen {
					slaveHardwareCommands <- DoorOpenLamp{TurnOn: true}
					if !doorOpenTimeout.Stop() {
						select {
						case <-doorOpenTimeout.C:
						default:
						}
					}
				} else if !event.IsObstructed && elevator.Behaviour == BehaviourDoorOpen {
					resetTimer(doorOpenTimeout, DoorOpenTimeoutDuration)
				}

			case Initialization:
				fmt.Println("Initialization detected:", event)
				if event.Floor == -1 {
					fmt.Println("Elevator started between two floors.")
					OnInitializeBetweenFloors(&slaveWorldview, slaveHardwareCommands)
				} else {
					fmt.Printf("Elevator started at a floor==%d\n", event.Floor)
					slaveWorldview.FloorLastVisited = event.Floor
				}

			default:
				panic(fmt.Sprintf("Unknown event type: %T\n", event))
			}

		case <-heartbeatTicker.C:
			fmt.Println("Slave heartbeat. Current SlaveWorldview:", slaveWorldview)
			slaveNetworkCommands <- slaveWorldview

		case <-doorOpenTimeout.C:
			fmt.Println("Door Open Timeout.")
			OnDoorOpenTimeout(&slaveWorldview, doorOpenTimeout, slaveHardwareCommands)
		}
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
		Behaviour:        BehaviourIdle,
		Direction:        DirectionStop,
		FloorLastVisited: -1,
		Calls:            CallsMatrix{Matrix: calls},
	}
}

func callTypeToMatrixIndex(callType CallType, networkID int) int {
	switch callType {
	case CallHallUp:
		return 0
	case CallHallDown:
		return 1
	case CallCab:
		hallCallTypeCount := 2                   // TODO: maybe not hardcode
		return hallCallTypeCount + networkID - 1 // 0-indexed
	default:
		panic(fmt.Sprintf("callTypeToMatrixIndex: invalid callType %v: ", callType))
	}
}

func matrixIndexToCallType(matrixIndex int) CallType {
	switch matrixIndex {
	case int(CallHallUp), int(CallHallDown):
		return CallType(matrixIndex)
	default:
		return CallCab
	}
}

func OnRequestButtonPress(buttonEvent CallButton, elevator *SlaveWorldview, doorOpenTimeout *time.Timer, slaveHardwareCommands chan<- interface{}, networkID int) {
	fmt.Printf("\n\n%s(%d, %d)\n", "OnRequestButtonPress", buttonEvent.Floor, buttonEvent.CallType)
	// e.Print()

	switch elevator.Behaviour {
	case BehaviourDoorOpen:
		if ShouldClearImmediately(*elevator, buttonEvent.Floor, buttonEvent.CallType) {
			fmt.Println("OnRequestButtonPress - BehaviourDoorOpen - ShouldClearImmediately=true: Wanting to start the timer again.")
			resetTimer(doorOpenTimeout, DoorOpenTimeoutDuration)
		} else {
			matrixIndex := callTypeToMatrixIndex(buttonEvent.CallType, networkID)
			elevator.Calls.Matrix[buttonEvent.Floor][matrixIndex] = CallStateOrder
			fmt.Println("OnRequestButtonPress - BehaviourDoorOpen - ShouldClearImmediately=false: Added order to e.Requests.")
		}

	case BehaviourMoving:
		matrixIndex := callTypeToMatrixIndex(buttonEvent.CallType, networkID)
		elevator.Calls.Matrix[buttonEvent.Floor][matrixIndex] = CallStateOrder
		fmt.Println("OnRequestButtonPress - EB_Moving: Added order to e.Requests.")

	case BehaviourIdle:
		matrixIndex := callTypeToMatrixIndex(buttonEvent.CallType, networkID)
		elevator.Calls.Matrix[buttonEvent.Floor][matrixIndex] = CallStateOrder
		fmt.Println("OnRequestButtonPress - BehaviourIdle: Added order to e.Requests.")
		directionBehaviourPair := ChooseDirectionBehaviour(*elevator)
		elevator.Direction = directionBehaviourPair.Direction
		elevator.Behaviour = directionBehaviourPair.Behaviour
		fmt.Println("OnRequestButtonPress - BehaviourIdle: Chosen new direction:", directionBehaviourPair)

		switch directionBehaviourPair.Behaviour {
		case BehaviourDoorOpen:
			slaveHardwareCommands <- DoorOpenLamp{TurnOn: true}
			resetTimer(doorOpenTimeout, DoorOpenTimeoutDuration)
			*elevator = ClearServedCallsAtCurrentFloor(*elevator)
			fmt.Println("OnRequestButtonPress - BehaviourIdle - New behaviour is BehaviourDoorOpen: Opened door, wanting to start the timer, maybe remove order(s) at floor if possible.")
		case BehaviourMoving:
			slaveHardwareCommands <- elevator.Direction
			fmt.Println("OnRequestButtonPress - BehaviourIdle - New behaviour is EB_Moving: set the motor direction: ", elevator.Direction)
		case BehaviourIdle:
			// Do nothing
			fmt.Println("OnRequestButtonPress - BehaviourIdle - New behaviour is BehaviourIdle: Doing nothing")
		}
	}

	fmt.Println("OnRequestButtonPress: Wanting to update all lights.")
	SetAllLights(*elevator, slaveHardwareCommands)
	// fmt.Println("\nNew state:")
	// elevator.Print()
}

// SetAllLights sets the button lamps based on requests
func SetAllLights(elevator SlaveWorldview, slaveHardwareCommands chan<- interface{}) {
	matrix := elevator.Calls.Matrix

	matrixIndices := []int{
		callTypeToMatrixIndex(CallHallUp, elevator.NetworkID),
		callTypeToMatrixIndex(CallHallDown, elevator.NetworkID),
		callTypeToMatrixIndex(CallCab, elevator.NetworkID),
	}

	for floor := range matrix {
		for _, matrixIndex := range matrixIndices {
			callState := matrix[floor][matrixIndex]
			isCallAssigned := int(callState) > 0
			turnOn := isCallAssigned || callState == CallStateCompleted
			slaveHardwareCommands <- ButtonLamp{
				CallType: matrixIndexToCallType(matrixIndex),
				Floor:    floor,
				TurnOn:   turnOn,
			}
		}
	}
	fmt.Println("SetAllLights: updated all lights from e.Requests.")
}

// OnInitializeBetweenFloors moves the elevator down if it starts between floors
func OnInitializeBetweenFloors(elevator *SlaveWorldview, slaveHardwareCommands chan<- interface{}) {
	fmt.Println("OnInitializeBetweenFloors: wanting to move down.")
	elevator.Direction = DirectionDown
	elevator.Behaviour = BehaviourMoving
	slaveHardwareCommands <- DirectionDown

}

// OnFloorArrival handles arriving at a floor
func OnFloorArrival(newFloor int, elevator *SlaveWorldview, doorOpenTimeout *time.Timer, slaveHardwareCommands chan<- interface{}) {
	fmt.Printf("\n\n%s(%d)\n", "OnFloorArrival", newFloor)
	// elevator.Print()

	fmt.Println("OnFloorArrival: wanting to set floor indicator.")
	elevator.FloorLastVisited = newFloor
	slaveHardwareCommands <- FloorIndicator{Floor: newFloor}
	switch elevator.Behaviour {
	case BehaviourMoving:
		if ShouldStop(*elevator) {
			fmt.Println("OnFloorArrival - EB_Moving - elevator.ShouldStop()==True: wanting to stop, open door, maybe clear order(s) at floor, reset timer, update lights.")
			slaveHardwareCommands <- DirectionStop
			slaveHardwareCommands <- DoorOpenLamp{TurnOn: true}
			*elevator = ClearServedCallsAtCurrentFloor(*elevator)
			resetTimer(doorOpenTimeout, DoorOpenTimeoutDuration)
			SetAllLights(*elevator, slaveHardwareCommands)
			elevator.Behaviour = BehaviourDoorOpen
		}
	default:
		// Can enter here if initializing elevator on a floor
		fmt.Printf("OnFloorArrival - elevator.Behaviour==default(%v): Doing nothing.\n", elevator.Behaviour)
		// Should not happen if strictly following FSM, but safe to ignore
	}

	// fmt.Println("\nNew state:")
	// elevator.Print()
}

// OnDoorOpenTimeout handles the door closing
func OnDoorOpenTimeout(elevator *SlaveWorldview, doorOpenTimeout *time.Timer, slaveHardwareCommands chan<- interface{}) {
	fmt.Printf("\n\n%s()\n", "OnDoorOpenTimeout")
	// elevator.Print()

	switch elevator.Behaviour {
	case BehaviourDoorOpen:
		fmt.Println("OnDoorOpenTimeout - BehaviourDoorOpen: choosing new direction.")
		directionBehaviourPair := ChooseDirectionBehaviour(*elevator)
		elevator.Direction = directionBehaviourPair.Direction
		elevator.Behaviour = directionBehaviourPair.Behaviour

		switch elevator.Behaviour {
		case BehaviourDoorOpen:
			fmt.Println("OnDoorOpenTimeout - BehaviourDoorOpen - new behaviour is BehaviourDoorOpen: wanting to reset timer, maybe clear order(s) at floor, updating lights.")
			resetTimer(doorOpenTimeout, DoorOpenTimeoutDuration)
			*elevator = ClearServedCallsAtCurrentFloor(*elevator)
			SetAllLights(*elevator, slaveHardwareCommands)
		case BehaviourMoving, BehaviourIdle:
			fmt.Printf("OnDoorOpenTimeout - BehaviourDoorOpen - new behaviour is %v: opening door, setting direction %v.\n", elevator.Behaviour, elevator.Direction)
			slaveHardwareCommands <- DoorOpenLamp{TurnOn: false}
			slaveHardwareCommands <- elevator.Direction
		}

	default:
		// If initializing the elevator on a floor, this happens
		fmt.Printf("OnDoorOpenTimeout - elevator.Behaviour==default(%v): Doing nothing.\n", elevator.Behaviour)
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

// TODO: We are here. Continue below.

type DirectionBehaviourPair struct {
	Direction ElevatorDirection
	Behaviour ElevatorBehaviour
}

func ChooseDirectionBehaviour(elevator SlaveWorldview) DirectionBehaviourPair {
	switch elevator.Direction {
	case DirectionUp:
		if hasCallsAbove(elevator) {
			return DirectionBehaviourPair{DirectionUp, BehaviourMoving}
		} else if hasCallsHere(elevator) {
			return DirectionBehaviourPair{DirectionDown, BehaviourDoorOpen} // Intention of going down because previous if statement verified that there are no requests above
		} else if hasCallsBelow(elevator) {
			return DirectionBehaviourPair{DirectionDown, BehaviourMoving}
		} else {
			return DirectionBehaviourPair{DirectionStop, BehaviourIdle}
		}
	case DirectionDown:
		if hasCallsBelow(elevator) {
			return DirectionBehaviourPair{DirectionDown, BehaviourMoving}
		} else if hasCallsHere(elevator) {
			return DirectionBehaviourPair{DirectionUp, BehaviourDoorOpen} // Intention of going up because previous if statement verified that there are no requests below
		} else if hasCallsAbove(elevator) {
			return DirectionBehaviourPair{DirectionUp, BehaviourMoving}
		} else {
			return DirectionBehaviourPair{DirectionStop, BehaviourIdle}
		}
	case DirectionStop:
		if hasCallsHere(elevator) {
			return DirectionBehaviourPair{DirectionStop, BehaviourDoorOpen}
		} else if hasCallsAbove(elevator) {
			return DirectionBehaviourPair{DirectionUp, BehaviourMoving}
		} else if hasCallsBelow(elevator) {
			return DirectionBehaviourPair{DirectionDown, BehaviourMoving}
		} else {
			return DirectionBehaviourPair{DirectionStop, BehaviourIdle}
		}
	default:
		return DirectionBehaviourPair{DirectionStop, BehaviourIdle}
	}
}

func isCallAssignedToElevator(callState CallState, networkID int) bool {
	if int(callState) == networkID {
		return true
	}
	return false
}

// ShouldStop implements requests_shouldStop
func ShouldStop(elevator SlaveWorldview) bool {
	callHallDownIndex := callTypeToMatrixIndex(CallHallDown, elevator.NetworkID)
	callHallUpIndex := callTypeToMatrixIndex(CallHallUp, elevator.NetworkID)
	callCabIndex := callTypeToMatrixIndex(CallCab, elevator.NetworkID)

	switch elevator.Direction {
	case DirectionDown:
		return isCallAssignedToElevator(elevator.Calls.Matrix[elevator.FloorLastVisited][callHallDownIndex], elevator.NetworkID) ||
			isCallAssignedToElevator(elevator.Calls.Matrix[elevator.FloorLastVisited][callCabIndex], elevator.NetworkID) ||
			!hasCallsBelow(elevator)
	case DirectionUp:
		return isCallAssignedToElevator(elevator.Calls.Matrix[elevator.FloorLastVisited][callHallUpIndex], elevator.NetworkID) ||
			isCallAssignedToElevator(elevator.Calls.Matrix[elevator.FloorLastVisited][callCabIndex], elevator.NetworkID) ||
			!hasCallsAbove(elevator)
	case DirectionStop:
		fallthrough
	default:
		return true
	}
}

// ShouldClearImmediately implements requests_shouldClearImmediately
func ShouldClearImmediately(elevator SlaveWorldview, buttonFloor int, buttonCallType CallType) bool {
	return elevator.FloorLastVisited == buttonFloor &&
		((elevator.Direction == DirectionUp && buttonCallType == CallHallUp) ||
			(elevator.Direction == DirectionDown && buttonCallType == CallHallDown) ||
			elevator.Direction == DirectionStop ||
			buttonCallType == CallCab)
}

// ClearServedCallsAtCurrentFloor implements requests_clearAtCurrentFloor
// This modifies the elevator state in place.
func ClearServedCallsAtCurrentFloor(elevator SlaveWorldview) SlaveWorldview {
	elevator.Calls.Matrix[elevator.FloorLastVisited][CallCab] = CallStateCompleted

	switch elevator.Direction {
	case DirectionUp:
		if !hasCallsAbove(elevator) && !isCallAssignedToElevator(elevator.Calls.Matrix[elevator.FloorLastVisited][CallHallUp], elevator.NetworkID) {
			elevator.Calls.Matrix[elevator.FloorLastVisited][CallHallDown] = CallStateCompleted // No one wants to go up, therefore if call down, it can be cleared because it will immediatly be server
		}
		elevator.Calls.Matrix[elevator.FloorLastVisited][CallHallUp] = CallStateCompleted
	case DirectionDown:
		if !hasCallsBelow(elevator) && !isCallAssignedToElevator(elevator.Calls.Matrix[elevator.FloorLastVisited][CallHallDown], elevator.NetworkID) {
			elevator.Calls.Matrix[elevator.FloorLastVisited][CallHallUp] = CallStateCompleted
		}
		elevator.Calls.Matrix[elevator.FloorLastVisited][CallHallDown] = CallStateCompleted
	case DirectionStop:
		fallthrough
	default:
		elevator.Calls.Matrix[elevator.FloorLastVisited][CallHallUp] = CallStateCompleted
		elevator.Calls.Matrix[elevator.FloorLastVisited][CallHallDown] = CallStateCompleted
	}
	return elevator
}
