package models

type Direction int

// Error because also implemented in types.go
// const (
// 	DirectionUp Direction = iota
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
	Direction Direction
}

type ButtonLamp struct {
	CallType     CallType
	Floor        int
	shouldTurnOn bool
}

type FloorIndicator struct {
	Floor int
}

type DoorOpenLamp struct {
	shouldTurnOn bool
}

type StopLamp struct {
	shouldTurnOn bool
}
