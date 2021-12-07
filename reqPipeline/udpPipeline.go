package reqPipeline

import (
    "context"
    "errors"
    "net"
    "strings"

    "github.com/Ishan27g/go-utils/mLogger"
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
    // removeClient from existing list, close connection and stop response listener
    removeClient(address string)
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
    udpConns        map[string]udpConn
    responses       map[string]chan []byte
    requestPipeline chan req
}

func InitSocket() SocketChannel {
    ctx, can := context.WithCancel(context.Background())
    sc := socketChannel{
        ctx:             ctx,
        cancel:          can,
        logger:          mLogger.Get("Client-Socket"),
        udpConns:        make(map[string]udpConn),
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
    for _, c := range sc.udpConns {
        c.c.Close()
        c.cancel()
    }
    sc.udpConns = nil
    sc.responses = nil
    sc.requestPipeline = nil
}


// addClient to communicate with
func (sc *socketChannel) addClient(address string) error {
    c := connect(address, sc)
    if c == nil {
        sc.logger.Error("could not connect to " + address)
        return errors.New("could not connect to " + address)
    }
    ctx, cancel := context.WithCancel(context.Background())
    sc.udpConns[address] = udpConn{
        ctx:    ctx,
        cancel: cancel,
        c:      c,
    }
    sc.startClientListener(address)
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

// connect opens a udp connection with this address
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
func (sc *socketChannel) removeClient(address string) {
    c := make(map[string]udpConn)
    for s, con := range sc.udpConns {
        if strings.Compare(s, address) != 0 {
            c[address] = con
        } else {
            c[address].c.Close() // close upd connection
            c[address].cancel()  // close response listener
        }
    }
    sc.udpConns = c
}

func (sc *socketChannel) SendOneWay(address string, data []byte) chan []byte {
    sc.requestPipeline <- req{
        address: address,
        data:    data,
    }
    if sc.addClient(address) == nil{
        sc.startClientListener(address)
        return sc.responses[address]
    }else {
        return nil
    }
}

func (sc *socketChannel) startRequestListener() {
    // for {
        select {
        case <-sc.ctx.Done():
            return
        case r := <- sc.requestPipeline:
            go func(r req) {
                c := sc.udpConns[r.address].c
                if c != nil {
                    sc.logger.Trace("Sending to UDP client " + c.RemoteAddr().String())
                    _, _ = c.Write(r.data)
                    sc.removeClient(r.address)
                }
            }(r)
        }
    //}
}

