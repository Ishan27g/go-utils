package udpChannel

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/hashicorp/go-hclog"
)

const MaxRequestsPipelined = 10

type SocketChannel interface {
	// Stop all communication
	Stop()
	// SendOneWay sends the data to this connected client. The request is pipelined
	// and the response is returned as channel
	SendOneWay(address string, data []byte) chan []byte
	// addClient to communicate with, open connection and start response listener
	addClient(address string) error
	// removeClients from existing list, close connections and stop response listener
	removeClients()
}
type req struct {
	address string
	data    []byte
}
type udpConn struct {
	ctx    context.Context
	cancel context.CancelFunc
	c      *net.UDPConn
}
type socketChannel struct {
	ctx             context.Context
	cancel          context.CancelFunc
	logger          hclog.Logger
	udpConns        map[string]*udpConn
	responses       map[string]chan []byte
	requestPipeline chan req
}

func InitSocket() SocketChannel {
	ctx, can := context.WithCancel(context.Background())
	sc := socketChannel{
		ctx:             ctx,
		cancel:          can,
		logger:          hclog.New(hclog.DefaultOptions),
		udpConns:        make(map[string]*udpConn),
		responses:       make(map[string]chan []byte),
		requestPipeline: make(chan req, MaxRequestsPipelined),
	}
	go sc.startRequestListener()
	return &sc
}

// Stop all communication with all clients.
// close send udp connections, quit response listeners, close response channels,
func (sc *socketChannel) Stop() {
	sc.cancel()
	close(sc.requestPipeline)
	for _, c := range sc.responses {
		close(c)
	}
	sc.removeClients()
	sc.udpConns = nil
	sc.responses = nil
	sc.requestPipeline = nil
}

// addClient to communicate with
func (sc *socketChannel) addClient(address string) error {
	if sc.udpConns[address] != nil {
		fmt.Println("already connected to " + address)
		return nil
	}
	c := connect(address, sc)
	if c == nil {
		sc.logger.Error("could not connect to " + address)
		return errors.New("could not connect to " + address)
	}
	ctx, cancel := context.WithCancel(context.Background())
	sc.udpConns[address] = &udpConn{
		ctx:    ctx,
		cancel: cancel,
		c:      c,
	}
	go sc.startClientListener(address)
	return nil
}

// startClientListener reads response messages from this client until context is done.
// response is sent to the response Channel for this client
func (sc *socketChannel) startClientListener(address string) {
	c := sc.udpConns[address].c
	defer sc.udpConns[address].cancel()
	for {
		select {
		case <-sc.udpConns[address].ctx.Done():
			return
		default:
			buffer := make([]byte, 1024)
			readLen, _, err := c.ReadFromUDP(buffer)
			if err != nil {
				sc.logger.Error(err.Error())
				return
			}
			buffer = buffer[:readLen]
			if buffer != nil {
				sc.responses[address] <- buffer
			}
		}
	}
}

// connect opens the udp connection to this address
func connect(address string, sc *socketChannel) *net.UDPConn {
	s, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		sc.logger.Error(err.Error())
		return nil
	}
	c, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		sc.logger.Error(err.Error())
		return nil
	}
	return c
}

// removeClient stops communication with this client
func (sc *socketChannel) removeClients() {
	for address := range sc.udpConns {
		sc.udpConns[address].cancel()      // close response listener
		_ = sc.udpConns[address].c.Close() // close upd connection
		delete(sc.udpConns, address)
	}
}

func (sc *socketChannel) SendOneWay(address string, data []byte) chan []byte {
	if sc.addClient(address) == nil {
		sc.requestPipeline <- req{
			address: address,
			data:    data,
		}
		return sc.responses[address]
	} else {
		return nil
	}
}

func (sc *socketChannel) startRequestListener() {
	for {
		select {
		case <-sc.ctx.Done():
			sc.removeClients()
			return
		case r := <-sc.requestPipeline:
			go func(r req) {
				c := sc.udpConns[r.address].c
				if c != nil {
					sc.logger.Trace("Sending to UDP client " + c.RemoteAddr().String())
					_, _ = c.Write(r.data)
				}
			}(r)
		}
	}
}
