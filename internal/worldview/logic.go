package worldview

import (
	elevator "Driver-go/elevio"

	"github.com/kritagya03/ttk4145-project/internal/call"
)

func (worldviewBase Master) MergedWith(worldviewNew Master) Master {
	worldviewMerged := worldviewBase

	for floor := range worldviewNew.Calls {
		for callIndex := range worldviewNew.Calls[floor] {
			if worldviewMerged.Calls[floor][callIndex] == call.None && worldviewNew.Calls[floor][callIndex].IsAssignedToAnyone() {
				worldviewMerged.Calls[floor][callIndex] = worldviewNew.Calls[floor][callIndex]
			}
		}
	}

	return worldviewMerged
}

func (slaveWorldview Slave) WithCompletedFloorCalls() Slave {
	floor := slaveWorldview.FloorLastVisited
	calls := slaveWorldview.Calls
	networkID := slaveWorldview.NetworkID
	
	cabIndex := call.GetCallIndex(elevator.BT_Cab, networkID)
	calls[floor][cabIndex] = calls[floor][cabIndex].CompleteIfAssignedTo(networkID)
	
	hallUpIndex := call.GetCallIndex(elevator.BT_HallUp, networkID)
	hallDownIndex := call.GetCallIndex(elevator.BT_HallDown, networkID)

	switch slaveWorldview.Direction {
	case elevator.MD_Up:
		calls[floor][hallUpIndex] = calls[floor][hallUpIndex].CompleteIfAssignedTo(networkID)

		// if !slaveWorldview.HasAssignedCallsAbove() && !calls[floor][hallUpIndex].IsAssignedTo(networkID) {
		// 	calls[floor][hallDownIndex] = calls[floor][hallDownIndex].CompleteIfAssignedTo(networkID)
		// }

	case elevator.MD_Down:
		calls[floor][hallDownIndex] = calls[floor][hallDownIndex].CompleteIfAssignedTo(networkID)

		// if !slaveWorldview.HasAssignedCallsBelow() && !calls[floor][hallDownIndex].IsAssignedTo(networkID) {
		// 	calls[floor][hallUpIndex] = calls[floor][hallUpIndex].CompleteIfAssignedTo(networkID)
		// }

	case elevator.MD_Stop:
		calls[floor][hallUpIndex] = calls[floor][hallUpIndex].CompleteIfAssignedTo(networkID)
		calls[floor][hallDownIndex] = calls[floor][hallDownIndex].CompleteIfAssignedTo(networkID)
	}

	slaveWorldview.Calls = calls
	return slaveWorldview
}

func (slaveWorldview Slave) WithNewOrder(buttonEvent elevator.ButtonEvent) Slave {
	callIndex := call.GetCallIndex(buttonEvent.Button, slaveWorldview.NetworkID)
	callStatus := slaveWorldview.Calls[buttonEvent.Floor][callIndex]

	if callStatus == call.None {
		slaveWorldview.Calls[buttonEvent.Floor][callIndex] = call.Order
	}

	return slaveWorldview
}

func (slaveWorldview Slave) HasAssignedCallsAbove() bool {
	floorCount := len(slaveWorldview.Calls)

	for floor := slaveWorldview.FloorLastVisited + 1; floor < floorCount; floor++ {
		for _, callStatus := range slaveWorldview.Calls[floor] {
			if callStatus.IsAssignedTo(slaveWorldview.NetworkID) {
				return true
			}
		}
	}

	return false
}

func (slaveWorldview Slave) HasAssignedCallsBelow() bool {
	for floor := range slaveWorldview.FloorLastVisited {
		for _, callStatus := range slaveWorldview.Calls[floor] {
			if callStatus.IsAssignedTo(slaveWorldview.NetworkID) {
				return true
			}
		}
	}
	return false
}

func (slaveWorldview Slave) HasAssignedCallsHere() bool {
	for _, callStatus := range slaveWorldview.Calls[slaveWorldview.FloorLastVisited] {
		if callStatus.IsAssignedTo(slaveWorldview.NetworkID) {
			return true
		}
	}
	return false
}
