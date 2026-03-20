package behaviour

const (
	Idle State = iota
	DoorOpen
	Moving
)

type State int
