package worldview

import (
	elevator "Driver-go/elevio"

	"github.com/kritagya03/ttk4145-project/internal/behaviour"
	"github.com/kritagya03/ttk4145-project/internal/call"
)

type Master struct {
	Calls     call.Calls
	NetworkID int
}

type Slave struct {
	Behaviour        behaviour.State
	Calls            call.Calls
	Direction        elevator.MotorDirection
	FloorLastVisited int
	NetworkID        int
}

func (Master) IsInNetworkToMasterInterface() {}
func (Slave) IsInNetworkToMasterInterface()  {}

func (worldview Master) DeepCopy() Master {
	worldview.Calls = worldview.Calls.DeepCopy()
	return worldview
}

func (worldview Slave) DeepCopy() Slave {
	worldview.Calls = worldview.Calls.DeepCopy()
	return worldview
}

func (worldview Master) Equal(worldviewOther Master) bool {
	return worldview.NetworkID == worldviewOther.NetworkID &&
		worldview.Calls.Equal(worldviewOther.Calls)

}

func (worldview Slave) Equal(worldviewOther Slave) bool {
	return worldview.NetworkID == worldviewOther.NetworkID &&
		worldview.Behaviour == worldviewOther.Behaviour &&
		worldview.Direction == worldviewOther.Direction &&
		worldview.FloorLastVisited == worldviewOther.FloorLastVisited &&
		worldview.Calls.Equal(worldviewOther.Calls)
}

func NewMaster(networkID int, floorCount int, callTypeCount int) Master {
	masterWorldview := Master{
		NetworkID: networkID,
		Calls:     call.NewCalls(floorCount, callTypeCount),
	}

	return masterWorldview
}

func NewSlave(networkID int, floorCount int, callTypeCount int) Slave {
	slaveWorldview := Slave{
		NetworkID:        networkID,
		Behaviour:        behaviour.DoorOpen,
		Direction:        elevator.MD_Stop,
		FloorLastVisited: -1,
		Calls:            call.NewCalls(floorCount, callTypeCount),
	}

	return slaveWorldview
}
