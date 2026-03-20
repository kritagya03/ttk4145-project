package master

import (
	"fmt"
	"time"

	"github.com/kritagya03/ttk4145-project/internal/assign"
	"github.com/kritagya03/ttk4145-project/internal/connection"
	"github.com/kritagya03/ttk4145-project/internal/messages"
	"github.com/kritagya03/ttk4145-project/internal/synchronize"
	"github.com/kritagya03/ttk4145-project/internal/timer"
	"github.com/kritagya03/ttk4145-project/internal/worldview"
)

type masterState int

const (
	inactive masterState = iota
	candidate
	active
	merging
)

func Server(networkToMaster <-chan messages.NetworkToMaster, masterToNetwork chan<- worldview.Master,
	networkID int, floorCount int, callTypeCount int, elevatorCount int, heartbeatInterval time.Duration, baseElectionDuration time.Duration, mergingMastersDuration time.Duration, durationUntilStuck time.Duration, hallRequestAssignerPath string) {

	masterWorldview := worldview.NewMaster(networkID, floorCount, callTypeCount)
	masterState := candidate

	electionDuration := baseElectionDuration * time.Duration(networkID)

	allSlaveWorldviews := make([]worldview.Slave, elevatorCount)
	allSlavesOnline := make([]bool, elevatorCount)
	slaveWatchdogTimestamps := make([]time.Time, elevatorCount)
	for slaveIndex := range allSlaveWorldviews {
		allSlaveWorldviews[slaveIndex] = worldview.NewSlave(slaveIndex+1, floorCount, callTypeCount)
		allSlavesOnline[slaveIndex] = false
		slaveWatchdogTimestamps[slaveIndex] = time.Now()
	}

	heartbeatTicker := time.NewTicker(heartbeatInterval)
	defer heartbeatTicker.Stop()
	mergingMastersTimeout := time.NewTimer(mergingMastersDuration)
	mergingMastersTimeout.Stop()

	electionTimeout := time.NewTimer(electionDuration)

	fmt.Println("The master started as Candidate.")

	for {
		select {
		case message := <-networkToMaster:
			switch event := message.(type) {
			case worldview.Slave:
				receivedSlaveWorldview := event
				slaveIndex := receivedSlaveWorldview.NetworkID - 1
				priorSlaveWorldview := allSlaveWorldviews[slaveIndex]

				if priorSlaveWorldview.Behaviour != receivedSlaveWorldview.Behaviour ||
					priorSlaveWorldview.FloorLastVisited != receivedSlaveWorldview.FloorLastVisited ||
					priorSlaveWorldview.Direction != receivedSlaveWorldview.Direction {

					slaveWatchdogTimestamps[slaveIndex] = time.Now()
				}

				allSlaveWorldviews[slaveIndex] = receivedSlaveWorldview
				if masterState == active {
					masterWorldview = synchronize.MasterWorldview(masterWorldview, receivedSlaveWorldview, allSlaveWorldviews, allSlavesOnline, slaveWatchdogTimestamps, durationUntilStuck, hallRequestAssignerPath, assign.RunExternalAssigner)
				}

			case worldview.Master:
				receivedMasterWorldview := event
				if receivedMasterWorldview.NetworkID != networkID {
					switch masterState {
					case active:
						masterState = merging
						timer.Reset(mergingMastersTimeout, mergingMastersDuration)
						fmt.Println("Received MasterWorldview while Active. Setting to Merging.")

					case candidate:
						masterState = inactive
						fmt.Println("Received MasterWorldview while Candidate. Setting to Inactive.")

					case inactive:
						masterWorldview.Calls = receivedMasterWorldview.Calls

					case merging:
						masterWorldview = masterWorldview.MergedWith(receivedMasterWorldview)
					}
				}

			case connection.NewSlave:
				slaveConnection := event
				slaveIndex := slaveConnection.NetworkID - 1
				allSlavesOnline[slaveIndex] = true
				slaveWatchdogTimestamps[slaveIndex] = time.Now()

				fmt.Printf("A new slave connected with id %d.\n", slaveConnection.NetworkID)

			case connection.MasterTimeout:
				if masterState != active {
					fmt.Println("The master has timed out while not Active. Setting to Candidate.")
					masterState = candidate
					timer.Reset(electionTimeout, electionDuration)
				}

			case connection.SlaveTimeout:
				slaveTimeout := event
				slaveIndex := slaveTimeout.NetworkID - 1
				allSlavesOnline[slaveIndex] = false

				fmt.Printf("A slave has timed out with id %d.\n", slaveTimeout.NetworkID)

			default:
				panic(fmt.Sprintf("The master received a message of unknown type %T with value %v.\n", event, event))
			}

		case <-heartbeatTicker.C:
			if masterState == active || masterState == merging {
				masterToNetwork <- masterWorldview.DeepCopy()
			}

		case <-electionTimeout.C:
			if masterState == candidate {
				masterState = active

				fmt.Println("Election timeout while Candidate. Setting to Active.")
			}
		case <-mergingMastersTimeout.C:
			if masterState == merging {
				masterState = candidate
				timer.Reset(electionTimeout, electionDuration)

				fmt.Println("Merging masters timer has timed out while Merging. Setting to Candidate.")
			}
		}
	}
}
