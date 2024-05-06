//go:build tinygo && pico

package main

import (
	"machine"

	"github.com/knieriem/t1s/examples/internal/tinygo/led"
	"github.com/knieriem/t1s/examples/internal/tinygo/spi"
)

var (
	csLAN865x = machine.GP13
	intrPin   = machine.GP15
)

var spidev = spi.Dev{
	Port:       &spi.Port{Intf: machine.SPI1},
	ChipSelect: csLAN865x,
	Conf: &spi.Conf{
		Freq: 4e6,
		Mode: spi.ClockIdlesLowFirstEdge,
	},
}

var mLED = led.PinLED{Pin: machine.LED}
