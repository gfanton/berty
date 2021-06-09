package mini

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"berty.tech/berty/v2/go/internal/notify"
	quict "github.com/libp2p/go-libp2p-quic-transport"
	tcpt "github.com/libp2p/go-tcp-transport"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-circuit/v2/client"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
)

const ServerRelayID = "/relay/msg/1.0.0"
const UploadRelayID = "/relay/upload/1.0.0"
const DownloadRelayID = "/relay/download/1.0.0"

// type relaySession struct {
// }

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

		handleStreamCommand(s, v)
	})

	v.v.host.SetStreamHandler(UploadRelayID, func(s network.Stream) {
		v.messages.Append(&historyMessage{
			messageType: messageTypeMeta,
			payload:     []byte("upload from: " + s.Conn().RemotePeer().ShortString()),
		})

		uploadHandler(v, s)
	})

	v.rpeerid = pi.ID
	v.relaymode = true
	v.messages.Append(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte("entering in relay server mode"),
	})

	return nil
}

func clientRelay(ctx context.Context, v *groupView, addr string) error {
	h, err := libp2p.New(ctx,
		libp2p.EnableRelay(),
		libp2p.Transport(quict.NewTransport),
		libp2p.Transport(tcpt.NewTCPTransport),
	)

	if v.v.host == nil {
		return fmt.Errorf("no host given")
	}

	v.messages.Append(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte(fmt.Sprintf("resolving addr: %s", addr)),
	})

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

	v.v.host.SetStreamHandler(UploadRelayID, func(s network.Stream) {
		v.messages.Append(&historyMessage{
			messageType: messageTypeMeta,
			payload:     []byte("upload from: " + s.Conn().RemotePeer().ShortString()),
		})

		uploadHandler(v, s)
	})

	v.messages.Append(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte(fmt.Sprintf("connecting to remote peer %s", peerid.ShortString())),
	})

	s, err := v.v.host.NewStream(network.WithUseTransient(ctx, "mini/chat"), peerid, ServerRelayID)
	if err != nil {
		return fmt.Errorf("unable to start stream: %w", err)
	}

	handleStreamCommand(s, v)

	v.nrelay = notify.New(&sync.Mutex{})

	v.rpeerid = peerid
	v.relaymode = true
	v.messages.Append(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte("entering in relay client mode"),
	})

	return nil
}

func (v *groupView) relayCommandParser(ctx context.Context, input string) error {
	if input == "/exit" {
		exitRelayMode(v, nil)
	} else if strings.HasPrefix(input, "/send") {
		v.messages.Append(&historyMessage{
			messageType: messageTypeMeta,
			payload:     []byte("start send"),
		})

		var size int64
		if _, err := fmt.Sscanf(input, "/send %d", &size); err != nil {
			return fmt.Errorf("unable to parse size: %w", err)
		}

		v.messages.Append(&historyMessage{
			messageType: messageTypeMeta,
			payload:     []byte("sending %d"),
		})

		s, err := v.v.host.NewStream(network.WithUseTransient(ctx, "mini/upload"), v.rpeerid, UploadRelayID)
		if err != nil {
			return fmt.Errorf("unable to make stream, is the remote peer using /recv ?: %w", err)
		}
		defer s.CloseWrite()

		w := bufio.NewWriter(s)
		if _, err := w.WriteString(fmt.Sprintf("%d\n", size)); err != nil {
			return fmt.Errorf("write error: %w", err)
		}

		v.messages.Append(&historyMessage{
			messageType: messageTypeMeta,
			payload:     []byte(fmt.Sprintf("start upload for: %d", size)),
		})

		buf := make([]byte, size)
		n, err := w.Write(buf)
		if err != nil {
			return fmt.Errorf("write error: %w", err)
		}

		endresult := fmt.Sprintf("write %d/%d", n, size)
		v.messages.Append(&historyMessage{
			messageType: messageTypeMeta,
			payload:     []byte(endresult),
		})
		v.messages.Append(&historyMessage{
			messageType: messageTypeMeta,
			payload:     []byte("send done"),
		})
	} else {
		v.sendMsg(input)
		v.messages.Append(&historyMessage{
			messageType: messageTypeRelay,
			payload:     []byte(input),
		})
	}

	return nil
}

func handleStreamCommand(s network.Stream, v *groupView) {
	go func() {
		err := relayReadString(s, v)
		exitRelayMode(v, err)
	}()

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := relayWriteString(ctx, s, v)
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

func relayReadString(s network.Stream, v *groupView) error {
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

func uploadHandler(v *groupView, s network.Stream) {
	defer s.Reset()
	r := bufio.NewReader(s)

	v.messages.Append(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte("start recv"),
	})

	var msg string
	msg, err := r.ReadString('\n')
	if err != nil {
		v.messages.Append(&historyMessage{
			messageType: messageTypeError,
			payload:     []byte(fmt.Sprintf("error: %s", err.Error())),
		})
		return
	}

	var size int64
	if _, err = fmt.Sscanf(msg, "%d", &size); err != nil {
		v.messages.Append(&historyMessage{
			messageType: messageTypeError,
			payload:     []byte(fmt.Sprintf("error: %s", err.Error())),
		})
		return
	}

	v.messages.Append(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte(fmt.Sprintf("start download for: %d", size)),
	})

	buf, err := io.ReadAll(s)
	if err != nil && err != io.EOF {
		v.messages.Append(&historyMessage{
			messageType: messageTypeError,
			payload:     []byte(fmt.Sprintf("error: %s", err.Error())),
		})
		return
	}

	endresult := fmt.Sprintf("read %d/%d", len(buf), size)
	v.messages.Append(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte(endresult),
	})
}

func (v *groupView) sendMsg(msg string) {
	v.nrelay.L.Lock()
	v.cmsg = msg
	// broadcast to waiter that our msg has been set
	v.nrelay.Broadcast()
	v.nrelay.L.Unlock()
}

func relayWriteString(ctx context.Context, s network.Stream, v *groupView) error {
	defer s.Reset()

	v.nrelay.L.Lock()
	defer v.nrelay.L.Unlock()

	w := bufio.NewWriter(s)
	for {
		if ok := v.nrelay.Wait(ctx); !ok {
			return ctx.Err()
		}

		// v.cmsg should be fill with our new msg
		if _, err := w.WriteString(fmt.Sprintf("%s\n", v.cmsg)); err != nil {
			return err
		}

		if err := w.Flush(); err != nil {
			return err
		}
	}
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
