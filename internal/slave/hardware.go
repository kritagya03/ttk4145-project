package slave

import (
	elevator "Driver-go/elevio"
	"time"

	"github.com/kritagya03/ttk4145-project/internal/behaviour"
	"github.com/kritagya03/ttk4145-project/internal/call"
	"github.com/kritagya03/ttk4145-project/internal/worldview"
)

func elevatorStartup(motorDirection elevator.MotorDirection, isDoorObstructed bool, doorOpenTimer *time.Timer, doorOpenDuration time.Duration) {
	elevator.SetMotorDirection(motorDirection)
	elevator.SetStopLamp(false)

	if elevator.GetFloor() != -1 && isDoorObstructed {
		elevator.SetDoorOpenLamp(true)
	} else {
		elevator.SetDoorOpenLamp(false)
	}
}

func updateButtonLamps(slaveWorldview worldview.Slave) {
	callsIndices := []int{
		call.GetCallIndex(elevator.BT_HallUp, slaveWorldview.NetworkID),
		call.GetCallIndex(elevator.BT_HallDown, slaveWorldview.NetworkID),
		call.GetCallIndex(elevator.BT_Cab, slaveWorldview.NetworkID),
	}

	for floor := range slaveWorldview.Calls {
		for _, callIndex := range callsIndices {
			buttonType := call.GetButtonType(callIndex)
			callStatus := slaveWorldview.Calls[floor][callIndex]
			turnOn := callStatus.IsAssignedToAnyone() || callStatus == call.Completed
			elevator.SetButtonLamp(buttonType, floor, turnOn)
		}
	}
}

func executeMovementState(slaveWorldview worldview.Slave) {
	switch slaveWorldview.Behaviour {
	case behaviour.DoorOpen:
		elevator.SetDoorOpenLamp(true)
	case behaviour.Moving, behaviour.Idle:
		elevator.SetDoorOpenLamp(false)
		elevator.SetMotorDirection(slaveWorldview.Direction)
	}
}
