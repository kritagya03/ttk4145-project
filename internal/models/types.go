package models

import (
	"Driver-go/elevio"
	"fmt"
)

type ElevatorBehaviour int

const (
	BehaviourIdle ElevatorBehaviour = iota
	BehaviourDoorOpen
	BehaviourMoving
)

// type ElevatorDirection int

// const (
// 	DirectionUp ElevatorDirection = iota
// 	DirectionDown
// 	DirectionStop
// )

type CallState int

const (
	CallStateNone      CallState = 0
	CallStateOrder     CallState = -1
	CallStateCompleted CallState = -2
)

func (callState CallState) String() string {
	switch callState {
	case CallStateNone:
		return "None"
	case CallStateOrder:
		return "Order"
	case CallStateCompleted:
		return "Completed"
	default:
		return fmt.Sprintf("Assigned:%d", callState)
	}
}

type CallsMatrix struct {
	Matrix [][]CallState
}

type MasterWorldview struct {
	NetworkID int
	Calls     CallsMatrix
}

type SlaveWorldview struct {
	NetworkID        int
	Behaviour        ElevatorBehaviour
	Direction        elevio.MotorDirection
	FloorLastVisited int
	Calls            CallsMatrix
}

// func (slaveWorldview SlaveWorldview) String() string {
// 	return fmt.Sprintf("NetworkID: %d, Behaviour: %s, Direction: %v, FloorLastVisited: %d, Calls: %s", slaveWorldview.NetworkID, slaveWorldview.Behaviour, slaveWorldview.Direction, slaveWorldview.FloorLastVisited, slaveWorldview.Calls)
// }

// func (elevatorBehaviour ElevatorBehaviour) String() string {
// 	return [...]string{"Idle", "DoorOpen", "Moving"}[elevatorBehaviour]
// }

type NewMasterConnection int
type MasterTimeout int

type NewSlaveConnection struct {
	NetworkID int
}

type SlaveTimeout struct {
	NetworkID int
}
