package ipfsutil

import (
	"context"
	"crypto/sha1"
	"fmt"
	"sync"
	"time"

	ipfs_interface "github.com/ipfs/interface-go-ipfs-core"
	ipfs_iopts "github.com/ipfs/interface-go-ipfs-core/options"
	p2p_disc "github.com/libp2p/go-libp2p-core/discovery"
	p2p_host "github.com/libp2p/go-libp2p-core/host"
	p2p_peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	p2p_pubsub "github.com/libp2p/go-libp2p-pubsub"

	"go.uber.org/zap"
)

type PubSubAPI struct {
	*p2p_pubsub.PubSub

	host   p2p_host.Host
	disc   p2p_disc.Discovery
	logger *zap.Logger

	muTopics sync.RWMutex
	topics   map[string]*p2p_pubsub.Topic
}

func NewPubSubAPI(ctx context.Context, logger *zap.Logger, disc p2p_disc.Discovery, h p2p_host.Host) (ipfs_interface.PubSubAPI, error) {
	logger.Info("creating pubsub")
	ps, err := p2p_pubsub.NewGossipSub(ctx, h,
		p2p_pubsub.WithMessageSigning(true),
		p2p_pubsub.WithFloodPublish(true),
		p2p_pubsub.WithDiscovery(disc),
	)
	logger.Info("done")

	if err != nil {
		return nil, err
	}

	return &PubSubAPI{
		host:   h,
		disc:   disc,
		PubSub: ps,
		logger: logger,
		topics: make(map[string]*p2p_pubsub.Topic),
	}, nil
}

func (ps *PubSubAPI) getTopic(topic string) (*p2p_pubsub.Topic, error) {
	ps.muTopics.Lock()
	defer ps.muTopics.Unlock()

	t, ok := ps.topics[topic]
	if !ok {
		var err error
		t, err = ps.PubSub.Join(topic)
		if err != nil {
			return nil, err
		}

		go func() {
			rkey := fmt.Sprintf("topic: `%s`", topic)
			for {
				time.Sleep(time.Second * 10)
				peers := t.ListPeers()
				fds := make([]zap.Field, len(peers))
				for i, peer := range peers {
					key := fmt.Sprintf("#%d", i)
					fds[i] = zap.String(key, peer.String())
				}

				ps.logger.Debug(rkey, fds...)
			}
		}()

		ps.topics[topic] = t
	}

	return t, nil
}

// Ls lists subscribed topics by name
func (ps *PubSubAPI) Ls(ctx context.Context) ([]string, error) {
	return ps.PubSub.GetTopics(), nil
}

// Peers list peers we are currently pubsubbing with
func (ps *PubSubAPI) Peers(ctx context.Context, opts ...ipfs_iopts.PubSubPeersOption) ([]p2p_peer.ID, error) {
	s, err := ipfs_iopts.PubSubPeersOptions(opts...)
	if err != nil {
		return nil, err
	}

	return ps.PubSub.ListPeers(s.Topic), nil
}

func (ps *PubSubAPI) Publish(ctx context.Context, topic string, msg []byte) error {
	// return ps.publish(ctx, getShaOnehash(topic), msg)
	return ps.publish(ctx, topic, msg)
}

// Publish a message to a given pubsub topic
func (ps *PubSubAPI) publish(ctx context.Context, topic string, msg []byte) error {
	const limit = 10

	ps.logger.Debug("HELLO IM: " + ps.host.ID().String())

	ps.logger.Debug("SUBSCRIBING", zap.String("topic", topic))
	if _, err := ps.subscribe(ctx, topic); err != nil {
		ps.logger.Warn("publish subscribe error", zap.Error(err))
	}

	ps.logger.Debug("PUBLISHING", zap.String("topic", topic), zap.Int("msglen", len(msg)))

	ps.logger.Debug("PUBLISHING: FIND PEERS")
	pc, err := ps.disc.FindPeers(ctx, topic, p2p_disc.Limit(limit))
	if err != nil {
		ps.logger.Warn("unable to findpeer to pubsub peer", zap.Error(err))
		return err
	}

	ready := false
	sc := sync.NewCond(&sync.Mutex{})
	sc.L.Lock()
	defer sc.L.Unlock()

	ps.logger.Debug("PUBLISHING: LOOP FOR PEERS")
	climit := 0
	for pi := range pc {
		if pi.ID == ps.host.ID() {
			continue
		}

		ps.host.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.TempAddrTTL)
		ps.logger.Debug("PUBLISHING: FOUND PEER:", zap.String("peer", pi.ID.String()))
		go func(pi p2p_peer.AddrInfo) {
			defer sc.Broadcast()

			ps.logger.Debug("PUBLISHING: CONNECTING TO PEER")
			context.WithTimeout(context.Background(), time.Second*10)
			if err := ps.host.Connect(context.Background(), pi); err != nil {
				ps.logger.Warn("unable to connect to pubsub peer", zap.Error(err))
				return
			}

			ps.logger.Debug("PUBLISHING: CONNECTING SUCCEED")
			sc.L.Lock()
			ready = true
			sc.L.Unlock()
			return
		}(pi)
		climit++
	}

	if climit == 0 {
		ps.logger.Debug("PUBLISHING: NO PEERS FOUND :(")
		return fmt.Errorf("Unable to publish: not enough peer on this topic")
	}

	ps.logger.Debug("PUBLISHING: WAIT TO BE READY")
	for i := 0; !ready && i < climit; i++ {
		sc.Wait()
	}

	t, err := ps.getTopic(topic)
	if err != nil {
		ps.logger.Warn("unable to get topic", zap.Error(err))
		return err
	}

	peers := t.ListPeers()
	if len(peers) == 0 {
		ps.logger.Warn("PUBLISHING: no peers connected to  this topic :(", zap.Error(err))
	}

	ps.logger.Debug("PUBLISHING: NOW", zap.String("topic", topic))
	return t.Publish(context.Background(), msg)
}

func (ps *PubSubAPI) Subscribe(ctx context.Context, topic string, opts ...ipfs_iopts.PubSubSubscribeOption) (ipfs_interface.PubSubSubscription, error) {
	// return ps.subscribe(ctx, getShaOnehash(topic), opts...)
	return ps.subscribe(ctx, topic, opts...)

}

// Subscribe to messages on a given topic
func (ps *PubSubAPI) subscribe(ctx context.Context, topic string, opts ...ipfs_iopts.PubSubSubscribeOption) (ipfs_interface.PubSubSubscription, error) {
	// s, err := ipfs_iopts.PubSubSubscribeOptions(opts...)
	// if err != nil {
	// 	return nil, err
	// }

	ps.logger.Debug("SUBSCRIBE GET:", zap.String("topic", topic))
	t, err := ps.getTopic(topic)
	if err != nil {
		ps.logger.Warn("unable to get topic", zap.Error(err))
		return nil, err
	}

	ps.logger.Debug("SUBSCRIBE:", zap.String("topic", topic))
	sub, err := t.Subscribe()
	if err != nil {
		ps.logger.Warn("unable to advertise pubsub peer", zap.Error(err))
		return nil, err
	}

	ps.logger.Debug("SUBSCRIBE ADVERTISE TOPIC:", zap.String("topic", topic))
	if _, err := ps.disc.Advertise(ctx, topic); err != nil {
		ps.logger.Warn("unable to advertise pubsub peer", zap.Error(err))
		return nil, err
	}

	ps.logger.Debug("SUBSCRIBE FIND PEER:", zap.String("topic", topic))
	pc, err := ps.disc.FindPeers(ctx, topic, p2p_disc.Limit(10))
	if err == nil {
		for pi := range pc {
			ps.logger.Debug("SUBSCRIBE FOUND PEER:", zap.String("peer", pi.ID.String()))
			if pi.ID == ps.host.ID() {
				continue
			}

			go func(pi p2p_peer.AddrInfo) {
				if err := ps.host.Connect(context.Background(), pi); err != nil {
					ps.logger.Warn("unable to connect to pubsub peer", zap.Error(err))
				}
			}(pi)

		}
	}

	return &pubsubSubscriptionAPI{ps.logger, sub}, nil
}

// PubSubSubscription is an active PubSub subscription
type pubsubSubscriptionAPI struct {
	logger *zap.Logger
	*p2p_pubsub.Subscription
}

// io.Closer
func (pss *pubsubSubscriptionAPI) Close() (_ error) {
	pss.Subscription.Cancel()
	return
}

// Next return the next incoming message
func (pss *pubsubSubscriptionAPI) Next(ctx context.Context) (ipfs_interface.PubSubMessage, error) {
	m, err := pss.Subscription.Next(ctx)
	if err != nil {
		pss.logger.Warn("UNABLE TO GET NEXT MESSAGE", zap.Error(err))
		return nil, err
	}

	pss.logger.Debug("GOT A MESSAGE",
		zap.Int("size", len(m.Message.Data)),
		zap.String("from", m.ReceivedFrom.String()),
		zap.Strings("topic", m.TopicIDs),
	)

	return &pubsubMessageAPI{m}, nil
}

// // PubSubMessage is a single PubSub message
type pubsubMessageAPI struct {
	*p2p_pubsub.Message
}

// // From returns id of a peer from which the message has arrived
func (psm *pubsubMessageAPI) From() p2p_peer.ID {
	return psm.Message.GetFrom()
}

// Data returns the message body
func (psm *pubsubMessageAPI) Data() []byte {
	return psm.Message.GetData()
}

// Seq returns message identifier
func (psm *pubsubMessageAPI) Seq() []byte {
	return psm.Message.GetSeqno()
}

// // Topics returns list of topics this message was set to
func (psm *pubsubMessageAPI) Topics() []string {
	return psm.Message.GetTopicIDs()
}

// func findPeerAndConnect(ctx context.Context, disc tinder.Driver, topic string, min int, max int) error {
// 	wg := sync.WaitGroup{}
// 	wg.Add(min)

// 	cp, err := disc.FindPeers(ctx, topic, p2p_discimpl.Limit(max))
// 	if err != nil {
// 		return err
// 	}

// }

// func (ps *PubSubAPI) MinTopicSize(size int) p2p_pubsub.RouterReady {
// 	return func(rt p2p_pubsub.PubSubRouter, topic string) (bool, error) {
// 		ps.logger.Debug("min peer started",
// 			zap.String("topic", topic))

// 		lp := ps.PubSub.ListPeers(topic)
// 		csize := len(lp)
// 		if csize < size {
// 			ps.logger.Warn("not enough peer",
// 				zap.String("topic", topic),
// 				zap.Int("need", size),
// 				zap.Int("have", csize))
// 			return false, nil
// 		}

// 		ps.logger.Debug("enough peer",
// 			zap.String("topic", topic),
// 			zap.Int("have", csize))

// 		return true, nil
// 	}
// }

// func peerEventLogger(logger *zap.Logger, topic string) p2p_pubsub.TopicEventHandlerOpt {
// 	return func(e *p2p_pubsub.TopicEventHandler) error {
// 		pe, err := e.NextPeerEvent(context.Background())
// 		if err != nil {
// 			return err
// 		}

// 		switch pe.Type {
// 		case p2p_pubsub.PeerJoin:
// 			logger.Debug("peer join",
// 				zap.String("topic", topic),
// 				zap.String("peer", pe.Peer.String()))
// 		case p2p_pubsub.PeerLeave:
// 			logger.Debug("peer leave",
// 				zap.String("topic", topic),
// 				zap.String("peer", pe.Peer.String()))
// 		}

// 		return nil
// 	}

// }

func getShaOnehash(b string) string {
	h := sha1.New()
	if _, err := h.Write([]byte(b)); err != nil {
		panic(err)
	}

	bs := h.Sum(nil)
	return fmt.Sprintf("%.10x", bs)
}
