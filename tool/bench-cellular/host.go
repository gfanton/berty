package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	mrand "math/rand"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	quict "github.com/libp2p/go-libp2p-quic-transport"
	ma "github.com/multiformats/go-multiaddr"
)

var bertyRelays = []string{
	"/ip4/51.159.21.214/udp/4040/quic/p2p/QmdT7AmhhnbuwvCpa5PH1ySK9HJVB82jr3fo1bxMxBPW6p",
	"/ip4/51.15.25.224/udp/4040/quic/p2p/12D3KooWHhDBv6DJJ4XDWjzEXq6sVNEs6VuxsV1WyBBEhPENHzcZ",
	// "/ip4/51.75.127.200/udp/4141/quic/p2p/12D3KooWPwRwwKatdy5yzRVCYPHib3fntYgbFB4nqrJPHWAqXD7z",
}

func loadRelaysAddrs() []*peer.AddrInfo {
	addrs := make([]*peer.AddrInfo, len(bertyRelays))
	for i, addr := range bertyRelays {
		a, err := ma.NewMultiaddr(addr)
		if err != nil {
			log.Printf("error: can't parse Multiaddr: %s: %v\n", addr, err)
			continue
		}

		pi, err := peer.AddrInfoFromP2pAddr(a)
		if err != nil {
			log.Printf("error: can't parse AddrInfo: %v\n", err)
			continue
		}
		addrs[i] = pi
	}

	return addrs
}

func createBasicHost(seed int64, port int, autorelay, insecure, quic, ip6 bool) (host.Host, error) {
	ctx := context.Background()

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

	relays := loadRelaysAddrs()

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
	}

	if autorelay {
		opts = append(
			opts,
			libp2p.EnableAutoRelay(),
			// libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			// 	return dht.New(context.Background(), h, dht.Mode(dht.ModeClient), dht.BootstrapPeers(dht.GetDefaultBootstrapPeerAddrInfos()...))
			// }),
		)

		var staticRelays []peer.AddrInfo

		for _, pi := range relays {
			staticRelays = append(staticRelays, *pi)
			log.Printf("ADDED %v\n", *pi)
		}

		opts = append(opts, libp2p.EnableRelay(), libp2p.EnableAutoRelay(), libp2p.StaticRelays(staticRelays))
	} else {
		opts = append(opts, libp2p.EnableRelay())
	}

	if quic {
		opts = append(opts, libp2p.Transport(quict.NewTransport))
	}

	if insecure {
		opts = append(opts, libp2p.NoSecurity)
	}

	h, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	for _, pi := range relays {
		h.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.PermanentAddrTTL)

		log.Printf("CONNECTING TO RELAY: %v\n", *pi)
		if err := h.Connect(ctx, *pi); err != nil {
			return nil, fmt.Errorf("unalbe to connect to relay addr")
		}

		log.Printf("CONNECTED TO RELAY: %v\n", *pi)
	}

	return h, nil
}
