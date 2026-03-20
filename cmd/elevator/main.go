package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/kritagya03/ttk4145-project/internal/call"
	"github.com/kritagya03/ttk4145-project/internal/config"
	"github.com/kritagya03/ttk4145-project/internal/master"
	"github.com/kritagya03/ttk4145-project/internal/messages"
	"github.com/kritagya03/ttk4145-project/internal/network"
	"github.com/kritagya03/ttk4145-project/internal/slave"
	"github.com/kritagya03/ttk4145-project/internal/worldview"
)

func main() {
	var isWorker bool
	flag.BoolVar(&isWorker, "start-worker", false, "internal use only: starts the elevator logic")

	config := config.Parse()

	if !isWorker {
		workerRestartDelay := config.TimeoutDuration + config.BaseElectionDuration + config.MergingMastersDuration

		runSupervisor(workerRestartDelay)
	} else {
		runWorker(config)
	}
}

func runSupervisor(workerRestartDelay time.Duration) {
	fmt.Println("Supervisor started.")

	for {
		fmt.Printf("A worker will start after a delay of %v to allow a different node to become master.\n", workerRestartDelay)

		time.Sleep(workerRestartDelay)

		commandLineArguments := append(os.Args[1:], "-start-worker")
		workerCommand := exec.Command(os.Args[0], commandLineArguments...)

		workerCommand.Stdout = os.Stdout
		workerCommand.Stderr = os.Stderr

		startTime := time.Now()

		workerError := workerCommand.Run()

		workerLifeDuration := time.Since(startTime)

		if workerError != nil {
			fmt.Printf("Worker crashed after %v. Error: %v.\n", workerLifeDuration, workerError)
		} else {
			fmt.Printf("Worker exited cleanly after %v.\n", workerLifeDuration)
		}
	}
}

func runWorker(config config.Config) {
	config.Print()

	callTypeCount := call.HallCallTypeCount + config.ElevatorCount

	masterToNetwork := make(chan worldview.Master, 1024)
	slaveToNetwork := make(chan worldview.Slave, 1024)
	networkToMaster := make(chan messages.NetworkToMaster, 1024)
	networkToSlave := make(chan worldview.Master, 1024)

	go network.Server(masterToNetwork, slaveToNetwork,
		networkToMaster, networkToSlave,
		config.NetworkPort, config.NetworkID,
		config.ElevatorCount, config.HeartbeatInterval,
		config.TimeoutDuration)

	go master.Server(networkToMaster, masterToNetwork,
		config.NetworkID, config.FloorCount,
		callTypeCount, config.ElevatorCount,
		config.HeartbeatInterval, config.BaseElectionDuration,
		config.MergingMastersDuration, config.DurationUntilStuck,
		config.HallRequestAssignerPath)

	go slave.Server(networkToSlave, slaveToNetwork,
		config.NetworkID, config.FloorCount,
		callTypeCount, config.HardwarePort,
		config.DoorOpenDuration, config.HeartbeatInterval)

	select {}
}
