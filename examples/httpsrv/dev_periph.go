//go:build !tinygo

package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
	"periph.io/x/host/v3/rpi"
)

type hwIntf struct {
	spiConn       spi.Conn
	spiPortCloser spi.PortCloser
	log           *slog.Logger
}

func (d *hwIntf) IntrActive() bool {
	active := intrPin.Read() == false
	if active {
		d.log.Info("intr")
	}
	return active
}

var (
	debugLevel uint

	resetPinName = flag.String("reset-pin", "GPIO13", "name of LAN865x reset pin")
	intrPinName  = flag.String("intr-pin", "GPIO26", "name of LAN865x interrupt pin")
	spidevName   = flag.String("spidev", "/dev/spidev0.1", "name of the SPI device")
)
var resetPin gpio.PinOut
var intrPin gpio.PinIn

func (d *hwIntf) Reset() error {
	if false {
		resetPin.Out(gpio.Low)
		time.Sleep(10 * time.Millisecond)
		resetPin.Out(gpio.High)
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

func (d *hwIntf) SpiTxRx(tx, rx []byte, done func(error)) error {
	// log.Printf("spi <- % x", tx)
	err := d.spiConn.Tx(tx, rx)
	// log.Printf("spi -> % x", rx)
	done(err)
	return err
}

func (d *hwIntf) setupSPI(spidev string) error {
	port, err := spireg.Open(spidev)
	if err != nil {
		return err
	}
	c, err := port.Connect(5*physic.MegaHertz, spi.Mode0, 8)
	if err != nil {
		port.Close()
		return err
	}
	d.spiPortCloser = port
	d.spiConn = c
	return nil
}

func newHardwareIntf(logger *slog.Logger) *hwIntf {
	useCSMACD := false
	flag.BoolVar(&useCSMACD, "csmacd", useCSMACD, "use CSMA/CD, disable PLCA")
	flag.UintVar(&plcaNodeID, "plca-id", plcaNodeID, "PLCA node id")
	flag.UintVar(&plcaNodeCount, "plca-count", plcaNodeCount, "PLCA node count")
	flag.StringVar(&ipAddr, "ip", ipAddr, "IP address")
	flag.UintVar(&debugLevel, "D", 0, "ethernet packet trace level")
	flag.Parse()

	if useCSMACD {
		inst.PLCA = nil
	}

	_, err := host.Init()
	if err != nil {
		log.Fatalf("failed to initialize periph: %v", err)
	}

	if !rpi.Present() {
		log.Fatal("not running on an RPi")
	}
	hw.log = logger

	resetPin = lookupPin(*resetPinName)
	resetPin.Out(gpio.High)

	intrPin = lookupPin(*intrPinName)
	intrPin.In(gpio.PullNoChange, gpio.NoEdge)

	err = hw.setupSPI(*spidevName)
	if err != nil {
		log.Fatalf("failed to initialize spidev: %v", err)
	}

	return &hw
}

func lookupPin(name string) gpio.PinIO {
	pin := gpioreg.ByName(name)
	if pin == nil {
		log.Fatal("pin not found:", name)
	}
	return pin
}

type ticksProvider struct {
	t0 time.Time
}

func (tp *ticksProvider) Milliseconds() uint32 {
	return uint32(time.Since(tp.t0) / 1e6)
}

var hw hwIntf

func newLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func setLED(state bool) {
	fmt.Println("LED state:", state)
}

func traceMsg(dir, proto string, packet []byte, err error) {
	if debugLevel == 0 {
		return
	}
	prefix := dir + " " + proto
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s error: %v\n", prefix, err)
		return
	}
	if len(packet) == 0 {
		return
	}
	if debugLevel == 1 {
		fmt.Fprintf(os.Stderr, "%s [%d]\n", prefix, len(packet))
		return
	}
	fmt.Fprintf(os.Stderr, "%s [%d] % x\n", prefix, len(packet), packet)
}
