package slave

import (
	. "github.com/kritagya03/ttk4145-project/internal/models"
)

func Server(slaveNetworkEvents <-chan MasterWorldview, slaveNetworkCommands chan<- SlaveWorldview) {
	for {
		masterEvent := <-slaveNetworkEvents
		_ = masterEvent
	}
}
