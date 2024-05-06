P=httpsrv
TARG=nucleo-g071rb
TARG=pico
LOGLEVEL=info

PROG=$P_$(TARG)_log$(LOGLEVEL)

OPT=z

TGOFLAGS=\
	-target $(TARG)\
	-size short\
	-tags log$(LOGLEVEL)\
	-stack-size=8kb\
	-opt $(OPT)\

all: rpi

tinygo: tinygo-hex
tinygo-hex: $(PROG).hex
tinygo-elf: $(PROG).elf

%.hex: *.go
	tinygo build -o $@ $(TGOFLAGS) .

%.elf: *.go
	tinygo build -o $@ $(TGOFLAGS) .

tinygo-flash:
	tinygo flash $(TGOFLAGS) .

rpi:
	GOARCH=arm64 \
	CGO_ENABLED=1 \
	CC=/usr/bin/aarch64-linux-gnu-gcc-12 \
	go build -o $P.rpi \
		-tags netgo,log$(LOGLEVEL) \
		-ldflags '-linkmode external -extldflags=-static -s -w'

clean:
	rm -f $P.rpi
	rm -f $P_*.hex
	rm -f $P_*.elf
