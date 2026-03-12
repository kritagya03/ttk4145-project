package models

// ! TODO: Temporar, remove

// type ButtonTypeWrapper elevio.ButtonType

// func (buttonType ButtonTypeWrapper) String() string {
// 	switch buttonType {
// 	case ButtonTypeWrapper(elevio.BT_HallUp):
// 		return "HallUp"
// 	case ButtonTypeWrapper(elevio.BT_HallDown):
// 		return "HallDown"
// 	case ButtonTypeWrapper(elevio.BT_Cab):
// 		return "Cab"
// 	default:
// 		return fmt.Sprintf("Unknown:%d", buttonType)
// 	}
// }

// type ButtonEventWrapper elevio.ButtonEvent

// func (buttonEvent ButtonEventWrapper) String() string {
// 	return fmt.Sprintf("Floor: %d, Button: %s", buttonEvent.Floor, buttonEvent.Button)
// }

// type MotorDirectionWrapper elevio.MotorDirection

// func (motorDirection MotorDirectionWrapper) String() string {
// 	switch motorDirection {
// 	case MotorDirectionWrapper(elevio.MD_Up):
// 		return "Up"
// 	case MotorDirectionWrapper(elevio.MD_Down):
// 		return "Down"
// 	case MotorDirectionWrapper(elevio.MD_Stop):
// 		return "Stop"
// 	default:
// 		return fmt.Sprintf("Unknown:%d", motorDirection)
// 	}
// }

// ! OLD BELOW

// type ElevatorDirection int

// Error because also implemented in types.go
// const (
// 	DirectionUp ElevatorDirection = iota
// 	DirectionDown
// 	DirectionStop
// )

// type CallType int

// const (
// 	CallHallUp CallType = iota
// 	CallHallDown
// 	CallCab
// )

// type HardwareEvent interface {
// 	CallButton | FloorEnter | Stop | DoorObstruction | Initialization
// }

// type HardwareCommand interface {
// 	ElevatorDirection | ButtonLamp | FloorIndicator | DoorOpenLamp | StopLamp
// }

// type CallButton struct {
// 	Floor    int
// 	CallType CallType
// }

// type FloorEnter struct {
// 	Floor int
// }

// type Stop struct {
// 	ToStop bool
// }

// type DoorObstruction struct {
// 	IsObstructed bool
// }

// type Initialization struct {
// 	Floor int
// }

// type ButtonLamp struct {
// 	CallType CallType
// 	Floor    int
// 	TurnOn   bool
// }

// type FloorIndicator struct {
// 	Floor int
// }

// type DoorOpenLamp struct {
// 	TurnOn bool
// }

// type StopLamp struct {
// 	TurnOn bool
// }
