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

var (
	macAddr = [6]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66}

	ipAddr = "192.168.5.100"

	plcaNodeID    uint = 1
	plcaNodeCount uint = 8
)

var inst = lan865x.Inst{
	MAC: &t1s.MACConf{
		Addr: macAddr,
	},
	PLCA: &plca,
}

var plca = t1s.PLCAConf{
	NodeID:     uint8(plcaNodeID),
	NodeCount:  uint8(plcaNodeCount),
	BurstCount: 0,
	BurstTimer: 128,
}

var (
	noRepeat = false
	svcPause = 10 * time.Millisecond
)

func main() {
	lan865x.Ticks = &ticksProvider{t0: time.Now()}

	log, srvLog, hwIntf := initPlatform()
	inst.Dev = hwIntf

	httpsrv.SetLED = setLED
	stack := httpsrv.Setup(srvLog, ipAddr, macAddr)
	inst.UpperProto = &proto{stack: stack}

	if ok := inst.Init(); !ok {
		log.Error("init failed")
		return
	}
	log.Info("init done")

	for {
		for {
			done := inst.Service()
			if done || noRepeat {
				break
			}
		}
		time.Sleep(svcPause)
	}
}
