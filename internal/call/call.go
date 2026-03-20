package call

import (
	elevator "Driver-go/elevio"
	"fmt"
	"slices"
)

const (
	None      Status = 0
	Order     Status = -1
	Completed Status = -2
)

const HallCallTypeCount int = 2

type Calls [][]Status

type Status int

func (calls Calls) DeepCopy() Calls {
	newCalls := make(Calls, len(calls))

	for i := range calls {
		newCalls[i] = make([]Status, len(calls[i]))
		copy(newCalls[i], calls[i])
	}

	return newCalls
}

func (calls Calls) Equal(callsOther Calls) bool {
	return slices.EqualFunc(calls, callsOther, func(a, b []Status) bool {
		return slices.Equal(a, b)
	})
}

func (callStatus Status) String() string {
	switch callStatus {
	case None:
		return "None"
	case Order:
		return "Order"
	case Completed:
		return "Completed"
	default:
		return fmt.Sprintf("Assigned(%d)", callStatus)
	}
}

func (callStatus Status) CompleteIfAssignedTo(networkID int) Status {
	if callStatus.IsAssignedTo(networkID) {
		return Completed
	}
	return callStatus
}

func (callStatus Status) IsAssignedToAnyone() bool {
	return int(callStatus) > 0
}

func (callStatus Status) IsAssignedTo(networkID int) bool {
	return int(callStatus) == networkID
}

func GetAssignedTo(networkID int) Status {
	return Status(networkID)
}

func GetCallIndex(buttonType elevator.ButtonType, networkID int) int {
	switch buttonType {
	case elevator.BT_HallUp:
		return 0
	case elevator.BT_HallDown:
		return 1
	case elevator.BT_Cab:
		return HallCallTypeCount + networkID - 1
	default:
		panic(fmt.Sprintf("Invalid buttonType %v.", buttonType))
	}
}

func GetButtonType(callIndex int) elevator.ButtonType {
	switch callIndex {
	case int(elevator.BT_HallUp), int(elevator.BT_HallDown):
		return elevator.ButtonType(callIndex)
	default:
		return elevator.BT_Cab
	}
}

func NewCalls(floorCount int, callTypeCount int) Calls {
	calls := make(Calls, floorCount)

	for floor := range floorCount {
		calls[floor] = make([]Status, callTypeCount)
		for callIndex := range callTypeCount {
			calls[floor][callIndex] = None
		}
	}

	return calls
}
