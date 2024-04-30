package spi

import (
	"machine"
)

const (
	CPOL0CPHA0 Mode = machine.Mode0
	CPOL0CPHA1 Mode = machine.Mode1
	CPOL1CPHA0 Mode = machine.Mode2
	CPOL1CPHA1 Mode = machine.Mode3

	ClockIdlesLowFirstEdge   = CPOL0CPHA0
	ClockIdlesLowSecondEdge  = CPOL0CPHA1
	ClockIdlesHighFirstEdge  = CPOL1CPHA0
	ClockIdlesHighSecondEdge = CPOL1CPHA1
)

type Mode int

type Port struct {
	Intf     *machine.SPI
	curConf  *Conf
	csActive bool
}

type Dev struct {
	Port       *Port
	ChipSelect machine.Pin
	Conf       *Conf
}

type Conf struct {
	Freq     uint32
	Mode     Mode
	LSBFirst bool
}

func (p *Port) NewDev(cs machine.Pin, conf *Conf) *Dev {
	return &Dev{
		Port:       p,
		ChipSelect: cs,
		Conf:       conf,
	}
}

func (d *Dev) txRxKeepCS(tx, rx []byte, keepCS bool) error {
	n := len(tx)
	if n == 0 {
		n = len(rx)
	}
	p := d.Port
	if n != 0 && !p.csActive {
		if p.curConf != d.Conf {
			p.Intf.Configure(machine.SPIConfig{
				Frequency: d.Conf.Freq,
				Mode:      uint8(d.Conf.Mode),
				LSBFirst:  d.Conf.LSBFirst})
			p.curConf = d.Conf
		}
	}
	if !p.csActive {
		d.ChipSelect.Low()
	}
	var err error
	if n != 0 {
		err = p.Intf.Tx(tx, rx)
	}
	if !keepCS {
		d.ChipSelect.High()
	}
	p.csActive = keepCS
	return err
}

func (d *Dev) TxRxKeepCS(tx, rx []byte) error {
	return d.txRxKeepCS(tx, rx, true)
}

func (d *Dev) TxRx(tx, rx []byte) error {
	return d.txRxKeepCS(tx, rx, false)
}
