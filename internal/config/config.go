package config

import (
	"flag"
	"fmt"
	"os"
	"time"
)

type Config struct {
	NetworkID               int
	FloorCount              int
	ElevatorCount           int
	DoorOpenDuration        time.Duration
	DurationUntilStuck      time.Duration
	HeartbeatInterval       time.Duration
	TimeoutDuration         time.Duration
	BaseElectionDuration    time.Duration
	MergingMastersDuration  time.Duration
	NetworkPort             int
	HardwarePort            int
	HallRequestAssignerPath string
}

func (c Config) Print() {
	fmt.Printf("The config is:\n")
	fmt.Printf("  Network ID: %d\n", c.NetworkID)
	fmt.Printf("  Floor Count: %d\n", c.FloorCount)
	fmt.Printf("  Elevator Count: %d\n", c.ElevatorCount)
	fmt.Printf("  Network Port: %d\n", c.NetworkPort)
	fmt.Printf("  Hardware Port: %d\n", c.HardwarePort)
	fmt.Printf("  Door Open Duration: %v\n", c.DoorOpenDuration)
	fmt.Printf("  Duration until stuck: %v\n", c.DurationUntilStuck)
	fmt.Printf("  Heartbeat Interval: %v\n", c.HeartbeatInterval)
	fmt.Printf("  Timeout Duration: %v\n", c.TimeoutDuration)
	fmt.Printf("  Base Election Duration: %v\n", c.BaseElectionDuration)
	fmt.Printf("  Merge Masters Duration: %v\n", c.MergingMastersDuration)
	fmt.Printf("  Hall Request Assigner Path: %v\n", c.HallRequestAssignerPath)
}

func Parse() Config {
	var config Config

	var doorOpenDurationSeconds, durationUntilStuckSeconds,
		heartbeatIntervalMilliseconds, heartbeatCountBeforeTimeout,
		heartbeatCountPerBaseElection, heartbeatCountPerMerging int

	flag.IntVar(&config.NetworkID, "network-id", -1,
		"Elevator network ID. Must be >= 1 and <= elevator-count")

	flag.IntVar(&config.FloorCount, "floor-count", 4,
		"Number of floors")

	flag.IntVar(&config.ElevatorCount, "elevator-count", 3,
		"Number of elevators")

	flag.IntVar(&config.NetworkPort, "network-port", 30045,
		"The network port for communication between nodes in the elevator system")

	flag.IntVar(&config.HardwarePort, "hardware-port", 15657,
		"The network port to communicate with the elevator hardware")

	flag.IntVar(&doorOpenDurationSeconds, "door-open-duration-seconds", 3,
		"Number of whole seconds the door should be open")

	flag.IntVar(&durationUntilStuckSeconds, "duration-until-stuck-seconds", 3,
		"Number of whole seconds until the master considers a slave to be stuck")

	flag.IntVar(&heartbeatIntervalMilliseconds, "heartbeat-interval-milliseconds", 20,
		"Number of whole milliseconds between each heartbeat")

	flag.IntVar(&heartbeatCountBeforeTimeout, "heartbeat-count-before-timeout", 100,
		"Number of heartbeat intervals before a node is considered timed out")

	flag.IntVar(&heartbeatCountPerBaseElection, "heartbeat-count-per-base-election", 3,
		"Number of heartbeat intervals in one base election. The base election is multiplied by the NetworkID to get the full election duration of one node")

	flag.IntVar(&heartbeatCountPerMerging, "heartbeat-count-per-merging", 100,
		"Number of heartbeat intervals in one merging of masters")

	flag.StringVar(&config.HallRequestAssignerPath, "hall-request-assigner-path", "hall_request_assigner",
		"The path to the hall request assigner executable")

	flag.Parse()

	config.DoorOpenDuration = time.Second * time.Duration(doorOpenDurationSeconds)
	config.DurationUntilStuck = time.Second * time.Duration(durationUntilStuckSeconds)
	config.HeartbeatInterval = time.Millisecond * time.Duration(heartbeatIntervalMilliseconds)
	config.TimeoutDuration = config.HeartbeatInterval * time.Duration(heartbeatCountBeforeTimeout)
	config.BaseElectionDuration = config.HeartbeatInterval * time.Duration(heartbeatCountPerBaseElection)
	config.MergingMastersDuration = config.HeartbeatInterval * time.Duration(heartbeatCountPerMerging)

	validate(config)

	return config
}

func validate(config Config) {
	var error error

	if config.NetworkID < 1 || config.NetworkID > config.ElevatorCount {
		error = fmt.Errorf("network-id must be >= 1 and <= %d (elevator-count), but is %d", config.ElevatorCount, config.NetworkID)
	} else if config.FloorCount < 1 {
		error = fmt.Errorf("floor-count must be >= 1, but is %d", config.FloorCount)
	} else if config.ElevatorCount < 1 {
		error = fmt.Errorf("elevator-count must be >= 1, but is %d", config.ElevatorCount)
	} else if config.NetworkPort < 1024 || config.NetworkPort > 65535 {
		error = fmt.Errorf("network-port must be in [1024, 65535], but is %d", config.NetworkPort)
	} else if config.HardwarePort < 1024 || config.HardwarePort > 65535 {
		error = fmt.Errorf("hardware-port must be in [1024, 65535], but is %d", config.HardwarePort)
	} else if config.DoorOpenDuration < time.Second {
		error = fmt.Errorf("door-open-duration-seconds must be >= 1, but is %v", config.DoorOpenDuration)
	} else if config.DurationUntilStuck < time.Second {
		error = fmt.Errorf("duration-until-stuck-seconds must be >= 1, but is %v", config.DoorOpenDuration)
	} else if config.HeartbeatInterval < time.Millisecond {
		error = fmt.Errorf("heartbeat-interval-milliseconds must be >= 1, but is %v", config.HeartbeatInterval)
	} else if config.HallRequestAssignerPath == "" {
		error = fmt.Errorf("hall-request-assigner-path must not be empty")
	} else {
		heartbeatInterval := int64(config.HeartbeatInterval)
		heartbeatCountBeforeTimeout := int64(config.TimeoutDuration) / heartbeatInterval
		heartbeatCountPerBaseElection := int64(config.BaseElectionDuration) / heartbeatInterval
		heartbeatCountPerMerging := int64(config.MergingMastersDuration) / heartbeatInterval

		if heartbeatCountBeforeTimeout < 2 {
			error = fmt.Errorf("heartbeat-count-before-timeout must be >= 2, but is %v", heartbeatCountBeforeTimeout)
		} else if heartbeatCountPerBaseElection < 2 {
			error = fmt.Errorf("heartbeat-count-per-base-election must be >= 2, but is %v", heartbeatCountPerBaseElection)
		} else if heartbeatCountPerMerging < 2 {
			error = fmt.Errorf("heartbeat-count-per-merging must be >= 2, but is %v", heartbeatCountPerMerging)
		}
	}

	if error != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", error)
		flag.Usage()
		os.Exit(1)
	}
}
