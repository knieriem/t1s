//go:build !tinygo

package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
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
	traceEth bool

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

func initPlatform() (mainLog, srvLog *slog.Logger, hwi *hwIntf) {
	useCSMACD := false
	logLevelSpec := "main=i,srv=e,t1s=e"
	flag.BoolVar(&useCSMACD, "csmacd", useCSMACD, "use CSMA/CD, disable PLCA")
	flag.UintVar(&plcaNodeID, "plca-id", plcaNodeID, "PLCA node id")
	flag.UintVar(&plcaNodeCount, "plca-count", plcaNodeCount, "PLCA node count")
	flag.StringVar(&ipAddr, "ip", ipAddr, "IP address")
	flag.BoolVar(&noRepeat, "svc-no-repeat", noRepeat, "skip service repetition")
	flag.DurationVar(&svcPause, "svc-pause", svcPause, "service pause duration")
	flag.StringVar(&logLevelSpec, "D", logLevelSpec, "log levels specification")
	flag.BoolVar(&traceEth, "E", false, "enable ethernet packet traces")
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

	err = updateLogLevelsFromSpec(logLevelSpec)
	if err != nil {
		log.Fatal(err)
	}
	mainLog = newTextLogger(mainLogLevel).WithGroup("main")
	srvLog = newTextLogger(srvLogLevel)
	t1sLog := newTextLogger(logLevel).WithGroup("t1s")
	inst.DebugInfo = t1sLog.Info
	inst.DebugError = t1sLog.Error
	hw.log = mainLog

	resetPin = lookupPin(*resetPinName)
	resetPin.Out(gpio.High)

	intrPin = lookupPin(*intrPinName)
	intrPin.In(gpio.PullNoChange, gpio.NoEdge)

	err = hw.setupSPI(*spidevName)
	if err != nil {
		log.Fatalf("failed to initialize spidev: %v", err)
	}

	return mainLog, srvLog, &hw
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

var mainLogLevel = slog.LevelInfo
var srvLogLevel = slog.LevelError
var t1sLogLevel = slog.LevelError

func updateLogLevelsFromSpec(logSpec string) error {
	for _, expr := range strings.Split(logSpec, ",") {
		var name string
		iAssign := strings.IndexByte(expr, '=')
		if iAssign != -1 {
			name = expr[:iAssign]
			expr = expr[iAssign+1:]
		}
		level, err := parseLogLevelExpr(expr)
		if err != nil {
			return err
		}
		switch name {
		case "all", "":
			mainLogLevel = level
			t1sLogLevel = level
			srvLogLevel = level
		case "main":
			mainLogLevel = level
		case "t1s":
			t1sLogLevel = level
		case "srv":
			srvLogLevel = level
		default:
			return fmt.Errorf("decoding log spec failed: unknown name: %q", name)
		}
	}
	return nil
}

func parseLogLevelExpr(expr string) (slog.Level, error) {
	var l slog.Level
	if len(expr) == 0 {
		return 0, errors.New("empty log expression")
	}

	s := expr[1:]
	switch expr[0] {
	default:
		s = expr
	case 'd':
		l = slog.LevelDebug
	case 'i':
		l = slog.LevelInfo
	case 'w':
		l = slog.LevelWarn
	case 'e':
		l = slog.LevelError
	}
	if len(s) != 0 {
		i, err := strconv.ParseInt(s, 10, 0)
		if err != nil {
			return 0, err
		}
		l += slog.Level(i)
	}
	return l, nil
}

func setLED(state bool) {
	fmt.Println("LED state:", state)
}

func traceMsg(dir, proto string, packet []byte, err error) {
	if !traceEth {
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
	fmt.Fprintf(os.Stderr, "%s [%d] % x\n", prefix, len(packet), packet)
}
