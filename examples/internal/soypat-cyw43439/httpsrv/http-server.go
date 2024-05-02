package httpsrv

import (
	"bufio"
	"io"
	"log/slog"
	"net/netip"
	"time"

	_ "embed"

	"github.com/soypat/seqs/httpx"
	"github.com/soypat/seqs/stacks"

	"github.com/knieriem/t1s/examples/internal/soypat-cyw43439/common"
	"github.com/knieriem/t1s/lan865x"
)

const connTimeout = 3 * time.Second
const maxconns = 3

// TODO: Not sure if the tcpbufsize is computed right here.
// The original source says: "tcpbufsize = 2030 // MTU - ethhdr - iphdr - tcphdr"
// With the MTU being 2048 - mtuPrefix = 2030, with mtuPrefix =  (2 + 12 + 4),
// it appears that ethhdr, iphdr and tcphdr have not been subtracted from the MTU
// value. So perhaps, in the original code, tcpbufsize could have been
// 2030 - 14 - 20 - 20 = 1986.
const tcpbufsize = lan865x.MTU - 14 - 20 - 20 // MTU - ethhdr - iphdr - tcphdr
const hostname = "http-pico"

var (
	// We embed the html file in the binary so that we can edit
	// index.html with pretty syntax highlighting.
	//
	//go:embed index.html
	webPage      []byte
	lastLedState bool
)

// This is our HTTP hander. It handles ALL incoming requests. Path routing is left
// as an excercise to the reader.
func httpHandler(respWriter io.Writer, resp *httpx.ResponseHeader, req *httpx.RequestHeader) {
	uri := string(req.RequestURI())
	resp.SetConnectionClose()
	switch uri {
	case "/":
		println("Got webpage request!")
		resp.SetContentType("text/html")
		resp.SetContentLength(len(webPage))
		respWriter.Write(resp.Header())
		respWriter.Write(webPage)

	case "/toggle-led":
		println("Got toggle led request!")
		respWriter.Write(resp.Header())
		lastLedState = !lastLedState
		SetLED(lastLedState)

	default:
		println("Path not found:", uri)
		resp.SetStatusCode(404)
		respWriter.Write(resp.Header())
	}
}

var SetLED = func(state bool) {}

func Setup(logger *slog.Logger, ipAddr string, mac [6]byte) *stacks.PortStack {
	stack, err := common.SetupWithDHCP(common.SetupConfig{
		MAC:         mac,
		RequestedIP: ipAddr,
		Hostname:    "TCP-pico",
		Logger:      logger,
		TCPPorts:    1,
	})
	if err != nil {
		panic("setup DHCP:" + err.Error())
	}
	go func() {
		// Start TCP server.
		const listenPort = 80
		listenAddr := netip.AddrPortFrom(stack.Addr(), listenPort)
		listener, err := stacks.NewTCPListener(stack, stacks.TCPListenerConfig{
			MaxConnections: maxconns,
			ConnTxBufSize:  tcpbufsize,
			ConnRxBufSize:  tcpbufsize,
		})
		if err != nil {
			panic("listener create:" + err.Error())
		}
		err = listener.StartListening(listenPort)
		if err != nil {
			panic("listener start:" + err.Error())
		}
		// Reuse the same buffers for each connection to avoid heap allocations.
		var req httpx.RequestHeader
		var resp httpx.ResponseHeader
		buf := bufio.NewReaderSize(nil, 1024)
		logger.Info("listening",
			slog.String("addr", "http://"+listenAddr.String()),
		)

		for {
			conn, err := listener.Accept()
			if err != nil {
				logger.Error("listener accept:", slog.String("err", err.Error()))
				time.Sleep(time.Second)
				continue
			}
			logger.Info("new connection", slog.String("remote", conn.RemoteAddr().String()))
			err = conn.SetDeadline(time.Now().Add(connTimeout))
			if err != nil {
				conn.Close()
				logger.Error("conn set deadline:", slog.String("err", err.Error()))
				continue
			}
			buf.Reset(conn)
			err = req.Read(buf)
			if err != nil {
				logger.Error("hdr read:", slog.String("err", err.Error()))
				conn.Close()
				continue
			}
			resp.Reset()
			httpHandler(conn, &resp, &req)
			// time.Sleep(100 * time.Millisecond)
			conn.Close()
		}
	}()
	return stack
}
