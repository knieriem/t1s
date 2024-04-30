# httpsrv

Httpsrv is an example HTTP server, based on
[soypat/cyw43439/examples/http-server], adapted
for being used with LAN865x.

[soypat/cyw43439/examples/http-server]: https://github.com/soypat/cyw43439/tree/main/examples/http-server

The server makes use of [soypat/seqs] as TCP/IP stack.


## httpsrv on Raspberry Pi Pico

The example may be run on boards supported by TinyGo
that provide at least 256 KiB ROM. Currently it has been
tested to run on the _Raspberry PI Pico_.

To compile `httpsrv` and flash the board, run

	make tinygo-flash

The node uses the IP address 192.168.5.100 on default.
Entering this address into a web browser will load a
simple web page from `httpsrv`, that contains a 
toggle button that will, for each click,
send a HTTP request to the `/toggle-led` URL implemented
by `httpsrv`. On the Pico, the green LED will toggle
accordingly.


## httpsrv on Raspberry Pi 4B

A binary for the Raspberry Pi 4B may be built
using `make rpi` or just `make`.

The executable supports some flags:

```
  -D uint
        ethernet packet trace level
  -intr-pin string
        name of LAN865x interrupt pin (default "GPIO26")
  -ip string
        IP address (default "192.168.5.100")
  -reset-pin string
        name of LAN865x reset pin (default "GPIO13")
  -spidev string
        name of the SPI device (default "/dev/spidev6.0")
```

To be able to access an SPI device like `/dev/spidev6.0`,
a device tree overlay can be applied, like

	spi6-1cs,cs0_pin=16
