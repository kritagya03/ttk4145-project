package models

type ElevatorBehaviour int

const (
	BehaviourIdle ElevatorBehaviour = iota
	BehaviourDoorOpen
	BehaviourMoving
)

type ElevatorDirection int

const (
	DirectionUp ElevatorDirection = iota
	DirectionDown
	DirectionStop
)

type CallState int

const (
	CallStateNone      CallState = 0
	CallStateOrder     CallState = -1
	CallStateCompleted CallState = -2
)

type CallsMatrix struct {
	Matrix [][]CallState
}

type MasterWorldview struct {
	Calls CallsMatrix
}

type SlaveWorldview struct {
	NetworkID        int
	Behaviour        ElevatorBehaviour
	Direction        ElevatorDirection
	FloorLastVisited int
	Calls            CallsMatrix
}
