package led

import (
	"machine"
)

type LED interface {
	Init() error
	Set(bool) error
}

var None = none{}

type none struct{}

func (none) Init() error    { return nil }
func (none) Set(bool) error { return nil }

type PinLED struct {
	Pin machine.Pin
}

func (led *PinLED) Init() error {
	led.Pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	return nil
}

func (led *PinLED) Set(state bool) error {
	led.Pin.Set(state)
	return nil
}
