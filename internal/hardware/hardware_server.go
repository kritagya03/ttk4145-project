package hardware

import (
	"fmt"

	. "github.com/kritagya03/ttk4145-project/internal/models"
)

func Server[hardwareEventType HardwareEvent, hardwareCommandType HardwareCommand](
	hardwareEvents <-chan []byte,
	slaveHardwareCommands <-chan hardwareCommandType,
	hardwareCommands chan<- []byte,
	slaveHardwareEvents chan<- hardwareEventType) {

	for {
		select {
		case hardwareEvent := <-hardwareEvents:
			fmt.Println("hardware_server.go case hardwareEvents.")
			_ = hardwareEvent
			// CONVERT FROM HARDWARE EVENT
			// slaveHardwareEvents <- CONVERTED FROM HARDWARE EVENT
		case slaveHardwareCommand := <-slaveHardwareCommands:
			fmt.Println("hardware_server.go case slaveHardwareCommands.")
			_ = slaveHardwareCommand
			// CONVERT FROM SLAVE HARDWARE COMMAND
			// hardwareCommands <- CONVERTED FROM SLAVE HARDWARE COMMAND

		}
	}
}
