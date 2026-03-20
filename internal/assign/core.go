package assign

import (
	elevator "Driver-go/elevio"
	"fmt"
	"strconv"
	"time"

	"github.com/kritagya03/ttk4145-project/internal/behaviour"
	"github.com/kritagya03/ttk4145-project/internal/call"
	"github.com/kritagya03/ttk4145-project/internal/worldview"
)

const (
	hallUpIndex   = 0
	hallDownIndex = 1
)

type InputFormat struct {
	HallCalls      [][]bool                 `json:"hallRequests"`
	ElevatorStates map[string]elevatorState `json:"states"`
}

type OutputFormat map[string][][]bool

type elevatorState struct {
	Behaviour        string `json:"behaviour"`
	FloorLastVisited int    `json:"floor"`
	Direction        string `json:"direction"`
	CabCalls         []bool `json:"cabRequests"`
}

func HallCalls(masterWorldview worldview.Master,
	allSlaveWorldviews []worldview.Slave,
	allSlavesOnline []bool,
	slaveWatchdogTimestamps []time.Time,
	durationUntilStuck time.Duration,
	hallRequestAssignerPath string,
	runExternalAssigner func(InputFormat, string) OutputFormat,
) worldview.Master {

	input := getInput(masterWorldview, allSlaveWorldviews, allSlavesOnline, slaveWatchdogTimestamps, durationUntilStuck)

	if len(input.ElevatorStates) == 0 {
		return masterWorldview
	}

	output := runExternalAssigner(input, hallRequestAssignerPath)
	masterWorldview = getMasterWorldviewWithOutput(masterWorldview, output)

	return masterWorldview
}

func getInput(
	masterWorldview worldview.Master,
	allSlaveWorldviews []worldview.Slave,
	allSlavesOnline []bool,
	slaveWatchdogTimestamps []time.Time,
	durationUntilStuck time.Duration,
) InputFormat {

	hallCallsMatrix := getHallCallMatrixFromMasterWorldview(masterWorldview)
	elevatorStatesMap := getElevatorStatesMap(masterWorldview, allSlaveWorldviews, allSlavesOnline, slaveWatchdogTimestamps, durationUntilStuck)

	input := InputFormat{
		HallCalls:      hallCallsMatrix,
		ElevatorStates: elevatorStatesMap,
	}

	return input
}

func getMasterWorldviewWithOutput(masterWorldview worldview.Master, output OutputFormat) worldview.Master {
	for networkIDString, hallCalls := range output {
		networkID, errorParsingNetworkID := strconv.Atoi(networkIDString)

		if errorParsingNetworkID != nil {
			panic(fmt.Sprintf("Failed to parse slave ID from HCA output: %v.", errorParsingNetworkID))
		}

		for floor := range hallCalls {
			for callIndex := range hallCalls[floor] {
				if hallCalls[floor][callIndex] {
					if callIndex == hallUpIndex || callIndex == hallDownIndex {
						masterWorldview.Calls[floor][callIndex] = call.GetAssignedTo(networkID)
					}
				}
			}
		}
	}
	return masterWorldview
}

func getHallCallMatrixFromMasterWorldview(masterWorldview worldview.Master) [][]bool {
	floorCount := len(masterWorldview.Calls)
	hallCallsMatrix := make([][]bool, floorCount)

	for floor := range floorCount {
		hallCallsMatrix[floor] = make([]bool, call.HallCallTypeCount)

		for hallIndex := range call.HallCallTypeCount {
			masterCall := masterWorldview.Calls[floor][hallIndex]

			if masterCall == call.Order || masterCall.IsAssignedToAnyone() {
				hallCallsMatrix[floor][hallIndex] = true
			} else {
				hallCallsMatrix[floor][hallIndex] = false
			}
		}
	}

	return hallCallsMatrix
}

func getElevatorStatesMap(
	masterWorldview worldview.Master,
	allSlaveWorldviews []worldview.Slave,
	allSlavesOnline []bool,
	slaveWatchdogTimestamps []time.Time,
	durationUntilStuck time.Duration,
) map[string]elevatorState {

	networkIDToState := make(map[string]elevatorState)

	behaviourToString := map[behaviour.State]string{
		behaviour.Idle:     "idle",
		behaviour.Moving:   "moving",
		behaviour.DoorOpen: "doorOpen",
	}

	directionToString := map[elevator.MotorDirection]string{
		elevator.MD_Up:   "up",
		elevator.MD_Down: "down",
		elevator.MD_Stop: "stop",
	}

	for _, slaveWorldview := range allSlaveWorldviews {
		slaveIndex := slaveWorldview.NetworkID - 1

		isSlaveOnline := allSlavesOnline[slaveIndex]

		isSlaveStuck := slaveWorldview.Behaviour != behaviour.Idle &&
			time.Since(slaveWatchdogTimestamps[slaveIndex]) > durationUntilStuck

		if isSlaveOnline && !isSlaveStuck {
			floorCount := len(masterWorldview.Calls)
			slaveCabCalls := make([]bool, floorCount)

			for floor := range floorCount {
				callIndex := call.GetCallIndex(elevator.BT_Cab, slaveWorldview.NetworkID)
				masterCall := masterWorldview.Calls[floor][callIndex]

				if masterCall == call.Order || masterCall.IsAssignedToAnyone() {
					slaveCabCalls[floor] = true
				} else {
					slaveCabCalls[floor] = false
				}
			}

			networkID := strconv.Itoa(slaveWorldview.NetworkID)

			networkIDToState[networkID] = elevatorState{
				Behaviour:        behaviourToString[slaveWorldview.Behaviour],
				FloorLastVisited: slaveWorldview.FloorLastVisited,
				Direction:        directionToString[slaveWorldview.Direction],
				CabCalls:         slaveCabCalls,
			}
		}
	}

	return networkIDToState
}
