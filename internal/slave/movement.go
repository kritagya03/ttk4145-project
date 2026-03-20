package slave

import (
	elevator "Driver-go/elevio"
	"fmt"

	"github.com/kritagya03/ttk4145-project/internal/behaviour"
	"github.com/kritagya03/ttk4145-project/internal/call"
	"github.com/kritagya03/ttk4145-project/internal/worldview"
)

type movementState struct {
	Direction elevator.MotorDirection
	Behaviour behaviour.State
}

func getSlaveWorldviewWithNextMovementState(slaveWorldview worldview.Slave) worldview.Slave {
	nextMovementState := determineNextMovementState(slaveWorldview)
	slaveWorldview.Direction = nextMovementState.Direction
	slaveWorldview.Behaviour = nextMovementState.Behaviour
	return slaveWorldview
}

func determineNextMovementState(slaveWorldview worldview.Slave) movementState {
	switch slaveWorldview.Direction {
	case elevator.MD_Up:
		if slaveWorldview.HasAssignedCallsAbove() {
			return movementState{elevator.MD_Up, behaviour.Moving}
		} else if slaveWorldview.HasAssignedCallsHere() {
			return movementState{elevator.MD_Down, behaviour.DoorOpen}
		} else if slaveWorldview.HasAssignedCallsBelow() {
			return movementState{elevator.MD_Down, behaviour.Moving}
		} else {
			return movementState{elevator.MD_Stop, behaviour.Idle}
		}
	case elevator.MD_Down:
		if slaveWorldview.HasAssignedCallsBelow() {
			return movementState{elevator.MD_Down, behaviour.Moving}
		} else if slaveWorldview.HasAssignedCallsHere() {
			return movementState{elevator.MD_Up, behaviour.DoorOpen}
		} else if slaveWorldview.HasAssignedCallsAbove() {
			return movementState{elevator.MD_Up, behaviour.Moving}
		} else {
			return movementState{elevator.MD_Stop, behaviour.Idle}
		}
	case elevator.MD_Stop:
		if slaveWorldview.HasAssignedCallsHere() {
			return movementState{elevator.MD_Stop, behaviour.DoorOpen}
		} else if slaveWorldview.HasAssignedCallsAbove() {
			return movementState{elevator.MD_Up, behaviour.Moving}
		} else if slaveWorldview.HasAssignedCallsBelow() {
			return movementState{elevator.MD_Down, behaviour.Moving}
		} else {
			return movementState{elevator.MD_Stop, behaviour.Idle}
		}
	default:
		panic(fmt.Sprintf("Invalid direction in slaveworldview: %v.", slaveWorldview.Direction))
	}
}

func shouldStop(slaveWorldview worldview.Slave) bool {
	hallDownIndex := call.GetCallIndex(elevator.BT_HallDown, slaveWorldview.NetworkID)
	hallUpIndex := call.GetCallIndex(elevator.BT_HallUp, slaveWorldview.NetworkID)
	cabIndex := call.GetCallIndex(elevator.BT_Cab, slaveWorldview.NetworkID)

	floorCalls := slaveWorldview.Calls[slaveWorldview.FloorLastVisited]

	hallUpCall := floorCalls[hallUpIndex]
	hallDownCall := floorCalls[hallDownIndex]
	cabCall := floorCalls[cabIndex]

	switch slaveWorldview.Direction {
	case elevator.MD_Down:
		return hallDownCall.IsAssignedTo(slaveWorldview.NetworkID) ||
			cabCall.IsAssignedTo(slaveWorldview.NetworkID) ||
			!slaveWorldview.HasAssignedCallsBelow()

	case elevator.MD_Up:
		return hallUpCall.IsAssignedTo(slaveWorldview.NetworkID) ||
			cabCall.IsAssignedTo(slaveWorldview.NetworkID) ||
			!slaveWorldview.HasAssignedCallsAbove()
	}

	return true
}
