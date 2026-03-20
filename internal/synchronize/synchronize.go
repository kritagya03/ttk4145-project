package synchronize

import (
	elevator "Driver-go/elevio"
	"time"

	"github.com/kritagya03/ttk4145-project/internal/assign"
	"github.com/kritagya03/ttk4145-project/internal/call"
	"github.com/kritagya03/ttk4145-project/internal/worldview"
)

func MasterWorldview(masterWorldview worldview.Master, slaveWorldview worldview.Slave, allSlaveWorldviews []worldview.Slave, allSlavesOnline []bool, slaveWatchdogTimestamps []time.Time, durationUntilStuck time.Duration, hallRequestAssignerPath string, runExternalAssigner func(assign.InputFormat, string) assign.OutputFormat) worldview.Master {
	for floor := range masterWorldview.Calls {
		for callIndex := range masterWorldview.Calls[floor] {
			masterCall := masterWorldview.Calls[floor][callIndex]
			slaveCall := slaveWorldview.Calls[floor][callIndex]

			if masterCall == call.None && slaveCall == call.Order {
				if call.GetButtonType(callIndex) == elevator.BT_Cab {
					masterWorldview.Calls[floor][callIndex] = call.GetAssignedTo(slaveWorldview.NetworkID)
				} else {
					masterWorldview.Calls[floor][callIndex] = call.Order
				}
			} else if masterCall.IsAssignedTo(slaveWorldview.NetworkID) && slaveCall == call.Completed {
				masterWorldview.Calls[floor][callIndex] = call.None
			}
		}
	}
	masterWorldview = assign.HallCalls(masterWorldview, allSlaveWorldviews, allSlavesOnline, slaveWatchdogTimestamps, durationUntilStuck, hallRequestAssignerPath, runExternalAssigner)
	return masterWorldview
}

func SlaveWorldview(slaveWorldview worldview.Slave, masterWorldview worldview.Master) worldview.Slave {
	for floor := range slaveWorldview.Calls {
		for callIndex := range slaveWorldview.Calls[floor] {
			masterCall := masterWorldview.Calls[floor][callIndex]
			slaveCall := slaveWorldview.Calls[floor][callIndex]

			if masterCall != slaveCall {
				if masterCall == call.None {
					if slaveCall == call.Completed || slaveCall.IsAssignedToAnyone() {
						slaveWorldview.Calls[floor][callIndex] = masterCall
					}
				} else if masterCall.IsAssignedToAnyone() {
					if slaveCall == call.None || slaveCall == call.Order || slaveCall.IsAssignedToAnyone() {
						slaveWorldview.Calls[floor][callIndex] = masterCall
					}
				}
			}

		}
	}

	return slaveWorldview
}
