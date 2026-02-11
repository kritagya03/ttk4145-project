package network

//empty channel

const floorCount int = 4
const elevatorCount int = 3
const buttonTypeCount int = 2 + elevatorCount // the number 2 is hall down and hall up, elevatorCount because one list of cab calls per elevator

type CallState int

const (
	CallStateNone      CallState = 0
	CallStateOrder     CallState = -1
	CallStateCompleted CallState = -2
)

type Calls struct {
	requests [floorCount][buttonTypeCount]CallState
}

type MasterWorldview struct {
	requests Calls
}

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

type SlaveWorldview struct {
	networkID        int
	behaviour        ElevatorBehaviour
	direction        ElevatorDirection
	floorLastVisited int
	calls            Calls
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
			break // Temporary
		case watchdogCommand := <-watchdogNetworkCommands:
			break // Temporary
		case masterCommand := <-masterNetworkCommands:
			break // Temporary
		case slaveCommand := <-slaveNetworkCommands:
			break // Temporary
		}
	}
}
