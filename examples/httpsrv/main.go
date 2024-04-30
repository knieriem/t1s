package main

import (
	"time"

	"github.com/soypat/seqs/stacks"

	"github.com/knieriem/t1s"
	"github.com/knieriem/t1s/examples/internal/soypat-cyw43439/httpsrv"
	"github.com/knieriem/t1s/lan865x"
)

type proto struct {
	stack *stacks.PortStack
}

func (p *proto) SendEthUp(packet []byte) error {
	traceMsg("->", "eth", packet, nil)
	return p.stack.RecvEth(packet)
}
func (p *proto) PollForEth(buf []byte) (int, error) {
	n, err := p.stack.HandleEth(buf)
	traceMsg("<-", "eth", buf[:n], err)
	return n, err
}

var ipAddr = "192.168.5.100"

var inst = lan865x.Inst{
	MAC: &t1s.MACConf{
		Addr: [6]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66},
	},
	PLCA: &plca,
}

var plca = t1s.PLCAConf{
	NodeID:     1,
	NodeCount:  8,
	BurstCount: 0,
	BurstTimer: 128,
}

func main() {
	hwIntf := newHardwareIntf()
	inst.Dev = hwIntf
	lan865x.Ticks = &ticksProvider{t0: time.Now()}

	logger := newLogger()
	inst.DebugInfo = logger.WithGroup("t1s").Info

	httpsrv.SetLED = setLED
	stack := httpsrv.Setup(logger, ipAddr)
	inst.UpperProto = &proto{stack: stack}

	if ok := inst.Init(); !ok {
		println("init failed")
		return
	}
	println("init done")

	for {
		inst.Service()
		time.Sleep(10 * time.Millisecond)
	}
}