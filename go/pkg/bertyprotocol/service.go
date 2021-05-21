package bertyprotocol

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	ipfs_mobile "github.com/ipfs-shipyard/gomobile-ipfs/go/pkg/ipfsmobile"
	ds "github.com/ipfs/go-datastore"
	ds_sync "github.com/ipfs/go-datastore/sync"
	ipfs_interface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/libp2p/go-libp2p-core/host"
	peer "github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"berty.tech/berty/v2/go/internal/ipfsutil"
	"berty.tech/berty/v2/go/internal/tinder"
	"berty.tech/berty/v2/go/pkg/bertyversion"
	"berty.tech/berty/v2/go/pkg/errcode"
	"berty.tech/berty/v2/go/pkg/protocoltypes"
	"berty.tech/berty/v2/go/pkg/tyber"
	"berty.tech/go-orbit-db/baseorbitdb"
	"berty.tech/go-orbit-db/iface"
	"berty.tech/go-orbit-db/pubsub/directchannel"
)

var _ Service = (*service)(nil)

// Service is the main Berty Protocol interface
type Service interface {
	protocoltypes.ProtocolServiceServer

	Close() error
	Status() Status
	IpfsCoreAPI() ipfs_interface.CoreAPI
}

type service struct {
	// variables
	ctx            context.Context
	logger         *zap.Logger
	ipfsCoreAPI    ipfsutil.ExtendedCoreAPI
	odb            *BertyOrbitDB
	accountGroup   *groupContext
	deviceKeystore DeviceKeystore
	openedGroups   map[string]*groupContext
	groups         map[string]*protocoltypes.Group
	lock           sync.RWMutex
	authSession    atomic.Value
	close          func() error
	startedAt      time.Time
	host           host.Host
}

// Opts contains optional configuration flags for building a new Client
type Opts struct {
	Logger                 *zap.Logger
	IpfsCoreAPI            ipfsutil.ExtendedCoreAPI
	DeviceKeystore         DeviceKeystore
	DatastoreDir           string
	RootDatastore          ds.Batching
	OrbitDB                *BertyOrbitDB
	TinderDriver           tinder.UnregisterDiscovery
	RendezvousRotationBase time.Duration
	Host                   host.Host
	PubSub                 *pubsub.PubSub
	LocalOnly              bool
	close                  func() error
}

func (opts *Opts) applyDefaults(ctx context.Context) error {
	if opts.Logger == nil {
		opts.Logger = zap.NewNop()
	}

	if opts.RootDatastore == nil {
		if opts.DatastoreDir == "" || opts.DatastoreDir == InMemoryDirectory {
			opts.RootDatastore = ds_sync.MutexWrap(ds.NewMapDatastore())
		} else {
			opts.RootDatastore = nil
		}
	}

	if opts.DeviceKeystore == nil {
		ks := ipfsutil.NewDatastoreKeystore(ipfsutil.NewNamespacedDatastore(opts.RootDatastore, ds.NewKey(NamespaceDeviceKeystore)))
		opts.DeviceKeystore = NewDeviceKeystore(ks)
	}

	if opts.RendezvousRotationBase.Nanoseconds() <= 0 {
		opts.RendezvousRotationBase = time.Hour * 24
	}

	if opts.IpfsCoreAPI == nil {
		dsync := ds_sync.MutexWrap(ds.NewMapDatastore())
		repo, err := ipfsutil.CreateMockedRepo(dsync)
		if err != nil {
			return err
		}

		mrepo := ipfs_mobile.NewRepoMobile("", repo)
		mnode, err := ipfsutil.NewIPFSMobile(ctx, mrepo, &ipfsutil.MobileOptions{
			ExtraOpts: map[string]bool{
				"pubsub": true,
			},
		})
		if err != nil {
			return err
		}

		opts.IpfsCoreAPI, err = ipfsutil.NewExtendedCoreAPIFromNode(mnode.IpfsNode)
		if err != nil {
			return err
		}
		opts.Host = mnode.PeerHost()

		oldClose := opts.close
		opts.close = func() error {
			if oldClose != nil {
				_ = oldClose()
			}

			return mnode.Close()
		}
	}

	if opts.OrbitDB == nil {
		orbitDirectory := InMemoryDirectory
		if opts.DatastoreDir != InMemoryDirectory {
			orbitDirectory = filepath.Join(opts.DatastoreDir, NamespaceOrbitDBDirectory)
		}
		odbOpts := &NewOrbitDBOptions{
			NewOrbitDBOptions: baseorbitdb.NewOrbitDBOptions{
				Directory: &orbitDirectory,
				Logger:    opts.Logger,
			},
			Datastore:      ipfsutil.NewNamespacedDatastore(opts.RootDatastore, ds.NewKey(NamespaceOrbitDBDatastore)),
			DeviceKeystore: opts.DeviceKeystore,
		}

		if opts.Host != nil {
			odbOpts.NewOrbitDBOptions.DirectChannelFactory = directchannel.InitDirectChannelFactory(opts.Host)
		}

		odb, err := NewBertyOrbitDB(ctx, opts.IpfsCoreAPI, odbOpts)
		if err != nil {
			return err
		}

		oldClose := opts.close
		opts.close = func() error {
			if oldClose != nil {
				_ = oldClose()
			}

			return odb.Close()
		}

		opts.OrbitDB = odb
	}

	return nil
}

// New initializes a new Service
func New(ctx context.Context, opts Opts) (_ Service, err error) {
	if err := opts.applyDefaults(ctx); err != nil {
		return nil, errcode.TODO.Wrap(err)
	}

	opts.Logger = opts.Logger.Named("pt")

	ctx, _, endSection := tyber.Section(tyber.ContextWithNewTraceID(ctx), opts.Logger, fmt.Sprintf("Initializing ProtocolService version %s", bertyversion.Version))
	defer func() { endSection(err, "") }()

	dbOpts := &iface.CreateDBOptions{LocalOnly: &opts.LocalOnly}

	acc, err := opts.OrbitDB.openAccountGroup(ctx, dbOpts, opts.IpfsCoreAPI)
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}

	opts.Logger.Debug("Opened account group", tyber.FormatStepLogFields(ctx, []tyber.Detail{{Name: "AccountGroup", Description: acc.group.String()}})...)

	if opts.TinderDriver != nil {
		s := NewSwiper(opts.Logger, opts.PubSub, opts.RendezvousRotationBase)
		opts.Logger.Debug("Tinder swiper is enabled", tyber.FormatStepLogFields(ctx, []tyber.Detail{})...)

		if err := initContactRequestsManager(ctx, s, acc.metadataStore, opts.IpfsCoreAPI, opts.Logger); err != nil {
			return nil, errcode.TODO.Wrap(err)
		}
	} else {
		opts.Logger.Warn("No tinder driver provided, incoming and outgoing contact requests won't be enabled", tyber.FormatStepLogFields(ctx, []tyber.Detail{})...)
	}

	s := &service{
		ctx:            ctx,
		host:           opts.Host,
		ipfsCoreAPI:    opts.IpfsCoreAPI,
		logger:         opts.Logger,
		odb:            opts.OrbitDB,
		deviceKeystore: opts.DeviceKeystore,
		close:          opts.close,
		accountGroup:   acc,
		startedAt:      time.Now(),
		groups: map[string]*protocoltypes.Group{
			string(acc.Group().PublicKey): acc.Group(),
		},
		openedGroups: map[string]*groupContext{
			string(acc.Group().PublicKey): acc,
		},
	}

	s.startTyberTinderMonitor()

	return s, nil
}

func (s *service) IpfsCoreAPI() ipfs_interface.CoreAPI {
	return s.ipfsCoreAPI
}

func (s *service) Close() error {
	endSection := tyber.FastSection(tyber.ContextWithNewTraceID(s.ctx), s.logger, "Closing ProtocolService")

	err := s.odb.Close()

	if s.close != nil {
		err = multierr.Append(err, s.close())
	}

	endSection(err)
	return err
}

func (s *service) startTyberTinderMonitor() {
	if s.host == nil {
		return
	}

	sub, err := s.host.EventBus().Subscribe([]interface{}{
		new(ipfsutil.EvtPubSubTopic),
		new(tinder.EvtDriverMonitor),
	})
	if err != nil {
		s.logger.Error("failed to create sub", zap.Error(errors.Wrap(err, "unable to subscribe pubsub topic event")))
		return
	}

	// @FIXME(gfanton): cached found peers should be done inside driver monitor
	cachedFoundPeers := make(map[peer.ID]ipfsutil.Multiaddrs)
	ch := sub.Out()
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case evt := <-ch:
				var monitorEvent *protocoltypes.MonitorGroup_EventMonitor

				switch e := evt.(type) {
				case ipfsutil.EvtPubSubTopic:
					// handle this event
					monitorEvent = monitorHandlePubsubEvent(&e, s.host)
				case tinder.EvtDriverMonitor:
					// check if we already know this peer in case of found peer
					newMS := ipfsutil.NewMultiaddrs(e.AddrInfo.Addrs)
					if ms, ok := cachedFoundPeers[e.AddrInfo.ID]; ok {
						if ipfsutil.MultiaddrIsEqual(ms, newMS) {
							continue
						}
					}

					cachedFoundPeers[e.AddrInfo.ID] = newMS
					monitorEvent = monitorHandleDiscoveryEvent(&e, s.host)
				default:
					monitorEvent = &protocoltypes.MonitorGroup_EventMonitor{
						Type: protocoltypes.TypeEventMonitorUndefined,
					}
				}

				switch monitorEvent.Type {
				case protocoltypes.TypeEventMonitorPeerFound:
					s.logger.Debug(tyber.WKENTinderPeerFound, tyber.FormatEventLogFields(s.ctx, []tyber.Detail{
						{Name: "Topic", Description: monitorEvent.PeerFound.Topic},
						{Name: "PeerID", Description: monitorEvent.PeerFound.PeerID},
						{Name: "DriverName", Description: monitorEvent.PeerFound.DriverName},
						{Name: "Maddrs", Description: fmt.Sprint(monitorEvent.PeerFound.Maddrs)},
					})...)
				case protocoltypes.TypeEventMonitorPeerJoin:
					s.logger.Debug(tyber.WKENTinderPeerJoined, tyber.FormatEventLogFields(s.ctx, []tyber.Detail{
						{Name: "Topic", Description: monitorEvent.PeerJoin.Topic},
						{Name: "PeerID", Description: monitorEvent.PeerJoin.PeerID},
						{Name: "IsSelf", Description: fmt.Sprint(monitorEvent.PeerJoin.IsSelf)},
						{Name: "Maddrs", Description: fmt.Sprint(monitorEvent.PeerJoin.Maddrs)},
					})...)
				case protocoltypes.TypeEventMonitorPeerLeave:
					s.logger.Debug(tyber.WKENTinderPeerLeft, tyber.FormatEventLogFields(s.ctx, []tyber.Detail{
						{Name: "Topic", Description: monitorEvent.PeerLeave.Topic},
						{Name: "PeerID", Description: monitorEvent.PeerLeave.PeerID},
						{Name: "IsSelf", Description: fmt.Sprint(monitorEvent.PeerLeave.IsSelf)},
					})...)
				}
			}
		}
	}()
}

// Status contains results of status checks
type Status struct {
	DB       error
	Protocol error
}

func (s *service) Status() Status {
	return Status{
		Protocol: nil,
	}
}
