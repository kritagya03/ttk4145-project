package models

import "time"

const NetworkBufferSize int = 1024
const HeartbeatTimeout time.Duration = 10 * time.Millisecond //Maybe change????
const HeartbeatInterval time.Duration = 5 * time.Millisecond
const BaseElectionTimeout time.Duration = 15 * time.Millisecond
const CombineMastersTimeoutDuration time.Duration = HeartbeatInterval * time.Duration(100)