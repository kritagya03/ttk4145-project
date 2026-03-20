package slave

import (
	elevator "Driver-go/elevio"
	"fmt"
	"time"

	"github.com/kritagya03/ttk4145-project/internal/behaviour"
	"github.com/kritagya03/ttk4145-project/internal/synchronize"
	"github.com/kritagya03/ttk4145-project/internal/timer"
	"github.com/kritagya03/ttk4145-project/internal/worldview"
)

func Server(networkToSlave <-chan worldview.Master, slaveToNetwork chan<- worldview.Slave,
	networkID int, floorCount int, callTypeCount int, elevatorServerPort int, doorOpenDuration time.Duration, heartbeatInterval time.Duration) {

	slaveWorldview := worldview.NewSlave(networkID, floorCount, callTypeCount)

	heartbeatTicker := time.NewTicker(heartbeatInterval)
	defer heartbeatTicker.Stop()
	doorOpenTimer := time.NewTimer(doorOpenDuration)
	doorOpenTimer.Stop()

	elevatorServerAddress := fmt.Sprintf("localhost:%d", elevatorServerPort)
	elevator.Init(elevatorServerAddress, floorCount)

	isDoorObstructed := elevator.GetObstruction()

	isElevatorBetweenFloors := elevator.GetFloor() == -1
	if isElevatorBetweenFloors {
		fmt.Println("Started elevator between floors.")
		slaveWorldview.Direction = elevator.MD_Down
		slaveWorldview.Behaviour = behaviour.Moving
		if isDoorObstructed {
			timer.Reset(doorOpenTimer, doorOpenDuration)
		}
	} else {
		fmt.Printf("Started elevator at floor %d.\n", elevator.GetFloor())
		slaveWorldview.FloorLastVisited = elevator.GetFloor()
		slaveWorldview.Direction = elevator.MD_Stop
		slaveWorldview.Behaviour = behaviour.Idle
	}

	elevatorStartup(slaveWorldview.Direction, isDoorObstructed, doorOpenTimer, doorOpenDuration)
	updateButtonLamps(slaveWorldview)

	buttonEvents := make(chan elevator.ButtonEvent)
	floorEvents := make(chan int)
	doorObstructionEvents := make(chan bool)
	stopEvents := make(chan bool)

	go elevator.PollButtons(buttonEvents)
	go elevator.PollFloorSensor(floorEvents)
	go elevator.PollObstructionSwitch(doorObstructionEvents)
	go elevator.PollStopButton(stopEvents)

	for {
		select {
		case masterWorldview := <-networkToSlave:
			oldSlaveWorldview := slaveWorldview.DeepCopy()
			slaveWorldview = synchronize.SlaveWorldview(slaveWorldview, masterWorldview)

			if !slaveWorldview.Equal(oldSlaveWorldview) {
				updateButtonLamps(slaveWorldview)

				if slaveWorldview.Behaviour == behaviour.Idle {
					slaveWorldview = getSlaveWorldviewWithNextMovementState(slaveWorldview)
					executeMovementState(slaveWorldview)
					if slaveWorldview.Behaviour == behaviour.DoorOpen {
						timer.Reset(doorOpenTimer, doorOpenDuration)
						slaveWorldview = slaveWorldview.WithCompletedFloorCalls()
					}
				}
			}

		case button := <-buttonEvents:
			fmt.Printf("Call button pressed at floor %d of type %v.\n", button.Floor, button.Button)

			if slaveWorldview.Behaviour == behaviour.DoorOpen && shouldKeepDoorOpen(slaveWorldview, button) && !isDoorObstructed {
				timer.Reset(doorOpenTimer, doorOpenDuration)
			} else {
				slaveWorldview = slaveWorldview.WithNewOrder(button)
			}

		case floor := <-floorEvents:
			fmt.Printf("Floor %d entered.\n", floor)
			slaveWorldview.FloorLastVisited = floor
			elevator.SetFloorIndicator(floor)

			if slaveWorldview.Behaviour == behaviour.Moving && shouldStop(slaveWorldview) {
				elevator.SetMotorDirection(elevator.MD_Stop)

				if slaveWorldview.HasAssignedCallsHere() {
					elevator.SetDoorOpenLamp(true)
					timer.Reset(doorOpenTimer, doorOpenDuration)
					slaveWorldview.Behaviour = behaviour.DoorOpen
					slaveWorldview = slaveWorldview.WithCompletedFloorCalls()
				} else {
					slaveWorldview.Behaviour = behaviour.Idle
				}
			}

		case isDoorObstructedEvent := <-doorObstructionEvents:
			fmt.Printf("Door obstruction toggled to %v.\n", isDoorObstructedEvent)
			isDoorObstructed = isDoorObstructedEvent
			timer.Reset(doorOpenTimer, doorOpenDuration)

		case isStopped := <-stopEvents:
			fmt.Printf("Stop button toggled to %v.\n", isStopped)
			elevator.SetStopLamp(isStopped)

		case <-heartbeatTicker.C:
			isFloorLastVisitedValid := slaveWorldview.FloorLastVisited >= 0 && slaveWorldview.FloorLastVisited < floorCount

			if isFloorLastVisitedValid {
				slaveToNetwork <- slaveWorldview.DeepCopy()
			}

		case <-doorOpenTimer.C:
			if isDoorObstructed {
				timer.Reset(doorOpenTimer, doorOpenDuration)
			} else {
				slaveWorldview = getSlaveWorldviewWithNextMovementState(slaveWorldview)
				executeMovementState(slaveWorldview)

				if slaveWorldview.Behaviour == behaviour.DoorOpen {
					timer.Reset(doorOpenTimer, doorOpenDuration)
					if slaveWorldview.HasAssignedCallsHere() {
						slaveWorldview = slaveWorldview.WithCompletedFloorCalls()
					}
				}
			}
		}
	}
}
