package master

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	. "github.com/kritagya03/ttk4145-project/internal/models"
)

type MasterState int

const (
	MasterInactive MasterState = iota
	MasterCandidate
	MasterActive
	MasterCombine
)

// TODO: Make the state machine more beautiful (good code quality, keep events on the outside and state on the inside)

func Server(masterNetworkEvents <-chan interface{}, masterNetworkCommands chan<- MasterWorldview, networkID int, floorCount int, buttonTypeCount int, elevatorCount int) {
	masterWorldview := getDefaultMasterWorldview(floorCount, buttonTypeCount)
	fmt.Println("Initial MasterWorldview:", masterWorldview)
	slaveWorldviewList := make([]SlaveWorldview, elevatorCount)
	fmt.Println("Initial SlaveWorldview list:", slaveWorldviewList)
	for i := range slaveWorldviewList {
		slaveWorldviewList[i] = getDefaultSlaveWorldview(i + 1)
	}
	fmt.Println("SlaveWorldview list after setting initial values:", slaveWorldviewList)
	masterState := MasterCandidate // testing
	applyMasterState(masterState)
	masterWorldview.Calls.Matrix[0][0] = CallStateOrder                         // testing
	masterWorldview.Calls.Matrix[1][1] = CallStateOrder                         // testing
	masterWorldview.Calls.Matrix[2][3] = CallStateOrder                         // testing
	fmt.Println("MasterWorldview after setting some orders:", masterWorldview)  // testing
	assignCalls(masterWorldview, slaveWorldviewList, elevatorCount, floorCount) // testing
	// Staggered election timeout
	electionTimeout := time.NewTimer(BaseElectionTimeout * time.Duration(networkID))
	electionTimeout.Stop()
	combineMastersTimeout := time.NewTimer(CombineMastersTimeoutDuration)
	combineMastersTimeout.Stop()
	heartbeatTicker := time.NewTicker(HeartbeatInterval)
	defer heartbeatTicker.Stop()
	for {
		select {
		case <-heartbeatTicker.C:
			if masterState == MasterActive {
				fmt.Println("Master heartbeat. Current MasterWorldview:", masterWorldview)
				masterNetworkCommands <- masterWorldview
			}
		case <-electionTimeout.C:
			if masterState != MasterCandidate {
				continue
			}
			fmt.Println("Election timeout. Transitioning to MasterActive state.")
			masterState = MasterActive
			applyMasterState(masterState)
		case <-combineMastersTimeout.C:
			fmt.Println("Combine Masters timeout. Transitioning to MasterInactive state.")
			masterState = MasterCandidate
			resetTimer(electionTimeout, BaseElectionTimeout*time.Duration(networkID)) // TODO: maybe move to applyMasterState?
			applyMasterState(masterState)
		case message := <-masterNetworkEvents:
			switch event := message.(type) {
			case SlaveWorldview:
				slaveWorldview := event
				fmt.Printf("master.go case masterNetworkEvents. Received SlaveWorldview: %v\n", slaveWorldview)
				if masterState != MasterActive {
					continue
				}
				slaveWorldviewList[slaveWorldview.NetworkID-1] = slaveWorldview
				masterWorldview = getNewMasterWorldview(masterWorldview, slaveWorldview, slaveWorldviewList, elevatorCount, floorCount)
			case MasterWorldview: // Combine master worldviews after network partition (in MasterCombine state)
				// worldview := event
				// fmt.Printf("master.go case masterNetworkEvents. Received MasterWorldview: %v\n", worldview)
				// // If we receive a MasterWorldview while we are in Candidate state, that means we have won the election and can transition to Active. If we receive a MasterWorldview while we are in Active state, that means there is another master in the network and we should transition to Combine. If we receive a MasterWorldview while we are in Combine state, that means there is still another master in the network, but we can stay in Combine state and just update our worldview. If we receive a MasterWorldview while we are in Inactive state, that means there is a master in the network, but since we are inactive, we can just stay in Inactive state and update our worldview.
				// if masterState == MasterCandidate {
				// 	fmt.Println("master.go case masterNetworkEvents. Received MasterWorldview while in Candidate state, transitioning to Active.")
				// 	masterState = MasterActive
				// 	applyMasterState(masterState)
				// } else if masterState == MasterActive {
				// 	fmt.Println("master.go case masterNetworkEvents. Received MasterWorldview while in Active state, transitioning to Combine.")
				// 	masterState = MasterCombine
				// 	applyMasterState(masterState)
				// } else if masterState == MasterCombine {
				// 	fmt.Printf("master.go case masterNetworkEvents. Received Master Combine event: %v\n", event)
				// } else if masterState == MasterInactive {
				// 	fmt.Printf("master.go case masterNetworkEvents. Received Master Inactive event: %v\n", event)
				// } else {
				// 	panic(fmt.Sprintf("master.go case masterNetworkEvents. Received unknown event type: %T, value: %v", event, event))
				// }
			case NewMasterConnection:
				fmt.Printf("master.go case masterNetworkEvents. Received New Master Connection: %d\n", event)
				switch masterState {
				case MasterCandidate:
					fmt.Println("master.go case masterNetworkEvents. Received New Master Connection while in Candidate state, transitioning to Active.")
					masterState = MasterInactive
					applyMasterState(masterState)
				case MasterActive:
					fmt.Println("master.go case masterNetworkEvents. Received New Master Connection while in Active state, transitioning to Inactive.")
					masterState = MasterCombine
					applyMasterState(masterState)
					resetTimer(combineMastersTimeout, CombineMastersTimeoutDuration)
				case MasterCombine:
					fmt.Printf("master.go case masterNetworkEvents. Received Master Combine event: %d\n", event)
				case MasterInactive:
					fmt.Printf("master.go case masterNetworkEvents. Received Master Inactive event: %d\n", event)
				default:
					panic(fmt.Sprintf("master.go case masterNetworkEvents. Received unknown event type: %T, value: %v", event, event))
				}
			case NewSlaveConnection:
				fmt.Printf("master.go case masterNetworkEvents. Received New Slave Connection: %d\n", event)
			case MasterTimeout:
				fmt.Printf("master.go case masterNetworkEvents. Received Master Timeout: %d\n", event)
				if masterState == MasterActive {
					// WHY DO WE CHECK IF MASTER? If we receive a master timeout, doesn't that mean we should transition to candidate regardless of our current state? Maybe we should only ignore the master timeout if we are already in candidate state, since that would mean we have already transitioned to candidate and are waiting for a new master worldview to transition to active?
					continue
				}
				masterState = MasterCandidate // TODO: maybe move to applyMasterState?
				resetTimer(electionTimeout, BaseElectionTimeout*time.Duration(networkID))
				applyMasterState(masterState)
			case SlaveTimeout:
				slaveID := event
				fmt.Printf("master.go case masterNetworkEvents. Received Slave Timeout: %d\n", slaveID)
				if masterState != MasterActive {
					continue
				}
			default:
				panic(fmt.Sprintf("master.go case masterNetworkEvents. Unknown type: %T\n", event))
			}
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

func getDefaultSlaveWorldview(networkID int) SlaveWorldview {
	return SlaveWorldview{
		NetworkID:        networkID,
		FloorLastVisited: 0,
		Direction:        DirectionStop,
		Behaviour:        BehaviourIdle,
	}
}

func applyMasterState(newState MasterState) {
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
func getNewMasterWorldview(masterWorldview MasterWorldview, slaveWorldview SlaveWorldview, slaveWorldviewList []SlaveWorldview, elevatorCount int, floorCount int) MasterWorldview {
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
	assignCalls(masterWorldview, slaveWorldviewList, elevatorCount, floorCount)
	return masterWorldview
}

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

	fmt.Println("Hall calls matrix prepared for HCA input:", hallCallsMatrix)

	elevatorStatesMap := make(map[string]HCAElevatorState)
	for _, slaveWorldview := range slaveWorldviewList {
		elevatorID := fmt.Sprintf("%d", slaveWorldview.NetworkID)
		cabCalls := make([]bool, floorCount)

		for floor := range floorCount {
			buttonType := 2 + slaveWorldview.NetworkID - 1 // TODO: Maybe not hardcode 2. Cab calls start after the hall calls in the button type indexing.
			if masterWorldview.Calls.Matrix[floor][buttonType] == CallStateOrder || isCallAssignedToSlave(masterWorldview.Calls.Matrix[floor][buttonType], elevatorCount) {
				cabCalls[floor] = true
			} else {
				cabCalls[floor] = false
			}
		}

		behaviourTextMapping := map[ElevatorBehaviour]string{
			BehaviourIdle:     "idle",
			BehaviourMoving:   "moving",
			BehaviourDoorOpen: "doorOpen",
		}
		directionTextMapping := map[ElevatorDirection]string{
			DirectionUp:   "up",
			DirectionDown: "down",
			DirectionStop: "stop",
		}

		elevatorStatesMap[elevatorID] = HCAElevatorState{
			Behaviour:        behaviourTextMapping[slaveWorldview.Behaviour],
			FloorLastVisited: slaveWorldview.FloorLastVisited,
			Direction:        directionTextMapping[slaveWorldview.Direction],
			CabCalls:         cabCalls,
		}
	}

	fmt.Println("Elevator states map prepared for HCA input:", elevatorStatesMap)

	input := HCAInput{
		HallCalls: hallCallsMatrix,
		States:    elevatorStatesMap,
	}

	jsonInput, errorJSONEncoding := json.Marshal(input)
	if errorJSONEncoding != nil {
		panic(fmt.Sprintf("Failed to marshal hall call assigner input: %v", errorJSONEncoding))
	}

	// Prepare command
	fmt.Println("Running hall call assigner with input:", string(jsonInput))
	workingDirectory, _ := os.Getwd()
	hallCallAssignerPath := filepath.Join(workingDirectory, "..", "..", "bin", "hall_call_assigner")
	// TODO: Maybe add command line parameters: travelDuration, doorOpenDuration, clearRequestType, includeCab
	// Command line arguments: https://github.com/TTK4145/Project-resources/tree/master/cost_fns/hall_request_assigner#command-line-arguments
	cmd := exec.Command(hallCallAssignerPath, "--input", string(jsonInput))

	// Capture stdout
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	// Capture stderr (debugging)
	// var stderr bytes.Buffer
	// cmd.Stderr = &stderr

	// Run command
	if errorRunningCommand := cmd.Run(); errorRunningCommand != nil {
		fmt.Printf(
			"HCA failed: %v\nstderr: %s\n",
			errorRunningCommand,
			// stderr.String(),
		)
		panic(fmt.Sprintf(
			"HCA failed: %v\nstderr: %s",
			errorRunningCommand,
			// stderr.String(),
		))
	}

	// /home/student/Desktop/KrittErKuL/ttk4145-project/bin/hall_call_assigner --input '{"hallRequests":[[false,true],[false,false],[false,false],[true,false]],"states":{"elevator0":{"behaviour":"idle","floor":1,"direction":"down","cabRequests":[false,true,false,false]},"elevator1":{"behaviour":"idle","floor":3,"direction":"up","cabRequests":[true,false,false,false]}}}'

	// Parse output JSON
	var output HCAOutput
	if errorJSONDecoring := json.Unmarshal(stdout.Bytes(), &output); errorJSONDecoring != nil {
		panic(fmt.Sprintf("Failed to parse hall call assigner output: %v", errorJSONDecoring))
	}

	// Convert HCA output to MasterWorldview format
	fmt.Println("HCA output:", output)
	for networkIDString, assignedCalls := range output {
		networkID, errorParsingSlaveID := strconv.Atoi(networkIDString)
		if errorParsingSlaveID != nil {
			panic(fmt.Sprintf("Failed to parse slave ID from HCA output: %v", errorParsingSlaveID))
		}
		for floor := range assignedCalls {
			for buttonType := range assignedCalls[floor] {
				assignedCallHallUpIndex := 0
				assignedCallHallDownIndex := 1
				assignedCallCabCallIndex := 2
				if assignedCalls[floor][buttonType] {
					switch buttonType {
					case assignedCallHallUpIndex, assignedCallHallDownIndex:
						masterWorldview.Calls.Matrix[floor][buttonType] = CallState(networkID)
					case assignedCallCabCallIndex:
						masterMatrixCabCallIndex := 2 + networkID - 1 // TODO: Maybe not hardcode 2. Cab calls start after the hall calls in the button type indexing.
						masterWorldview.Calls.Matrix[floor][masterMatrixCabCallIndex] = CallState(networkID)
					default:
						panic(fmt.Sprintf("Invalid button type index from HCA output: %d", buttonType))
					}
				}
			}
		}
	}

	fmt.Println("Updated MasterWorldview after hall call assignment:", masterWorldview)

	return masterWorldview
}

// TODO: This is reused from network_server.go
func resetTimer(timer *time.Timer, duration time.Duration) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(duration)
}
