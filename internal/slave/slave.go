package slave

import (
	"fmt"

	. "github.com/kritagya03/ttk4145-project/internal/models"
)

func Server(slaveNetworkEvents <-chan MasterWorldview, slaveNetworkCommands chan<- SlaveWorldview) {
	for {
		masterEvent := <-slaveNetworkEvents
		fmt.Printf("slave.go case slaveNetworkEvents. Received MasterWorldview: %+v\n", masterEvent)
		_ = masterEvent
	}
}
