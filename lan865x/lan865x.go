// Package lan865x implements an experimental driver wrapper around
// Microchip's OA TC6 library for LAN8650/1.

package lan865x

import (
	"errors"
	"unsafe"

	"github.com/knieriem/t1s"
	"github.com/knieriem/t1s/lan865x/internal/cgo"
)

// #cgo CFLAGS: -Iinclude
// #cgo CFLAGS: -Ilib/oa-tc6/inc
// #include <stdint.h>
// #include <tc6.h>
// #include <tc6-regs.h>
//
// extern	int	t1s_sendRawEthPacket(TC6_t *pInst, uint8_t *pTx, uint16_t len, uint8_t tsc);
//
import "C"

const MTU = 1536

// Inst contains the state of one LAN865x driver instance.
// To create an instance, public fields MAC, PLCA,
// UpperProto, Dev should be populated first.
// Then .Init may be called to initialize the driver.
//
// If PLCA is set to nil, then CSMA/CD is used.
type Inst struct {
	handle unsafe.Pointer

	MAC  *t1s.MACConf
	PLCA *t1s.PLCAConf

	UpperProto t1s.UpperProto
	Dev        HwIntf

	tc6         *C.TC6_t
	needService bool

	pbuf      []byte
	rxInvalid bool

	txBuf     []byte
	txBufBusy bool

	spiTag uint8

	DebugInfo func(msg string, a ...any)
}

func (inst *Inst) info(msg string, a ...any) {
	if inst.DebugInfo == nil {
		return
	}
	inst.DebugInfo(msg, a...)
}

var pbuf [MTU]byte
var txbuf [MTU]byte

type HwIntf interface {
	Reset() error
	IntrActive() bool
	SpiTxRx(tx, rx []byte, done func(err error)) error
}

func cBool(v bool) C.int {
	if v {
		return 1
	}
	return 0
}

var nullPLCAConf t1s.PLCAConf

func (inst *Inst) Init() bool {
	inst.txBuf = txbuf[:]
	inst.pbuf = pbuf[:0]
	h := cgo.NewHandle(inst)
	inst.handle = unsafe.Pointer(&h)
	p := C.TC6_Init(inst.handle)
	if p == nil {
		return false
	}
	mac := inst.MAC
	enablePLCA := true
	plca := inst.PLCA
	if plca == nil {
		enablePLCA = false
		plca = &nullPLCAConf
	}
	ret := C.TC6Regs_Init(p, inst.handle, (*C.uint8_t)(&inst.MAC.Addr[0]),
		cBool(enablePLCA),
		C.uint8_t(plca.NodeID), C.uint8_t(plca.NodeCount),
		C.uint8_t(plca.BurstCount), C.uint8_t(plca.BurstTimer),
		cBool(mac.CopyAllFrames), cBool(mac.TxCutThrough), cBool(mac.RxCutThrough))
	if ret == 0 {
		return false
	}
	for C.TC6Regs_GetInitDone(p) == 0 {
		C.TC6_Service(p, cBool(true))
	}
	inst.tc6 = p
	return true
}

func instFromHandle(context unsafe.Pointer) *Inst {
	h := *(*cgo.Handle)(context)
	return h.Value().(*Inst)
}

func (inst *Inst) Service() (allDone bool) {
	allDone = true

	intrTriggered := inst.Dev.IntrActive()
	if intrTriggered || inst.needService {
		allDone = C.TC6_Service(inst.tc6, cBool(!intrTriggered)) != 0
		if allDone {
			intrTriggered = false
		}
	}

	if !inst.txBufBusy {
		// Check for ethernet frames to be sent down.
		nTx, err := inst.UpperProto.PollForEth(inst.txBuf)
		if err != nil {
		} else if nTx != 0 {
			// After SendEthDown txBufBusy will be set.
			// It is reset later by a callback when
			// transmission finished.
			inst.SendEthDown(inst.txBuf[:nTx])
		}
	}
	C.TC6Regs_CheckTimers()
	return allDone
}

// Ticks must be set to an actual implementation of [TicksProvider]
// before a driver can be initialized.
var Ticks TicksProvider

type TicksProvider interface {
	Milliseconds() uint32
}

func (inst *Inst) SendEthDown(packet []byte) error {
	ret := C.t1s_sendRawEthPacket(inst.tc6, (*C.uint8_t)(&packet[0]), C.uint16_t(len(packet)), 0)
	if ret != 0 {
		return ErrSendFailure
	}
	return nil
}

//export t1s_onRawTxPacket
func t1s_onRawTxPacket(gTag, pTx unsafe.Pointer, nTx uint16) {
	inst := instFromHandle(gTag)
	inst.txBufBusy = false
}

func (inst *Inst) SetPLCA(enable bool, nodeId uint8, nodeCount uint8) error {
	ret := C.TC6Regs_SetPlca(inst.tc6, cBool(enable), C.uint8_t(nodeId), C.uint8_t(nodeCount))
	if ret == 0 {
		return ErrRegsFailure
	}
	return nil
}

//export tc6_onRxEthernetSlice
func tc6_onRxEthernetSlice(_ *C.TC6_t, pRx unsafe.Pointer, offset uint16, nRx uint16, gTag unsafe.Pointer) {
	inst := instFromHandle(gTag)
	//	log.Println("onRxSlice", offset, nRx, inst.rxInvalid)
	if inst.rxInvalid {
		return
	}
	newLen := int(offset + nRx)
	if newLen > cap(inst.pbuf) {
		inst.rxInvalid = true
		return
	}
	if offset != 0 {
		if len(inst.pbuf) == 0 {
			inst.rxInvalid = true
			return
		}
	} else if len(inst.pbuf) != 0 {
		inst.rxInvalid = true
		return
	}
	rx := unsafe.Slice((*byte)(pRx), nRx)
	//	log.Println("onRxSlice added", nRx, newLen)

	inst.pbuf = inst.pbuf[:newLen]
	copy(inst.pbuf[offset:], rx)
}

const (
	// IEEE 802.3 Ethernet Frame Format
	ethPreambleSize       = 7
	ethStartFrameDelSize  = 1
	ethPreambleAndSFDSize = ethPreambleSize + ethStartFrameDelSize

	ethMACDestSize  = 6
	ethMACSrcSize   = 6
	eth8021QtagSize = 4
	ethLengthSize   = 2
	ipHeaderMinSize = 20
	ethFCSsize      = 4

	minPacketSize = ethMACDestSize + ethMACSrcSize +
		ethLengthSize +
		ipHeaderMinSize +
		ethFCSsize
)

//export tc6_onRxEthernetPacket
func tc6_onRxEthernetPacket(_ *C.TC6_t, success int, packetLen uint16, rxTimestamp *uint64, gTag unsafe.Pointer) {
	inst := instFromHandle(gTag)
	status := "ok"
	defer func() {
		inst.info("onRxPacket", "len", packetLen, "status", status)
	}()
	pbuf := inst.pbuf
	inst.pbuf = pbuf[:0]
	rxInvalid := inst.rxInvalid
	inst.rxInvalid = false
	if success == 0 {
		status = "bad"
		return
	}
	if rxInvalid || len(pbuf) == 0 {
		status = "inval"
		return
	}
	if len(pbuf) != int(packetLen) {
		status = "wrong length"
		return
	}
	if packetLen < minPacketSize {
		status = "too short"
		return
	}
	err := inst.UpperProto.SendEthUp(pbuf)
	if err != nil {
		status = "dropped"
	}
}

//export tc6_onError
func tc6_onError(_ *C.TC6_t, e C.TC6_Error_t, gTag unsafe.Pointer) {
	inst := instFromHandle(gTag)
	inst.info("onError", "err", e)
}

//export tc6_onNeedService
func tc6_onNeedService(_ *C.TC6_t, gTag unsafe.Pointer) {
	inst := instFromHandle(gTag)
	inst.needService = true
}

var ErrSendFailure = errors.New("tc6send failed")
var ErrRegsFailure = errors.New("tc6regs call failed")

//export tc6regs_onEvent
func tc6regs_onEvent(_ *C.TC6_t, event C.TC6Regs_Event_t, pTag unsafe.Pointer) {
	inst := instFromHandle(pTag)
	ev := regsEvent(event)
	reinit := ev.needsReinit()
	if reinit {
		C.TC6Regs_Reinit(inst.tc6)
	}
	inst.info("onEvent", "code", ev, "reinit", reinit)
}

type regsEvent uint8

func (ev regsEvent) needsReinit() bool {
	switch ev {
	case C.TC6Regs_Event_Loss_of_Framing_Error,
		C.TC6Regs_Event_RX_Non_Recoverable_Error,
		C.TC6Regs_Event_TX_Non_Recoverable_Error:
		return true
	}
	return false
}

//export tc6_onSpiTransaction
func tc6_onSpiTransaction(instIndex uint8, pTx, pRx unsafe.Pointer, size uint16, gTag unsafe.Pointer) C.int {
	inst := instFromHandle(gTag)
	tx := unsafe.Slice((*byte)(pTx), size)
	rx := unsafe.Slice((*byte)(pRx), size)
	inst.spiTag = instIndex
	err := inst.Dev.SpiTxRx(tx, rx, inst.spiDone)
	return cBool(err == nil)
}

func (inst *Inst) spiDone(err error) {
	C.TC6_SpiBufferDone(C.uint8_t(inst.spiTag), cBool(err == nil))
}

//export tc6regs_getTicksMs
func tc6regs_getTicksMs() uint32 {
	return Ticks.Milliseconds()
}
