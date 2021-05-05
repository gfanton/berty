package tinder

import (
	"context"
	"sync"
	"time"

	p2p_discovery "github.com/libp2p/go-libp2p-core/discovery"
	p2p_event "github.com/libp2p/go-libp2p-core/event"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	p2p_peer "github.com/libp2p/go-libp2p-core/peer"
	"go.uber.org/zap"
)

// discovery monitor will send a event everytime we found/advertise a peer
type discoveryMonitor struct {
	host    host.Host
	logger  *zap.Logger
	disc    p2p_discovery.Discovery
	emitter p2p_event.Emitter

	muConnMonitor sync.Mutex
	connMonitor   map[p2p_peer.ID]struct{}
}

// Advertise advertises a service
func (d *discoveryMonitor) Advertise(ctx context.Context, ns string, opts ...p2p_discovery.Option) (time.Duration, error) {
	ttl, err := d.disc.Advertise(ctx, ns, opts...)
	if err == nil {
		d.Emit(&EvtDriverMonitor{
			EventType: TypeEventMonitorAdvertise,
			AddrInfo: p2p_peer.AddrInfo{
				ID:    d.host.ID(),
				Addrs: d.host.Addrs(),
			},
			Topic: ns,
		})
	}

	return ttl, err
}

// FindPeers discovers peers providing a service
func (d *discoveryMonitor) FindPeers(ctx context.Context, ns string, opts ...p2p_discovery.Option) (<-chan p2p_peer.AddrInfo, error) {
	c, err := d.disc.FindPeers(ctx, ns, opts...)
	if err != nil {
		return nil, err
	}

	retc := make(chan p2p_peer.AddrInfo)
	go func() {
		for p := range c {
			d.muConnMonitor.Lock()
			if _, ok := d.connMonitor[p.ID]; ok {
				d.muConnMonitor.Unlock()
				continue
			}

			d.connMonitor[p.ID] = struct{}{}
			d.muConnMonitor.Unlock()

			go d.BackoffConnectorLog(p)

			d.Emit(&EvtDriverMonitor{
				EventType: TypeEventMonitorFoundPeer,
				AddrInfo:  p,
				Topic:     ns,
			})

			retc <- p
		}

		close(retc)
	}()

	return retc, err
}

func (d *discoveryMonitor) Emit(evt *EvtDriverMonitor) {
	if err := d.emitter.Emit(*evt); err != nil {
		d.logger.Warn("unable to emit `EvtDriverMonitor`", zap.Error(err))
	}
}

func (d *discoveryMonitor) BackoffConnectorLog(p p2p_peer.AddrInfo) {
	connectedness := d.host.Network().Connectedness(p.ID)

	for connectedness != network.Connected {
		time.Sleep(time.Second * 2)

		d.logger.Debug("tinder connecting", zap.String("peerID", p.ID.Pretty()), zap.Any("addrs", p.Addrs))

		if err := d.host.Connect(context.Background(), p); err != nil {
			d.logger.Warn("tinder unable to connect", zap.String("peerID", p.ID.Pretty()), zap.Any("addrs", p.Addrs), zap.Error(err))
		}

		connectedness = d.host.Network().Connectedness(p.ID)
	}

	conns := d.host.Network().ConnsToPeer(p.ID)
	maddrs := make([]string, len(conns))
	for i, conn := range conns {
		maddrs[i] = conn.RemoteMultiaddr().String()
	}
	d.logger.Debug("tinder connected!", zap.String("peerID", p.ID.Pretty()), zap.Strings("addrs", maddrs))

	d.muConnMonitor.Lock()
	delete(d.connMonitor, p.ID)
	d.muConnMonitor.Unlock()

}
