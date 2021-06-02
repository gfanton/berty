package mini

import (
	"bufio"
	"context"
	"fmt"
	"sync"
	"time"

	"berty.tech/berty/v2/go/internal/notify"
	"github.com/libp2p/go-libp2p-circuit/v2/client"
	host "github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
)

const ServerRelayID = "/relay/test/1.0.0"

func serverRelay(ctx context.Context, v *groupView, addr string) error {
	if v.v.host == nil {
		return fmt.Errorf("no host given")
	}

	maddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		return fmt.Errorf("can't parse Multiaddr: %w", err)
	}

	pi, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return fmt.Errorf("can't parse AddrInfo: %w", err)
	}

	circuitaddr := ma.StringCast("/p2p-circuit/p2p/" + v.v.host.ID().Pretty())

	v.messages.Append(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte(fmt.Sprintf("connecting to %s", pi.ID.ShortString())),
	})

	if err := v.v.host.Connect(ctx, *pi); err != nil {
		return fmt.Errorf("unable to connect to relay")
	}

	v.messages.Append(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte("making a reservation to " + pi.ID.ShortString()),
	})

	rsvp, err := client.Reserve(ctx, v.v.host, *pi)
	if err != nil {
		return fmt.Errorf("unable to make a reservation to `%s`: %w", pi.ID.ShortString(), err)
	}

	v.messages.Append(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte(fmt.Sprintf("made a reservation on `%s`, expire in %fs", pi.ID.ShortString(), time.Until(rsvp.Expiration).Minutes())),
	})

	fulladdr := maddr.Encapsulate(circuitaddr)
	endmsg := fmt.Sprintf("connect to this peer with `/relay client %s`", fulladdr.String())
	copyToClipboard(v, fulladdr.String())
	v.messages.Append(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte(endmsg),
	})

	v.nrelay = notify.New(&sync.Mutex{})
	v.v.host.SetStreamHandler(ServerRelayID, func(s network.Stream) {
		v.messages.Append(&historyMessage{
			messageType: messageTypeMeta,
			payload:     []byte("connected: " + s.Conn().RemotePeer().ShortString()),
		})

		handleStream(s, v)
	})

	v.relaymode = true
	v.messages.Append(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte("entering in relay server mode"),
	})

	return nil
}

func clientRelay(ctx context.Context, v *groupView, addr string) error {
	if v.v.host == nil {
		return fmt.Errorf("no host given")
	}

	maddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		return fmt.Errorf("can't parse Multiaddr: %w", err)
	}

	relayaddr, peerid, err := getRelayAddrs(maddr)
	if err != nil {
		return fmt.Errorf("can't extract relay addr: %w", err)
	}

	v.messages.Append(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte(fmt.Sprintf("adding to %s ps: %s", peerid.Pretty(), maddr.String())),
	})

	// add addr to our peerstore
	v.v.host.Peerstore().AddAddr(peerid, maddr, peerstore.AddressTTL)

	pi, err := peer.AddrInfoFromP2pAddr(relayaddr)
	if err != nil {
		return fmt.Errorf("can't parse relayaddr: %w", err)
	}

	v.messages.Append(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte(fmt.Sprintf("connecting to relay %s", pi.ID.ShortString())),
	})

	if err := v.v.host.Connect(ctx, *pi); err != nil {
		return fmt.Errorf("unable to connect to relay: %w", err)
	}

	v.messages.Append(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte(fmt.Sprintf("connecting to remote peer %s", peerid.ShortString())),
	})

	s, err := v.v.host.NewStream(network.WithUseTransient(ctx, "mini/chat"), peerid, ServerRelayID)
	if err != nil {
		return fmt.Errorf("unable to start stream: %w", err)
	}

	v.nrelay = notify.New(&sync.Mutex{})
	handleStream(s, v)

	v.relaymode = true
	v.messages.Append(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte("entering in relay client mode"),
	})

	return nil
}

func (v *groupView) relayCommandParser(ctx context.Context, input string) error {
	switch input {
	case "/exit":
		exitRelayMode(v, nil)
	default:
		v.nrelay.L.Lock()
		v.cmsg = input
		v.nrelay.Broadcast()
		v.nrelay.L.Unlock()

		v.messages.Append(&historyMessage{
			messageType: messageTypeRelay,
			payload:     []byte(input),
		})

	}

	return nil
}

func handleStream(s network.Stream, v *groupView) {
	go func() {
		err := relayReadData(s, v)
		exitRelayMode(v, err)
	}()

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := relayWriteData(ctx, s, v)
		exitRelayMode(v, err)
	}()
}

func exitRelayMode(v *groupView, err error) {
	if err != nil {
		v.messages.Append(&historyMessage{
			messageType: messageTypeError,
			payload:     []byte(fmt.Sprintf("error: %s", err.Error())),
		})
	}

	v.messages.Append(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte("exiting relay mode"),
	})

	v.relaymode = false
}

func relayReadData(s network.Stream, v *groupView) error {
	defer s.Reset()

	sender := fmt.Sprintf("%.8s", s.Conn().RemotePeer().Pretty())
	r := bufio.NewReader(s)
	for {
		msg, err := r.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading from buffer: %w", err)
		}

		v.messages.Append(&historyMessage{
			sender:      []byte(sender),
			messageType: messageTypeRelay,
			payload:     []byte(msg),
		})
	}
}

func relayWriteData(ctx context.Context, s network.Stream, v *groupView) error {
	defer s.Reset()

	v.nrelay.L.Lock()
	defer v.nrelay.L.Unlock()

	w := bufio.NewWriter(s)
	for {
		if ok := v.nrelay.Wait(ctx); !ok {
			return ctx.Err()
		}

		if _, err := w.WriteString(fmt.Sprintf("%s\n", v.cmsg)); err != nil {
			return err
		}

		if err := w.Flush(); err != nil {
			return err
		}
	}
}

func setupRelayHandler(h host.Host) error {

	if h == nil {
		return fmt.Errorf("no host given")
	}

	return nil
}

func getRelayAddrs(m ma.Multiaddr) (ma.Multiaddr, peer.ID, error) {
	relayaddr, dest := ma.SplitFunc(m, func(c ma.Component) bool {
		return c.Protocol().Code == ma.P_CIRCUIT
	})

	if !manet.IsThinWaist(relayaddr) {
		return nil, "", fmt.Errorf("maddr is not thin waist: %s", m.String())
	}

	pid, err := dest.ValueForProtocol(ma.P_P2P)
	if err != nil {
		return nil, "", fmt.Errorf("unable to get peer id from `%s`: %w", dest.String(), err)
	}

	peerid, _ := peer.IDB58Decode(pid)
	return relayaddr, peerid, nil
}
