package t1s

// UpperProto defines the interface to the layer above
// the ethernet layer.
type UpperProto interface {
	// SendEthUp propagates a packet to the upper layer.
	SendEthUp(pkt []byte) error

	// PollForEth calls the upper protocol layers for ethernet packets
	// to be transmitted. If a packet is available, it is copied
	// to the buffer, and the packet's number of bytes is returned.
	// The capacity of buf must match the MTU of the upper layer.
	PollForEth(buf []byte) (n int, err error)
}

type MACConf struct {
	Addr [6]byte

	CopyAllFrames bool

	TxCutThrough bool
	RxCutThrough bool
}

// PLCAConf defines the Physical Layer Collision Avoidance
// mode settings.
type PLCAConf struct {
	NodeID     uint8
	NodeCount  uint8
	BurstCount uint8
	BurstTimer uint8
}
