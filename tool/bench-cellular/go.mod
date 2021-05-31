module berty.tech/berty/v2/tool/bench-cellular

go 1.15

require (
	github.com/ipfs/go-log v1.0.4
	github.com/libp2p/go-libp2p v0.13.0
	github.com/libp2p/go-libp2p-core v0.8.5
	github.com/libp2p/go-libp2p-kad-dht v0.11.1
	github.com/libp2p/go-libp2p-peer v0.2.0
	github.com/libp2p/go-libp2p-peerstore v0.2.7
	github.com/libp2p/go-libp2p-quic-transport v0.10.0
	github.com/libp2p/go-libp2p-routing v0.1.0
	github.com/libp2p/go-tcp-transport v0.2.1
	github.com/multiformats/go-multiaddr v0.3.1
	github.com/peterbourgon/ff/v3 v3.0.0
)

replace (
	github.com/libp2p/go-libp2p => github.com/gfanton/go-libp2p v0.5.3-0.20210526094535-a667b7d55d56
	github.com/libp2p/go-libp2p-circuit => github.com/libp2p/go-libp2p-circuit v0.4.1-0.20210309082447-90d67900a8fe // relay v2
)
