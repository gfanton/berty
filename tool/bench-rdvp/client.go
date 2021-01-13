package main

import (
	"context"
	"crypto/rand"
	"fmt"

	libp2p "github.com/libp2p/go-libp2p"
	p2p_circuit "github.com/libp2p/go-libp2p-circuit"
	p2p_crypto "github.com/libp2p/go-libp2p-core/crypto"
	host "github.com/libp2p/go-libp2p-core/host"
	p2p_peer "github.com/libp2p/go-libp2p-core/peer"
	p2p_ps "github.com/libp2p/go-libp2p-core/peerstore"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	p2p_quic "github.com/libp2p/go-libp2p-quic-transport"
	p2p_rp "github.com/libp2p/go-libp2p-rendezvous"
	"go.uber.org/zap"

	ma "github.com/multiformats/go-multiaddr"
)

type ClientOpts struct {
	Listeners      []ma.Multiaddr
	Logger         *zap.Logger
	Identity       p2p_crypto.PrivKey
	RendezVousPeer p2p_peer.AddrInfo
}

func (o *ClientOpts) fillDefault(ctx context.Context) error {
	if o.Logger == nil {
		o.Logger = zap.NewNop()
	}

	if o.RendezVousPeer.ID == p2p_peer.ID("") || len(o.RendezVousPeer.Addrs) == 0 {
		return fmt.Errorf("no rendezvous peer provided")
	}

	if o.Identity == nil {
		priv, _, err := p2p_crypto.GenerateKeyPairWithReader(p2p_crypto.RSA, 2048, rand.Reader)
		if err != nil {
			return err
		}
		o.Identity = priv
	}

	if o.Listeners == nil {
		o.Listeners = []ma.Multiaddr{
			ma.StringCast("/ip4/0.0.0.0/tcp/0"),
			ma.StringCast("/ip6/::/tcp/0"),
			ma.StringCast("/ip4/0.0.0.0/udp/0/quic"),
			ma.StringCast("/ip6/::/udp/0/quic"),
		}
	}

	return nil
}

type Client struct {
	host.Host

	// rpc    p2p_rp.RendezvousClient
	logger *zap.Logger
	ps     *pubsub.PubSub
}

func NewClientHost(ctx context.Context, opts *ClientOpts) (*Client, error) {
	if err := opts.fillDefault(ctx); err != nil {
		return nil, fmt.Errorf("invalid opts: %w", err)
	}

	logger := opts.Logger

	h, err := libp2p.New(ctx,
		// default tpt + quic
		libp2p.DefaultTransports,
		libp2p.Transport(p2p_quic.NewTransport),

		// Nat & Relay service
		// libp2p.EnableNATService(),
		libp2p.EnableRelay(p2p_circuit.OptHop),

		// swarm listeners
		libp2p.ListenAddrs(opts.Listeners...),

		// identity
		libp2p.Identity(opts.Identity),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to configure host: %w", err)
	}

	logger.Debug("connecting to rdvp",
		zap.String("ID", opts.RendezVousPeer.ID.String()),
		zap.Strings("addrs", convertMaddrToString(opts.RendezVousPeer.Addrs)))

	if err := h.Connect(ctx, opts.RendezVousPeer); err != nil {
		return nil, fmt.Errorf("unable to connect to rdvp: %w", err)
	}

	h.Peerstore().AddAddrs(opts.RendezVousPeer.ID, opts.RendezVousPeer.Addrs, p2p_ps.PermanentAddrTTL)

	// create rendez vous client
	rp := p2p_rp.NewRendezvousDiscovery(h, opts.RendezVousPeer.ID)
	ps, err := pubsub.NewGossipSub(ctx, h, pubsub.WithDiscovery(rp))
	if err != nil {
		return nil, fmt.Errorf("unable to configure pubusb: %w", err)
	}

	return &Client{
		Host: h,

		// rpc:    rp,
		logger: logger,
		ps:     ps,
	}, err
}

func (c *Client) Join(ctx context.Context, ns string) error {
	sub, err := c.ps.Join(ns)
	if err != nil {
		return err
	}

	hevt, err := sub.EventHandler()
	if err != nil {
		return err
	}

	for {
		evt, err := hevt.NextPeerEvent(ctx)
		if err != nil {
			return err
		}

		switch evt.Type {
		case pubsub.PeerJoin:
			addrs := c.Peerstore().Addrs(evt.Peer)
			c.logger.Info("peer joined", zap.String("id", evt.Peer.String()), zap.Strings("addrs", convertMaddrToString(addrs)))
		case pubsub.PeerLeave:
			addrs := c.Peerstore().Addrs(evt.Peer)
			c.logger.Info("peer joined", zap.String("id", evt.Peer.String()), zap.Strings("addrs", convertMaddrToString(addrs)))

		}
	}
}

func convertMaddrToString(ma []ma.Multiaddr) []string {
	addrs := make([]string, len(ma))
	for i, m := range ma {
		addrs[i] = m.String()
	}
	return addrs
}
