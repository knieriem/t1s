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

var macAddr = [6]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66}

var ipAddr = "192.168.5.100"

var inst = lan865x.Inst{
	MAC: &t1s.MACConf{
		Addr: macAddr,
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
	lan865x.Ticks = &ticksProvider{t0: time.Now()}

	log := newLogger()
	t1sLog := log.WithGroup("t1s")

	hwIntf := newHardwareIntf(t1sLog)
	inst.Dev = hwIntf
	inst.DebugInfo = t1sLog.Info

	httpsrv.SetLED = setLED
	stack := httpsrv.Setup(log, ipAddr, macAddr)
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
