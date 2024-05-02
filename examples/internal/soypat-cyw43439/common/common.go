package common

import (
	"log/slog"
	"net/netip"

	"github.com/knieriem/t1s/lan865x"
	"github.com/soypat/seqs/stacks"
)

const mtu = lan865x.MTU

type SetupConfig struct {
	// MAC address
	MAC [6]byte

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

	logger.Info("initializing stack...")

	stack := stacks.NewPortStack(stacks.PortStackConfig{
		MAC:             cfg.MAC,
		MaxOpenPortsUDP: int(cfg.UDPPorts),
		MaxOpenPortsTCP: int(cfg.TCPPorts),
		MTU:             mtu,
		Logger:          logger,
	})
	stack.SetAddr(reqAddr) // It's important to set the IP address after DHCP completes.

	return stack, nil
}
