// this file is just a precaution to avoid golang to process this directory

module berty.tech/berty/tool/v2

go 1.14

require (
	berty.tech/berty/v2 v2.0.0
	github.com/libp2p/go-libp2p v0.13.0
	github.com/libp2p/go-libp2p-circuit v0.4.0
	github.com/libp2p/go-libp2p-core v0.8.0
	github.com/libp2p/go-libp2p-quic-transport v0.10.0
	github.com/mdp/qrterminal/v3 v3.0.0
	github.com/multiformats/go-multiaddr v0.3.1
	github.com/oklog/run v1.1.0
	github.com/peterbourgon/ff/v3 v3.0.0
	go.uber.org/zap v1.16.0
    github.com/libp2p/go-libp2p-rendezvous v0.0.0-20180418151804-b7dd840ce441
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	moul.io/u v1.20.0
)

replace (
    	github.com/libp2p/go-libp2p-rendezvous => github.com/berty/go-libp2p-rendezvous v0.0.0-20201028141428-5b2e7e8ff19a // use berty fork of go-libp2p-rendezvous
)

replace berty.tech/berty/v2 => ../
