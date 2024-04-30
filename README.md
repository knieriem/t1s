# LAN865x 10BASE-T1S user space driver for Go

This repository contains a Go wrapper around Microchip's
OPEN Alliance TC6 Protocol Driver for LAN8650/1, [oa-tc6-lib].

The experimental driver has been created as a development tool to help getting started using the LAN8651 on a generic microcontroller.

The [HTTP server example] from soypat's [cyw43439 driver package] has been adapted to provide a simple http server over T1S (see [examples/internal/soypat-cyw43439] for license and imported files).
 The HTTP server example can be run on Raspberry Pi 4B making use of the [periph] library,
or, compiled with TinyGo, on the Raspberry Pi Pico.

The HTTP server can be accessed from a client at another T1S node,
which can, for instance, be a RPi 4 using Microchip's LAN865x linux driver.
As in the original example for the Pico W, an LED can be toggled
from within a web browser running on the client node.

To access a T1S network a [Two-Wire ETH Click] board or similar boards can be used.


[oa-tc6-lib]: https://github.com/MicrochipTech/oa-tc6-lib

[HTTP server example]: https://github.com/soypat/cyw43439/tree/main/examples/http-server

[examples/internal/soypat-cyw43439]: ./examples/internal/soypat-cyw43439

[cyw43439 driver package]: https://github.com/soypat/cyw43439
[periph]: https://periph.io

[soypat/seqs]: https://github.com/soypat/seqs

[Two-Wire Eth Click]: https://www.mikroe.com/two-wire-eth-click
