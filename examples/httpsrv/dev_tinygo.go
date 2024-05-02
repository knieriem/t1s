//go:build tinygo

package main

import (
	"log/slog"
	"machine"
	"time"

	"github.com/knieriem/t1s/examples/internal/tinygo/spi"
)

type hwIntf struct {
	spidev  *spi.Dev
	intrPin machine.Pin
	log     *slog.Logger
}

func (d *hwIntf) IntrActive() bool {
	active := d.intrPin.Get() == false
	if active {
		d.log.Info("intr")
	}
	return active
}

func (d *hwIntf) Reset() error {
	return nil
}

func (d *hwIntf) SpiTxRx(tx, rx []byte, done func(error)) error {
	//	log.Printf("spi <- % x", tx)
	err := d.spidev.TxRx(tx, rx)
	//	log.Printf("spi -> % x", rx)
	done(err)
	return err
}

var hw = hwIntf{
	spidev:  &spidev,
	intrPin: intrPin,
}

func newHardwareIntf(log *slog.Logger) *hwIntf {
	hw.log = log
	pinout(csLAN865x)
	mLED.Init()
	return &hw
}

type ticksProvider struct {
	t0 time.Time
}

func (tp *ticksProvider) Milliseconds() uint32 {
	return uint32(time.Since(tp.t0) / 1e6)
}

func pinout(p machine.Pin) machine.Pin {
	p.Configure(machine.PinConfig{Mode: machine.PinOutput})
	return p
}

func setLED(state bool) {
	mLED.Set(state)
}

func traceMsg(dir, proto string, packet []byte, err error) {

}
