package master

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	. "github.com/kritagya03/ttk4145-project/internal/models"
)

type MasterState int

const (
	MasterInactive MasterState = iota
	MasterCandidate
	MasterActive
)

// States: Inactive, Candidate, Active

// Add periodically sending our MasterWorldview to the network

func Server(masterNetworkEvents <-chan interface{}, masterNetworkCommands chan<- MasterWorldview, networkID int, floorCount int, buttonTypeCount int, elevatorCount int) {
	masterWorldview := getDefaultMasterWorldview(floorCount, buttonTypeCount)
	slaveWorldviewList := make(SlaveWorldview, elevatorCount)
	currentMasterState := MasterCandidate
	changeMasterState(currentMasterState)
	for {
		event := <-masterNetworkEvents
		switch v := event.(type) {
		case SlaveWorldview:
			// var slaveWorldview SlaveWorldview = event
			slaveWorldview := event
			fmt.Printf("master.go case masterNetworkEvents. Received SlaveWorldview: %v\n", slaveWorldview)
			if currentMasterState != MasterActive {
				continue
			}
			slaveWorldviewList[slaveWorldview.NetworkID-1] = slaveWorldview
			masterWorldview = getNewMasterWorldview(masterWorldview, slaveWorldview, elevatorCount)
			// Maybe call assignCalls from inside getNewMasterWorldview such that the master doen't accidentally send a call marked CallOrder with its heartbeat
			masterWorldview = assignCalls(masterWorldview, slaveWorldviewList, elevatorCount, floorCount)
			// From the SlaveWorldview, proccess calls that are marked as Order and Completed, and update the MasterWorldview accordingly. Then supdate our MasterWorldview
		case MasterTimeout:
			fmt.Printf("master.go case masterNetworkEvents. Received Master Timeout: %d\n", v)
			if currentMasterState == MasterActive {
				continue
			}
			changeMasterState(MasterCandidate)
		case SlaveTimeout:
			slaveID := event
			fmt.Printf("master.go case masterNetworkEvents. Received Slave Timeout: %d\n", v)
			if currentMasterState != MasterActive {
				continue
			}
		default:
			fmt.Printf("master.go case masterNetworkEvents. Unknown type: %T\n", v)
		}
	}
}

// Maybe change from all calls default being 0 (None)
func getDefaultMasterWorldview(floorCount int, buttonTypeCount int) MasterWorldview {
	calls := make([][]CallState, floorCount)
	for i := range calls {
		calls[i] = make([]CallState, buttonTypeCount)
	}
	return MasterWorldview{
		Calls: CallsMatrix{Matrix: calls},
	}
}

func changeMasterState(newState MasterState) {
	fmt.Printf("Changing master state to: %v\n", newState)
	// Implement any additional logic needed when changing states, such as resetting timers or clearing data.
}

func isCallAssignedToSlave(callState CallState, elevatorCount int) bool {
	if int(callState) > 0 && int(callState) <= elevatorCount {
		return true
	}
	return false
}

// Assumes master matrix and slave matrix are of the same dimensions
//Change name to UpdateMasterWorldview
func getNewMasterWorldview(masterWorldview MasterWorldview, slaveWorldview SlaveWorldview, elevatorCount int) MasterWorldview {
	masterMatrix := masterWorldview.Calls.Matrix
	slaveMatrix := slaveWorldview.Calls.Matrix
	for floor := range masterMatrix {
		for buttonType := range masterMatrix[floor] {
			if masterMatrix[floor][buttonType] == CallStateNone && slaveMatrix[floor][buttonType] == CallStateOrder {
				masterMatrix[floor][buttonType] = CallStateOrder // Hall Request Assigner automatically assigns the order to a slave
			} else if isCallAssignedToSlave(masterMatrix[floor][buttonType], elevatorCount) && slaveMatrix[floor][buttonType] == CallStateCompleted {
				// Any slave can mark a call as completed, even if the slave was not assigned the order
				masterMatrix[floor][buttonType] = CallStateNone
			}
		}
	}
	masterWorldview.Calls.Matrix = masterMatrix
	return masterWorldview
}

const hallCallAssignerPath = "bin/hall_call_assigner"

type HCAInput struct {
	HallCalls [][]bool                    `json:"hallRequests"`
	States    map[string]HCAElevatorState `json:"states"`
}

type HCAElevatorState struct {
	Behaviour        string `json:"behaviour"`
	FloorLastVisited int    `json:"floor"`
	Direction        string `json:"direction"`
	CabCalls         []bool `json:"cabRequests"`
}

type HCAOutput map[string][][]bool

func assignCalls(masterWorldview MasterWorldview, slaveWorldviewList []SlaveWorldview, elevatorCount int, floorCount int) MasterWorldview {
	hallCallsMatrix := make([][]bool, floorCount)
	for floor := range floorCount {
		hallCallsMatrix[floor] = make([]bool, 2) // TODO: Maybe don't hardcode 2, but rather have a const for the number of hall button types
		for buttonType := range 2 {              // TODO: Maybe don't hardcode 2, but rather have a const for the number of hall button types
			if masterWorldview.Calls.Matrix[floor][buttonType] == CallStateOrder || isCallAssignedToSlave(masterWorldview.Calls.Matrix[floor][buttonType], elevatorCount) {
				hallCallsMatrix[floor][buttonType] = true
			} else {
				hallCallsMatrix[floor][buttonType] = false
			}
		}
	}

	elevatorStatesMap := make(map[string]HCAElevatorState)
	for _, slaveWorldview := range slaveWorldviewList {
		elevatorID := fmt.Sprintf("elevator%d", slaveWorldview.NetworkID)
		cabCalls := make([]bool, floorCount)

		for floor := range floorCount {
			buttonType := 2 + slaveWorldview.NetworkID - 1 // TODO: Maybe not hardcode 2. Cab calls start after the hall calls in the button type indexing.
			if masterWorldview.Calls.Matrix[floor][buttonType] == CallStateOrder || isCallAssignedToSlave(masterWorldview.Calls.Matrix[floor][buttonType], elevatorCount) {
				cabCalls[floor] = true
			} else {
				cabCalls[floor] = false
			}
		}

		elevatorStatesMap[elevatorID] = HCAElevatorState{
			Behaviour:        fmt.Sprintf("%d", slaveWorldview.Behaviour),
			FloorLastVisited: slaveWorldview.FloorLastVisited,
			Direction:        fmt.Sprintf("%d", slaveWorldview.Direction),
			CabCalls:         cabCalls,
		}
	}

	input := HCAInput{
		HallCalls: hallCallsMatrix,
		States:    elevatorStatesMap,
	}

	jsonInput, errorJSONEncoding := json.Marshal(input)
	if errorJSONEncoding != nil {
		panic(fmt.Sprintf("Failed to marshal hall call assigner input: %v", errorJSONEncoding))
	}

	// Prepare command
	// TODO: Maybe add command line parameters: travelDuration, doorOpenDuration, clearRequestType, includeCab
	// Command line arguments: https://github.com/TTK4145/Project-resources/tree/master/cost_fns/hall_request_assigner#command-line-arguments
	cmd := exec.Command(hallCallAssignerPath, "--input", string(jsonInput))

	// Capture stdout
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	// Capture stderr (VERY useful for debugging)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Run command
	if errorRunningCommand := cmd.Run(); errorRunningCommand != nil {
		panic(fmt.Sprintf(
			"HCA failed: %v\nstderr: %s",
			errorRunningCommand,
			stderr.String(),
		))
	}

	// Parse output JSON
	var output HCAOutput
	if errorJSONDecoring := json.Unmarshal(stdout.Bytes(), &output); errorJSONDecoring != nil {
		panic(fmt.Sprintf("Failed to parse hall call assigner output: %v", errorJSONDecoring))
	}

	// Convert HCA output to MasterWorldview format
	// return output
}
