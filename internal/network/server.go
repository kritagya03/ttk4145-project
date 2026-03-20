package network

import (
	broadcast "Network-go/network/bcast"
	"time"

	"github.com/kritagya03/ttk4145-project/internal/connection"
	"github.com/kritagya03/ttk4145-project/internal/messages"
	"github.com/kritagya03/ttk4145-project/internal/worldview"
)

func Server(masterToNetwork <-chan worldview.Master, slaveToNetwork <-chan worldview.Slave,
	networkToMaster chan<- messages.NetworkToMaster, networkToSlave chan<- worldview.Master,
	networkPort int, networkID int, elevatorCount int, heartbeatInterval time.Duration, timeoutDuration time.Duration) {

	broadcastTransmitMasterWorldview := make(chan worldview.Master, 1024)
	broadcastReceiveMasterWorldview := make(chan worldview.Master, 1024)
	broadcastTransmitSlaveWorldview := make(chan worldview.Slave, 1024)
	broadcastReceiveSlaveWorldview := make(chan worldview.Slave, 1024)

	go broadcast.Transmitter(networkPort, broadcastTransmitMasterWorldview, broadcastTransmitSlaveWorldview)
	go broadcast.Receiver(networkPort, broadcastReceiveMasterWorldview, broadcastReceiveSlaveWorldview)

	checkTimeoutsTicker := time.NewTicker(heartbeatInterval)
	defer checkTimeoutsTicker.Stop()

	lastSeenSlave := make([]time.Time, elevatorCount)
	isSlaveTimedOut := make([]bool, elevatorCount)
	for slaveIndex := range isSlaveTimedOut {
		isSlaveTimedOut[slaveIndex] = true
	}

	lastSeenMaster := time.Time{}
	isDifferentMasterTimedOut := true

	for {
		select {

		case masterWorldview := <-broadcastReceiveMasterWorldview:
			isBroadcastFromSelf := masterWorldview.NetworkID == networkID

			if !isBroadcastFromSelf {
				lastSeenMaster = time.Now()

				if isDifferentMasterTimedOut {
					isDifferentMasterTimedOut = false
				}

				networkToSlave <- masterWorldview
				networkToMaster <- masterWorldview
			}

		case slaveWorldview := <-broadcastReceiveSlaveWorldview:
			isBroadcastFromSelf := slaveWorldview.NetworkID == networkID

			if !isBroadcastFromSelf {
				slaveIndex := slaveWorldview.NetworkID - 1
				lastSeenSlave[slaveIndex] = time.Now()

				if isSlaveTimedOut[slaveIndex] {
					networkToMaster <- connection.NewSlave{NetworkID: slaveWorldview.NetworkID}
					isSlaveTimedOut[slaveIndex] = false
				}

				networkToMaster <- slaveWorldview
			}

		case masterWorldview := <-masterToNetwork:
			broadcastTransmitMasterWorldview <- masterWorldview

			// Make single elevator work on network errors:
			networkToSlave <- masterWorldview

		case slaveWorldview := <-slaveToNetwork:
			broadcastTransmitSlaveWorldview <- slaveWorldview

			slaveIndex := slaveWorldview.NetworkID - 1
			lastSeenSlave[slaveIndex] = time.Now()

			if isSlaveTimedOut[slaveIndex] {
				networkToMaster <- connection.NewSlave{NetworkID: slaveWorldview.NetworkID}
				isSlaveTimedOut[slaveIndex] = false
			}

			// Make single elevator work on network errors:
			networkToMaster <- slaveWorldview

		case <-checkTimeoutsTicker.C:
			if !isDifferentMasterTimedOut && time.Since(lastSeenMaster) > timeoutDuration {
				isDifferentMasterTimedOut = true
				networkToMaster <- connection.MasterTimeout{}
			}

			for slaveIndex := range elevatorCount {
				if !isSlaveTimedOut[slaveIndex] && time.Since(lastSeenSlave[slaveIndex]) > timeoutDuration {
					slaveNetworkID := slaveIndex + 1
					isSlaveTimedOut[slaveIndex] = true
					networkToMaster <- connection.SlaveTimeout{NetworkID: slaveNetworkID}
				}
			}
		}
	}
}
