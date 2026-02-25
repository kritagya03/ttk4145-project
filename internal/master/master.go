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
	MasterMerging // TODO: maybe change name because also merging master worldviews if not in this state
)

// TODO: Make the state machine more beautiful (good code quality, keep events on the outside and state on the inside)

func Server(masterNetworkEvents <-chan interface{}, masterNetworkCommands chan<- MasterWorldview, networkID int, floorCount int, buttonTypeCount int, elevatorCount int) {
	masterWorldview := getDefaultMasterWorldview(networkID, floorCount, buttonTypeCount)
	masterState := MasterCandidate

	slaveWorldviewList := make([]SlaveWorldview, elevatorCount) // TODO: maybe change variable name
	slaveOnlineList := make([]bool, elevatorCount)              // TODO: maybe change variable name
	for slaveIndex := range slaveWorldviewList {
		slaveWorldviewList[slaveIndex] = getDefaultSlaveWorldview(slaveIndex + 1)
		slaveOnlineList[slaveIndex] = false
	}

	electionTimeout := time.NewTimer(BaseElectionTimeout * time.Duration(networkID))
	electionTimeout.Stop()

	mergingMastersTimeout := time.NewTimer(MergingMastersTimeoutDuration)
	mergingMastersTimeout.Stop()

	heartbeatTicker := time.NewTicker(HeartbeatInterval)
	defer heartbeatTicker.Stop()

	// Testing
	// masterWorldview.Calls.Matrix[0][0] = CallStateOrder                                               // testing
	// masterWorldview.Calls.Matrix[1][1] = CallStateOrder                                               // testing
	// masterWorldview.Calls.Matrix[2][3] = CallStateOrder                                               // testing
	// fmt.Println("MasterWorldview after setting some orders:", masterWorldview)                        // testing
	// assignCalls(masterWorldview, slaveWorldviewList, slaveOnlineList, elevatorCount, floorCount)          // testing

	for {
		select {
		case message := <-masterNetworkEvents:
			switch event := message.(type) {
			case SlaveWorldview:
				slaveWorldview := event
				fmt.Printf("master.go case masterNetworkEvents. Received SlaveWorldview: %v\n", slaveWorldview)
				if masterState == MasterActive {
					slaveWorldviewList[slaveWorldview.NetworkID-1] = slaveWorldview
					masterWorldview = getNewMasterWorldview(masterWorldview, slaveWorldview, slaveWorldviewList, slaveOnlineList)
				}

			case MasterWorldview:
				receivedMasterWorldview := event
				fmt.Printf("master.go case masterNetworkEvents. Received MasterWorldview: %v\n", receivedMasterWorldview)
				masterWorldview = getMergedMasterWorldview(masterWorldview, receivedMasterWorldview, elevatorCount)

			case NewMasterConnection:
				fmt.Printf("master.go case masterNetworkEvents. Received New Master Connection: %d\n", event)
				switch masterState {
				case MasterCandidate:
					fmt.Println("master.go case masterNetworkEvents. Received New Master Connection while in Candidate state.")
					masterState = MasterInactive
				case MasterActive:
					fmt.Println("master.go case masterNetworkEvents. Received New Master Connection while in Active state.")
					masterState = MasterMerging
					resetTimer(mergingMastersTimeout, MergingMastersTimeoutDuration)
				case MasterMerging:
					fmt.Println("master.go case masterNetworkEvents. Received New Master Connection while in Merging state")
					resetTimer(mergingMastersTimeout, MergingMastersTimeoutDuration)
				case MasterInactive:
					fmt.Println("master.go case masterNetworkEvents. Received New Master Connection while in Inactive state")
				default:
					panic(fmt.Sprintf("master.go case masterNetworkEvents. Received unknown event type: %T, value: %v", event, event))
				}

			case NewSlaveConnection:
				slaveConnection := event
				fmt.Printf("master.go case masterNetworkEvents. Received New Slave Connection with network ID: %d\n", slaveConnection.NetworkID)
				slaveOnlineList[slaveConnection.NetworkID-1] = true

			case MasterTimeout:
				fmt.Println("master.go case masterNetworkEvents. Received Master Timeout.")
				if masterState != MasterActive {
					masterState = MasterCandidate
					resetTimer(electionTimeout, BaseElectionTimeout*time.Duration(networkID)) // TODO: Currently resetting this timer in two locations, maybe only need to reset in one location.
				}

			case SlaveTimeout:
				slaveTimeout := event
				fmt.Printf("master.go case masterNetworkEvents. Received Slave Timeout with Network ID: %d\n", slaveTimeout.NetworkID)
				slaveOnlineList[slaveTimeout.NetworkID-1] = false

			default:
				panic(fmt.Sprintf("master.go case masterNetworkEvents. Unknown type: %T\n", event))
			}
		case <-heartbeatTicker.C:
			if masterState == MasterActive || masterState == MasterMerging {
				fmt.Println("Master heartbeat. Current MasterWorldview:", masterWorldview)
				masterNetworkCommands <- masterWorldview
			}
		case <-electionTimeout.C:
			if masterState == MasterCandidate {
				fmt.Println("Election timeout. Transitioning to MasterActive state.")
				masterState = MasterActive
			}
		case <-mergingMastersTimeout.C:
			if masterState == MasterMerging {
				fmt.Println("Merging Masters timeout. Transitioning to MasterCandidate state.")
				masterState = MasterCandidate
				resetTimer(electionTimeout, BaseElectionTimeout*time.Duration(networkID)) // TODO: Currently resetting this timer in two locations, maybe only need to reset in one location.
			}
		}
	}
}

// TODO: weird to have elevatorCount as an argument to the function.
func getMergedMasterWorldview(masterWorldviewBase MasterWorldview, masterWorldviewNew MasterWorldview, elevatorCount int) MasterWorldview {
	masterWorldviewMerged := masterWorldviewBase
	matrixMerged := masterWorldviewMerged.Calls.Matrix
	matrixNew := masterWorldviewNew.Calls.Matrix
	// TODO: currently assuming both matrixes have the same dimensions
	for floor := range matrixNew {
		for buttonType := range matrixNew[floor] {
			if matrixMerged[floor][buttonType] == CallStateNone && isCallAssignedToSlave(matrixNew[floor][buttonType], elevatorCount) {
				matrixMerged[floor][buttonType] = matrixNew[floor][buttonType]
			}
		}
	}
	masterWorldviewMerged.Calls.Matrix = matrixMerged
	return masterWorldviewMerged
}

func getDefaultMasterWorldview(networkID int, floorCount int, buttonTypeCount int) MasterWorldview {
	calls := make([][]CallState, floorCount)
	for floor := range floorCount {
		calls[floor] = make([]CallState, buttonTypeCount)
		for buttonType := range buttonTypeCount {
			calls[floor][buttonType] = CallStateNone
		}
	}
	return MasterWorldview{
		NetworkID: networkID,
		Calls:     CallsMatrix{Matrix: calls},
	}
}

func getDefaultSlaveWorldview(networkID int) SlaveWorldview {
	return SlaveWorldview{
		NetworkID:        networkID,
		FloorLastVisited: -1,
		Direction:        DirectionStop,
		Behaviour:        BehaviourIdle,
	}
}

func isCallAssignedToSlave(callState CallState, elevatorCount int) bool {
	if int(callState) > 0 && int(callState) <= elevatorCount {
		return true
	}
	return false
}

// TODO: assumes master matrix and slave matrix are of the same dimensions
func getNewMasterWorldview(masterWorldview MasterWorldview, slaveWorldview SlaveWorldview, slaveWorldviewList []SlaveWorldview, slaveOnlineList []bool) MasterWorldview {
	elevatorCount := len(slaveWorldviewList)
	masterMatrix := masterWorldview.Calls.Matrix
	slaveMatrix := slaveWorldview.Calls.Matrix
	for floor := range masterMatrix {
		for buttonType := range masterMatrix[floor] {
			if masterMatrix[floor][buttonType] == CallStateNone && slaveMatrix[floor][buttonType] == CallStateOrder {
				masterMatrix[floor][buttonType] = CallStateOrder
			} else if isCallAssignedToSlave(masterMatrix[floor][buttonType], elevatorCount) &&
				slaveMatrix[floor][buttonType] == CallStateCompleted &&
				masterMatrix[floor][buttonType] == CallState(slaveWorldview.NetworkID) {

				masterMatrix[floor][buttonType] = CallStateNone
			}
		}
	}
	masterWorldview.Calls.Matrix = masterMatrix
	assignCalls(masterWorldview, slaveWorldviewList, slaveOnlineList)
	return masterWorldview
}

type HallCallAssignerInput struct {
	HallCalls      [][]bool                                 `json:"hallRequests"`
	ElevatorStates map[string]HallCallAssignerElevatorState `json:"states"`
}

type HallCallAssignerElevatorState struct {
	Behaviour        string `json:"behaviour"`
	FloorLastVisited int    `json:"floor"`
	Direction        string `json:"direction"`
	CabCalls         []bool `json:"cabRequests"`
}

type HallCallAssignerOutput map[string][][]bool

// WE ARE HERE!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

func assignCalls(masterWorldview MasterWorldview, slaveWorldviewList []SlaveWorldview, slaveOnlineList []bool) MasterWorldview {
	floorCount := len(masterWorldview.Calls.Matrix)
	elevatorCount := len(slaveWorldviewList)

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

	elevatorStatesMap := make(map[string]HallCallAssignerElevatorState)
	for _, slaveWorldview := range slaveWorldviewList {
		if slaveOnlineList[slaveWorldview.NetworkID-1] == false {
			continue
		}

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

		elevatorID := fmt.Sprintf("%d", slaveWorldview.NetworkID)

		elevatorStatesMap[elevatorID] = HallCallAssignerElevatorState{
			Behaviour:        behaviourTextMapping[slaveWorldview.Behaviour],
			FloorLastVisited: slaveWorldview.FloorLastVisited,
			Direction:        directionTextMapping[slaveWorldview.Direction],
			CabCalls:         cabCalls,
		}
	}

	if len(elevatorStatesMap) == 0 {
		fmt.Println("Don't assign calls because no connected slaves.")
		return masterWorldview
	}

	fmt.Println("Elevator states map prepared for HCA input:", elevatorStatesMap)

	input := HallCallAssignerInput{
		HallCalls:      hallCallsMatrix,
		ElevatorStates: elevatorStatesMap,
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
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Run command
	if errorRunningCommand := cmd.Run(); errorRunningCommand != nil {
		fmt.Printf(
			"HCA failed: %v\nstderr: %s\n",
			errorRunningCommand,
			stderr.String(),
		)
		panic(fmt.Sprintf(
			"HCA failed: %v\nstderr: %s",
			errorRunningCommand,
			stderr.String(),
		))
	}

	// /home/student/Desktop/KrittErKuL/ttk4145-project/bin/hall_call_assigner --input '{"hallRequests":[[false,true],[false,false],[false,false],[true,false]],"states":{"elevator0":{"behaviour":"idle","floor":1,"direction":"down","cabRequests":[false,true,false,false]},"elevator1":{"behaviour":"idle","floor":3,"direction":"up","cabRequests":[true,false,false,false]}}}'

	// Parse output JSON
	var output HallCallAssignerOutput
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
