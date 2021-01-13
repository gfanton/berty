package main

import (
	"context"
	"crypto/rand"
	"fmt"

	libp2p "github.com/libp2p/go-libp2p"
	p2p_circuit "github.com/libp2p/go-libp2p-circuit"
	p2p_crypto "github.com/libp2p/go-libp2p-core/crypto"
	host "github.com/libp2p/go-libp2p-core/host"
	p2p_quic "github.com/libp2p/go-libp2p-quic-transport"
	p2p_rp "github.com/libp2p/go-libp2p-rendezvous"
	p2p_rpdb "github.com/libp2p/go-libp2p-rendezvous/db/sqlite"
	"go.uber.org/zap"

	ma "github.com/multiformats/go-multiaddr"
)

type RDVPOpts struct {
	Listeners []ma.Multiaddr
	Logger    *zap.Logger
	Identity  p2p_crypto.PrivKey
	DB        *p2p_rpdb.DB
}

func (o *RDVPOpts) fillDefault(ctx context.Context) error {
	if o.Logger == nil {
		o.Logger = zap.NewNop()
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

	if o.DB == nil {
		db, err := p2p_rpdb.OpenDB(ctx, ":memory:")
		if err != nil {
			return err
		}

		o.DB = db
	}

	return nil
}

func NewRDVPHost(ctx context.Context, opts *RDVPOpts) (host.Host, error) {
	if err := opts.fillDefault(ctx); err != nil {
		return nil, fmt.Errorf("unable to apply default opts: %w", err)
	}

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

	// start service
	_ = p2p_rp.NewRendezvousService(h, opts.DB)

	return h, err
}
