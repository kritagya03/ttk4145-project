package master

import (
	. "github.com/kritagya03/ttk4145-project/internal/models"
)

func Server(masterNetworkEvents <-chan SlaveWorldview, masterNetworkCommands chan<- MasterWorldview) {
	for {
		slaveEvent := <-masterNetworkEvents
		_ = slaveEvent
	}
}
