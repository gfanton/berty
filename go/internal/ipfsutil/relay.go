package ipfsutil

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-circuit/v2/client"
	discovery "github.com/libp2p/go-libp2p-discovery"
	discutil "github.com/libp2p/go-libp2p-discovery"
	ma "github.com/multiformats/go-multiaddr"

	host "github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"go.uber.org/zap"
)

const (
	RelaysMaximumLimit = 100
)

type AutoRelayV2Opts struct {
	Logger           *zap.Logger
	StaticRelays     []ma.Multiaddr
	BackoffConnector *discutil.BackoffConnector
}

func (o *AutoRelayV2Opts) applyDefault(h host.Host) error {
	var err error

	if o.BackoffConnector == nil {
		bdisc := discovery.NewFixedBackoff(time.Minute)

		fmt.Printf("peer: %s\n", h.ID().Pretty())
		o.BackoffConnector, err = discovery.NewBackoffConnector(h, 10, time.Minute*10, bdisc)
		if err != nil {
			return err
		}
	}

	return err
}

type AutoRelayV2 struct {
	rootCtx   context.Context
	cancelCtx context.CancelFunc
	host      host.Host
	logger    *zap.Logger

	bconn *discutil.BackoffConnector
	cpeer chan peer.AddrInfo
	conce sync.Once

	muRelays sync.RWMutex
	relays   map[peer.ID]struct{}

	muResvs sync.RWMutex
	resvs   map[peer.ID]*client.Reservation
}

func NewAutoRelayV2(h host.Host, opts *AutoRelayV2Opts) (*AutoRelayV2, error) {
	if err := opts.applyDefault(h); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	logger := opts.Logger
	cpeer := make(chan peer.AddrInfo, 100)

	fmt.Printf("backoff ok: %v\n", opts.BackoffConnector == nil)
	// start backoff connector
	go opts.BackoffConnector.Connect(ctx, cpeer)
	ar := &AutoRelayV2{
		rootCtx:   ctx,
		cancelCtx: cancel,

		logger: logger,

		host:   h,
		bconn:  opts.BackoffConnector,
		cpeer:  cpeer,
		relays: make(map[peer.ID]struct{}),
		resvs:  make(map[peer.ID]*client.Reservation),
	}

	for _, maddr := range opts.StaticRelays {
		ai, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			logger.Warn("unable to parse static relay", zap.String("relay", ai.ID.Pretty()))
		} else {
			ar.registerRelay(*ai)
		}

	}

	return ar, nil
}

func (ar *AutoRelayV2) registerRelay(ai peer.AddrInfo) {
	ar.muRelays.Lock()
	if _, ok := ar.relays[ai.ID]; !ok {
		ar.logger.Debug("registering new relay", zap.String("relay", ai.ID.Pretty()))
		ar.relays[ai.ID] = struct{}{}
	}
	ar.muRelays.Unlock()

	if ar.host.Network().Connectedness(ai.ID) != network.Connected {
		fmt.Printf("peer: %v\n", ai)
		ar.cpeer <- ai
	}
}

func (ar *AutoRelayV2) reserve(ai peer.AddrInfo) {
	ar.muResvs.RLock()
	_, ok := ar.resvs[ai.ID]
	ar.muResvs.RUnlock()

	if ok {
		return
	}

	resv, err := client.Reserve(context.Background(), ar.host, ai)
	if err != nil {
		ar.logger.Warn("unable to reserve to relay", zap.String("relay", ai.ID.Pretty()))
	}

	ar.muResvs.Lock()
	ar.resvs[ai.ID] = resv
	ar.muResvs.Unlock()

	ar.logger.Info("successfully reserved relay server", zap.String("relay", ai.ID.Pretty()))
}

func (ar *AutoRelayV2) Close() {
	ar.conce.Do(func() {
		close(ar.cpeer)
	})
}

// Notifee
func (ar *AutoRelayV2) Listen(network.Network, ma.Multiaddr)      {}
func (ar *AutoRelayV2) ListenClose(network.Network, ma.Multiaddr) {}
func (ar *AutoRelayV2) Connected(n network.Network, c network.Conn) {
	go func() {
		ar.muRelays.RLock()
		if _, ok := ar.relays[c.RemotePeer()]; ok {
			addrs := ar.host.Peerstore().Addrs(c.RemotePeer())
			peer := peer.AddrInfo{
				ID:    c.RemotePeer(),
				Addrs: addrs,
			}

			go ar.reserve(peer)
		}
		ar.muRelays.RUnlock()
	}()
}

func (ar *AutoRelayV2) Disconnected(net network.Network, c network.Conn) {

}

func (ar *AutoRelayV2) OpenedStream(network.Network, network.Stream) {}
func (ar *AutoRelayV2) ClosedStream(network.Network, network.Stream) {}
