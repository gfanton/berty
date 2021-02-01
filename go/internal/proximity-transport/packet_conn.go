package proximitytransport

import (
	"context"
	"fmt"
	"net"
	"time"

	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"go.uber.org/zap"
)

// func newPacketConn(ctx context.Context, t *proximityTransport, remoteMa ma.Multiaddr,
// 	remotePID peer.ID, inbound bool) (tpt.CapableConn, error) {
// 	t.logger.Debug("newConn()", zap.String("remoteMa", remoteMa.String()), zap.Bool("inbound", inbound))

// 	// Creates a manet.Conn
// 	pr, pw := io.Pipe()
// 	connCtx, cancel := context.WithCancel(t.listener.ctx)

// 	maconn := &Conn{
// 		readIn:    pw,
// 		readOut:   pr,
// 		localMa:   t.listener.localMa,
// 		remoteMa:  remoteMa,
// 		ready:     false,
// 		cache:     NewRingBufferMap(t.logger, 128),
// 		mp:        newMplex(connCtx, t.logger),
// 		order:     0,
// 		ctx:       connCtx,
// 		cancel:    cancel,
// 		transport: t,
// 	}

// 	// Stores the conn in connMap, will be deleted during conn.Close()
// 	t.connMap.Store(maconn.RemoteAddr().String(), maconn)

// 	// Configure mplex and run it
// 	maconn.mp.addInputCache(t.cache)
// 	maconn.mp.addInputCache(maconn.cache)
// 	maconn.mp.setOutput(pw)
// 	go maconn.mp.run(maconn.RemoteAddr().String())

// 	// Returns an upgraded CapableConn (muxed, addr filtered, secured, etc...)
// 	if inbound {
// 		return t.upgrader.UpgradeInbound(ctx, t, maconn)
// 	}
// 	return t.upgrader.UpgradeOutbound(ctx, t, maconn, remotePID)
// }

type CustomAddr struct {
	network string
	value   string
}

type RcvMsg struct {
	Addr    net.Addr
	Payload []byte
}

func NewCustomAddr(network string, value string) *CustomAddr {
	return &CustomAddr{network, value}
}

// Network name (for example, "tcp", "udp")
func (a *CustomAddr) Network() string {
	return a.network
}

// String form of address (for example, "192.0.2.1:25", "[2001:db8::1]:80")
func (a *CustomAddr) String() string {
	return a.value
}

type PacketWriter interface {
	WriteTo(p []byte, addr Addr) (n int, err error)
}

var _ manet.PacketConn = (*PacketConn)(nil)

// PacketConn is a generic packet-oriented network connection.
// Multiple goroutines may invoke methods on a PacketConn simultaneously.
type PacketConn struct {
	cmsg chan *RcvMsg

	lmaddr ma.Multiaddr
	laddr  net.Addr

	ctx    context.Context
	cancel context.CancelFunc

	logger *zap.Logger
	driver NativeDriver
}

func NewPacketConn(logger *zap.Logger, lmaddr ma.Multiaddr, driver NativeDriver) (*PacketConn, error) {
	laddr, err := manet.ToNetAddr(lmaddr)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &PacketConn{
		logger: logger,
		lmaddr: lmaddr,
		laddr:  laddr,
		driver: driver,
		ctx:    ctx,
		cancel: cancel,
		cmsg:   make(chan *RcvMsg),
	}, nil
}

func (c *PacketConn) Pipe(p []byte, addr net.Addr) (err error) {
	msg := &RcvMsg{
		Payload: p,
		Addr:    addr,
	}

	select {
	case <-c.ctx.Done():
		err = c.ctx.Err()
	case c.cmsg <- msg:
	}

	return
}

// ReadFrom reads a packet from the connection,
// copying the payload into p. It returns the number of
// bytes copied into p and the return address that
// was on the packet.
// It returns the number of bytes read (0 <= n <= len(p))
// and any error encountered. Callers should always process
// the n > 0 bytes returned before considering the error err.
// ReadFrom can be made to time out and return an error after a
// fixed time limit; see SetDeadline and SetReadDeadline.
func (c *PacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	var msg *RcvMsg
	select {
	case <-c.ctx.Done():
		return 0, nil, c.ctx.Err()
	case msg = <-c.cmsg:
	}

	copy(p, msg.Payload)
	return len(msg.Payload), msg.Addr, nil
}

func (c *PacketConn) ReadFromMultiaddr(b []byte) (int, ma.Multiaddr, error) {
	n, addr, err := c.ReadFrom(b)
	maddr, _ := manet.FromNetAddr(addr)
	return n, maddr, err
}

// WriteTo writes a packet with payload p to addr.
// WriteTo can be made to time out and return an Error after a
// fixed time limit; see SetDeadline and SetWriteDeadline.
// On packet-oriented connections, write timeouts are rare.
func (c *PacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	if !c.driver.SendToPeer(addr.String(), p) {
		c.logger.Debug("Conn.Write failed")
		return 0, fmt.Errorf("error: Conn.Write failed: native write failed")
	}

	return len(p), nil
}

func (c *PacketConn) WriteToMultiaddr(b []byte, maddr ma.Multiaddr) (int, error) {
	addr, err := manet.ToNetAddr(maddr)
	if err != nil {
		return 0, err
	}
	return c.WriteTo(b, addr)
}

// LocalMultiaddr returns the bound local Multiaddr.
func (c *PacketConn) LocalMultiaddr() ma.Multiaddr {
	return c.lmaddr
}

// LocalAddr returns the local network address.
func (c *PacketConn) LocalAddr() net.Addr {
	return c.laddr
}

// Close closes the connection.
// Any blocked ReadFrom or WriteTo operations will be unblocked and return errors.
func (c *PacketConn) Close() error {
	c.cancel()
	return nil
}

// @TODO: setup setdeadine

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail instead of blocking. The deadline applies to all future
// and pending I/O, not just the immediately following call to
// Read or Write. After a deadline has been exceeded, the
// connection can be refreshed by setting a deadline in the future.
//
// If the deadline is exceeded a call to Read or Write or to other
// I/O methods will return an error that wraps os.ErrDeadlineExceeded.
// This can be tested using errors.Is(err, os.ErrDeadlineExceeded).
// The error's Timeout method will return true, but note that there
// are other possible errors for which the Timeout method will
// return true even if the deadline has not been exceeded.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful ReadFrom or WriteTo calls.
//
// A zero value for t means I/O operations will not time out.
func (c *PacketConn) SetDeadline(t time.Time) error {
	return nil
}

// @TODO: setup setdeadine

// SetReadDeadline sets the deadline for future ReadFrom calls
// and any currently-blocked ReadFrom call.
// A zero value for t means ReadFrom will not time out.
func (c *PacketConn) SetReadDeadline(t time.Time) error {
	return nil
}

// @TODO: setup setdeadine

// SetWriteDeadline sets the deadline for future WriteTo calls
// and any currently-blocked WriteTo call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means WriteTo will not time out.
func (c *PacketConn) SetWriteDeadline(t time.Time) error {
	return nil
}
