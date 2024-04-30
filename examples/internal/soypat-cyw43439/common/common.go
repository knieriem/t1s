package common

import (
	"log/slog"
	"net/netip"

	"github.com/knieriem/t1s/lan865x"
	"github.com/soypat/seqs/stacks"
)

const mtu = lan865x.MTU

type SetupConfig struct {
	// DHCP requested hostname.
	Hostname string
	// DHCP requested IP address. On failing to find DHCP server is used as static IP.
	RequestedIP string
	Logger      *slog.Logger
	// Number of UDP ports to open for the stack. (we'll actually open one more than this for DHCP)
	UDPPorts uint16
	// Number of TCP ports to open for the stack.
	TCPPorts uint16
}

type proto struct {
}

func SetupWithDHCP(cfg SetupConfig) (*stacks.PortStack, error) {
	logger := cfg.Logger
	var err error
	var reqAddr netip.Addr
	if cfg.RequestedIP != "" {
		reqAddr, err = netip.ParseAddr(cfg.RequestedIP)
		if err != nil {
			return nil, err
		}
	}

	// cfg.Logger = logger // Uncomment to see in depth info on wifi device functioning.
	logger.Info("initializing pico W device...")
	mac := [6]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66}

	stack := stacks.NewPortStack(stacks.PortStackConfig{
		MAC:             mac,
		MaxOpenPortsUDP: int(cfg.UDPPorts),
		MaxOpenPortsTCP: int(cfg.TCPPorts),
		MTU:             mtu,
		Logger:          logger,
	})

	//	dev.RecvEthHandle(stack.RecvEth)

	// Begin asynchronous packet handling.
	//	go nicLoop(dev, stack)
	stack.SetAddr(reqAddr) // It's important to set the IP address after DHCP completes.

	return stack, nil
}
