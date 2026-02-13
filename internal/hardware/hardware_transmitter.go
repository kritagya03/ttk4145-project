package hardware

// A lot of common code with network_transmitter.go
// Make good quality code in network_transmitter.go before using it as a reference.
func Transmitter(hardwareCommands <-chan []byte, address string) {
	// Setup connection

	// Foor loop of reading packet from hardwareCommands channel and writing the packet to the connection
}
