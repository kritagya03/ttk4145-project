package models

import "time"

// TODO: Change all????
const NetworkBufferSize int = 1024
const HeartbeatInterval time.Duration = 5 * time.Millisecond
const HeartbeatTimeout time.Duration = HeartbeatInterval * time.Duration(100)
const BaseElectionTimeout time.Duration = 15 * time.Millisecond
const MergingMastersTimeoutDuration time.Duration = HeartbeatInterval * time.Duration(100)
