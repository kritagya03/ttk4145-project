package slave

import (
	elevator "Driver-go/elevio"

	"github.com/kritagya03/ttk4145-project/internal/worldview"
)

func shouldKeepDoorOpen(slaveWorldview worldview.Slave, buttonEvent elevator.ButtonEvent) bool {
	if slaveWorldview.FloorLastVisited != buttonEvent.Floor {
		return false
	}

	switch slaveWorldview.Direction {
	case elevator.MD_Up:
		return buttonEvent.Button == elevator.BT_HallUp || buttonEvent.Button == elevator.BT_Cab
	case elevator.MD_Down:
		return buttonEvent.Button == elevator.BT_HallDown || buttonEvent.Button == elevator.BT_Cab
	case elevator.MD_Stop:
		return true
	}

	return false
}
