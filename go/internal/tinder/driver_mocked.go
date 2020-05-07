// from https://github.com/libp2p/go-libp2p-discovery/blob/master/mocks_test.go

package tinder

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	p2p_discovery "github.com/libp2p/go-libp2p-core/discovery"
	p2p_host "github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	p2p_peer "github.com/libp2p/go-libp2p-core/peer"

	p2p_routing "github.com/libp2p/go-libp2p-core/routing"
)

type MockDriverServer struct {
	muPeers sync.RWMutex
	peers   map[p2p_peer.ID]p2p_peer.AddrInfo

	stores  map[string][]byte
	muStore sync.RWMutex // for value store

	muDB sync.RWMutex
	db   map[string]map[p2p_peer.ID]*discoveryRegistration
}

type discoveryRegistration struct {
	info       p2p_peer.AddrInfo
	expiration time.Time
}

func NewMockedDriverServer() *MockDriverServer {
	return &MockDriverServer{
		peers:  make(map[p2p_peer.ID]p2p_peer.AddrInfo),
		db:     make(map[string]map[p2p_peer.ID]*discoveryRegistration),
		stores: make(map[string][]byte),
	}
}

func (s *MockDriverServer) Advertise(ns string, info p2p_peer.AddrInfo, ttl time.Duration) (time.Duration, error) {
	s.muDB.Lock()
	defer s.muDB.Unlock()

	peers, ok := s.db[ns]
	if !ok {
		peers = make(map[p2p_peer.ID]*discoveryRegistration)
		s.db[ns] = peers
	}
	peers[info.ID] = &discoveryRegistration{info, time.Now().Add(ttl)}
	return ttl, nil
}

func (s *MockDriverServer) FindPeers(ns string, limit int) (<-chan p2p_peer.AddrInfo, error) {
	s.muDB.Lock()
	defer s.muDB.Unlock()

	peers, ok := s.db[ns]
	if !ok || len(peers) == 0 {
		emptyCh := make(chan p2p_peer.AddrInfo)
		close(emptyCh)
		return emptyCh, nil
	}

	count := len(peers)
	if limit != 0 && count > limit {
		count = limit
	}

	iterTime := time.Now()
	ch := make(chan p2p_peer.AddrInfo, count)
	numSent := 0
	for p, reg := range peers {
		if numSent == count {
			break
		}
		if iterTime.After(reg.expiration) {
			delete(peers, p)
			continue
		}

		numSent++
		ch <- reg.info
	}

	if len(peers) == 0 {
		delete(s.db, ns)
	}

	close(ch)

	return ch, nil
}

func (s *MockDriverServer) RegisterPeer(p p2p_peer.AddrInfo) {
	s.muPeers.Lock()
	s.peers[p.ID] = p
	s.muPeers.Unlock()
}

func (s *MockDriverServer) FindPeer(pid p2p_peer.ID) (p p2p_peer.AddrInfo, ok bool) {
	s.muPeers.RLock()
	p, ok = s.peers[pid]
	s.muPeers.RUnlock()
	return
}

func (s *MockDriverServer) PutValue(key string, value []byte) error {
	s.muStore.Lock()
	s.stores[key] = value
	s.muStore.Unlock()
	return nil
}

func (s *MockDriverServer) GetValue(key string) ([]byte, error) {
	if strings.HasPrefix(key, "/error/") {
		return nil, errors.New(key[len("/error/"):])
	}

	s.muStore.RLock()
	defer s.muStore.RUnlock()

	if ret, ok := s.stores[key]; ok {
		return ret, nil
	}

	return nil, p2p_routing.ErrNotFound
}

func (s *MockDriverServer) Unregister(ns string, pid p2p_peer.ID) {
	s.muDB.Lock()
	if peers, ok := s.db[ns]; ok {
		delete(peers, pid)

		if len(peers) == 0 {
			delete(s.db, ns)
		}
	}
	s.muDB.Unlock()
}

func (s *MockDriverServer) HasPeerRecord(ns string, pid p2p_peer.ID) bool {
	s.muDB.RLock()
	defer s.muDB.RUnlock()

	if peers, ok := s.db[ns]; ok {
		_, ok := peers[pid]
		return ok
	}
	return false
}

func (s *MockDriverServer) Reset() {
	s.muDB.Lock()
	s.db = make(map[string]map[p2p_peer.ID]*discoveryRegistration)
	s.muDB.Unlock()
}

var _ Driver = (*mockDriverClient)(nil)

type mockDriverClient struct {
	host   p2p_host.Host
	server *MockDriverServer
}

func NewMockedDriverClient(host p2p_host.Host, server *MockDriverServer) Routing {
	server.RegisterPeer(host.Peerstore().PeerInfo(host.ID()))
	return &mockDriverClient{host, server}
}

var _ p2p_discovery.Discovery = (*mockDriverClient)(nil)

func (d *mockDriverClient) Advertise(ctx context.Context, ns string, opts ...p2p_discovery.Option) (time.Duration, error) {
	var options p2p_discovery.Options
	err := options.Apply(opts...)
	if err != nil {
		return 0, err
	}

	return d.server.Advertise(ns, *p2p_host.InfoFromHost(d.host), options.Ttl)
}

func (d *mockDriverClient) FindPeers(ctx context.Context, ns string, opts ...p2p_discovery.Option) (<-chan p2p_peer.AddrInfo, error) {
	var options p2p_discovery.Options
	err := options.Apply(opts...)
	if err != nil {
		return nil, err
	}

	return d.server.FindPeers(ns, options.Limit)
}

func (d *mockDriverClient) Unregister(ctx context.Context, ns string) error {
	d.server.Unregister(ns, d.host.ID())
	return nil
}

var _ p2p_routing.PeerRouting = (*mockDriverClient)(nil)

func (d *mockDriverClient) FindPeer(_ context.Context, p p2p_peer.ID) (p2p_peer.AddrInfo, error) {
	if pi, ok := d.server.FindPeer(p); ok {
		return pi, nil
	}

	return p2p_peer.AddrInfo{}, fmt.Errorf("unable to find peer")
}

var _ p2p_routing.PeerRouting = (*mockDriverClient)(nil)

func (s *mockDriverClient) PutValue(ctx context.Context, key string, value []byte, opts ...p2p_routing.Option) error {
	return s.server.PutValue(key, value)
}

func (s *mockDriverClient) GetValue(ctx context.Context, key string, opts ...p2p_routing.Option) ([]byte, error) {
	return s.server.GetValue(key)
}

func (s *mockDriverClient) SearchValue(ctx context.Context, key string, opts ...p2p_routing.Option) (<-chan []byte, error) {
	out := make(chan []byte)
	go func() {
		defer close(out)
		v, err := s.server.GetValue(key)
		if err == nil {
			select {
			case out <- v:
			case <-ctx.Done():
			}
		}
	}()
	return out, nil
}

func (s *mockDriverClient) FindProvidersAsync(ctx context.Context, c cid.Cid, count int) <-chan peer.AddrInfo {
	caddrs, err := s.FindPeers(ctx, c.KeyString(), p2p_discovery.Limit(count))
	if err != nil {
		return nil
	}

	return caddrs

}

func (s *mockDriverClient) Provide(ctx context.Context, c cid.Cid, local bool) (err error) {
	_, err = s.Advertise(ctx, c.KeyString())
	return
}

func (s *mockDriverClient) Bootstrap(context.Context) error {
	return nil
}

func (d *mockDriverClient) Name() string { return "mock" }
