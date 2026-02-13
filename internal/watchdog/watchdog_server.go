package watchdog

import (
	"fmt"
	"time"
	"github.com/coreos/go-systemd/daemon"
)

func Server() {
	intervalBeforeKilled, err := daemon.SdWatchdogEnabled(false) // 'false' means we are checking the main process's PID
	if err != nil {
		panic(fmt.Sprintf("Error checking for systemd watchdog: %v\n", err))
	}
	
	if intervalBeforeKilled == 0 {
		panic("Systemd watchdog not enabled for this service.\n")
	}

	// Send keep-alive at roughly half the time before systemd would kill the inactive service
	intervalMessage := time.Duration(intervalBeforeKilled / 2)
	
	fmt.Printf("Systemd watchdog enabled with interval: %v. Starting to send keep-alive messages with interval: %v.\n", intervalBeforeKilled, intervalMessage)
	
	ticker := time.NewTicker(intervalMessage)
	defer ticker.Stop()

	for range ticker.C {
		if isServiceHealthy() {
			// Send the keep-alive notification
			_, err := daemon.SdNotify(false, "WATCHDOG=1")
			if err != nil {
				panic(fmt.Sprintf("Error sending systemd watchdog notification: %v\n", err))
			}
		} else {
			panic("Service is unhealthy, NOT sending keep-alive. Systemd should restart us soon.\n")
		}
	}
}

func isServiceHealthy() bool {
	// Implement health check logic here
	return true
}

