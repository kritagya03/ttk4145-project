package models

import "time"

// TODO: Change all????
const NetworkBufferSize int = 1024

const HeartbeatInterval time.Duration = 5 * time.Millisecond

// const HeartbeatInterval time.Duration = 3 * time.Second // ! Temp
const HeartbeatTimeout time.Duration = HeartbeatInterval * time.Duration(100)
const BaseElectionTimeout time.Duration = HeartbeatInterval * time.Duration(3)
const MergingMastersTimeoutDuration time.Duration = HeartbeatInterval * time.Duration(100)
const DoorOpenTimeoutDuration time.Duration = time.Duration(time.Second * 3) // TODO: change
// const HardwarePollInterval time.Duration = time.Millisecond * 20
