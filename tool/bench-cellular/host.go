package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	mrand "math/rand"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	quict "github.com/libp2p/go-libp2p-quic-transport"
)

func createBasicHost(seed int64, port int, insecure, quic, ip6 bool) (host.Host, error) {
	var r io.Reader
	if seed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(seed))
	}

	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}

	var listener libp2p.Option
	if ip6 {
		if quic {
			listener = libp2p.ListenAddrStrings(fmt.Sprintf("/ip6/::/udp/%d/quic", port))
		} else {
			listener = libp2p.ListenAddrStrings(fmt.Sprintf("/ip6/::/tcp/%d", port))
		}
	} else {
		if quic {
			listener = libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic", port))
		} else {
			listener = libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
		}
	}
	opts := []libp2p.Option{
		listener,
		libp2p.Identity(priv),
		libp2p.DisableRelay(),
	}

	if quic {
		opts = append(opts, libp2p.Transport(quict.NewTransport))
	}

	if insecure {
		opts = append(opts, libp2p.NoSecurity)
	}

	return libp2p.New(context.Background(), opts...)
}
