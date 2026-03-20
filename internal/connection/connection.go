package connection

type MasterTimeout struct{}

type NewSlave struct {
	NetworkID int
}

type SlaveTimeout struct {
	NetworkID int
}

func (MasterTimeout) IsInNetworkToMasterInterface() {}
func (NewSlave) IsInNetworkToMasterInterface()      {}
func (SlaveTimeout) IsInNetworkToMasterInterface()  {}
