package models

// type ElevatorDirection int

// Error because also implemented in types.go
// const (
// 	DirectionUp ElevatorDirection = iota
// 	DirectionDown
// 	DirectionStop
// )

type CallType int

const (
	CallHallUp CallType = iota
	CallHallDown
	CallCab
)

type HardwareEvent interface {
	CallButton | FloorEnter | Stop | Obstruction
}

type HardwareCommand interface {
	MotorDirection | ButtonLamp | FloorIndicator | DoorOpenLamp | StopLamp
}

type CallButton struct {
	Floor    int
	CallType CallType
}

type FloorEnter struct {
	Floor int
}

type Stop struct {
	ToStop bool
}

type Obstruction struct {
	IsObstructed bool
}

type MotorDirection struct {
	Direction ElevatorDirection
}

type ButtonLamp struct {
	CallType     CallType
	Floor        int
	TurnOn bool
}

type FloorIndicator struct {
	Floor int
}

type DoorOpenLamp struct {
	TurnOn bool
}

type StopLamp struct {
	TurnOn bool
}
