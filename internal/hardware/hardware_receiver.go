package hardware

// A lot of common code with network_receiver.go.
// Make good quality code in network_receiver.go before using it as a reference.
func Receiver(hardwareEvents chan<- []byte, address string) {
	// Setup connection

	// Foor loop of reading from connection and sending packet to hardwareEvents channel
}
