package proximitytransport

import (
	"context"
	"fmt"
	"net"

	peer "github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	kcp "github.com/xtaci/kcp-go/v5"
)

// Listener is a tpt.Listener.
var _ manet.Listener = (*Listener)(nil)

// Listener is an interface closely resembling the net.Listener interface. The
// only real difference is that Accept() returns Conn's of the type in this
// package, and also exposes a Multiaddr method as opposed to a regular Addr
// method.
type Listener struct {
	net.Listener

	lmaddr ma.Multiaddr
}

// connReq holds data necessary for inbound conn creation.
type connReq struct {
	remoteMa  ma.Multiaddr
	remotePID peer.ID
}

var (
	DataNShard_TODO                 = 0
	ParityShard_TODO                = 0
	BlockCrypt_TODO  kcp.BlockCrypt = nil
)

// newListener starts the native driver then returns a new Listener.
func NewListener(ctx context.Context, pc manet.PacketConn) (manet.Listener, error) {
	kcpl, err := kcp.ServeConn(BlockCrypt_TODO, DataNShard_TODO, ParityShard_TODO, pc)
	if err != nil {
		return nil, err
	}

	return &Listener{
		Listener: kcpl,
		lmaddr:   pc.LocalMultiaddr(),
	}, nil
}

func (l *Listener) Accept() (manet.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	rmaddr, err := manet.FromNetAddr(c.RemoteAddr())
	if err != nil {
		return nil, fmt.Errorf("unable to accept conn, cannot convert addr to maddr: %w", err)
	}

	return NewConn(c, l.lmaddr, rmaddr), nil
}

func (l *Listener) Multiaddr() ma.Multiaddr {
	return l.lmaddr
}
