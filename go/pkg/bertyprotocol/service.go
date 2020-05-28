package bertyprotocol

import (
	"context"
	"sync"
	"time"

	"berty.tech/berty/v2/go/internal/ipfsutil"
	"berty.tech/berty/v2/go/internal/tinder"
	"berty.tech/berty/v2/go/internal/tracer"
	"berty.tech/berty/v2/go/pkg/bertytypes"
	"berty.tech/berty/v2/go/pkg/errcode"
	orbitdb "berty.tech/go-orbit-db"
	"berty.tech/go-orbit-db/cache"
	"github.com/ipfs/go-datastore"
	ds_sync "github.com/ipfs/go-datastore/sync"
	ipfs_core "github.com/ipfs/go-ipfs/core"
	"go.opentelemetry.io/otel/api/trace"
	"go.uber.org/zap"
)

var _ Service = (*service)(nil)

// Service is the main Berty Protocol interface
type Service interface {
	ProtocolServiceServer

	Close() error
	Status() Status
}

type service struct {
	// variables
	ctx            context.Context
	logger         *zap.Logger
	ipfsCoreAPI    ipfsutil.ExtendedCoreAPI
	odb            *bertyOrbitDB
	accountGroup   *groupContext
	deviceKeystore DeviceKeystore
	openedGroups   map[string]*groupContext
	groups         map[string]*bertytypes.Group
	lock           sync.RWMutex
	close          func() error
}

// Opts contains optional configuration flags for building a new Client
type Opts struct {
	Tracer                 trace.Tracer
	Logger                 *zap.Logger
	IpfsCoreAPI            ipfsutil.ExtendedCoreAPI
	DeviceKeystore         DeviceKeystore
	MessageKeystore        *MessageKeystore
	RootContext            context.Context
	RootDatastore          datastore.Batching
	OrbitDirectory         string
	OrbitCache             cache.Interface
	TinderDriver           tinder.Driver
	RendezvousRotationBase time.Duration
	close                  func() error
}

func defaultClientOptions(opts *Opts) error {
	if opts.Tracer == nil {
		opts.Tracer = tracer.New("bertyprotocol")
	}

	if opts.Logger == nil {
		opts.Logger = zap.NewNop()
	}

	if opts.OrbitDirectory == "" {
		opts.OrbitDirectory = ":memory:"
	}

	if opts.RootContext == nil {
		opts.RootContext = context.TODO()
	}

	if opts.RootDatastore == nil {
		opts.RootDatastore = ds_sync.MutexWrap(datastore.NewMapDatastore())
	}

	if opts.DeviceKeystore == nil {
		ks := ipfsutil.NewDatastoreKeystore(ipfsutil.NewNamespacedDatastore(opts.RootDatastore, datastore.NewKey("accountGroup")))
		opts.DeviceKeystore = NewDeviceKeystore(ks)
	}

	if opts.MessageKeystore == nil {
		mk := ipfsutil.NewNamespacedDatastore(opts.RootDatastore, datastore.NewKey("messages"))
		opts.MessageKeystore = NewMessageKeystore(mk)
	}

	if opts.RendezvousRotationBase.Nanoseconds() <= 0 {
		opts.RendezvousRotationBase = time.Hour * 24
	}

	if opts.IpfsCoreAPI == nil {
		var err error
		var createdIPFSNode *ipfs_core.IpfsNode

		opts.IpfsCoreAPI, createdIPFSNode, err = ipfsutil.NewCoreAPI(opts.RootContext, &ipfsutil.CoreAPIConfig{})
		if err != nil {
			return errcode.TODO.Wrap(err)
		}

		oldClose := opts.close
		opts.close = func() error {
			if oldClose != nil {
				_ = oldClose()
			}

			return createdIPFSNode.Close()
		}
	}

	return nil
}

// New initializes a new Service
func New(opts Opts) (Service, error) {
	if err := defaultClientOptions(&opts); err != nil {
		return nil, errcode.TODO.Wrap(err)
	}

	orbitDirectory := opts.OrbitDirectory
	odb, err := newBertyOrbitDB(opts.RootContext, opts.IpfsCoreAPI, opts.DeviceKeystore, opts.MessageKeystore, &orbitdb.NewOrbitDBOptions{
		Cache:     opts.OrbitCache,
		Directory: &orbitDirectory,
		Logger:    opts.Logger,
		Tracer:    opts.Tracer,
	})
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}

	acc, err := odb.OpenAccountGroup(opts.RootContext, nil)
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}

	if opts.TinderDriver != nil {
		s := newSwiper(opts.TinderDriver, opts.Logger, opts.RendezvousRotationBase)
		opts.Logger.Debug("tinder swiper is enabled")

		if err := initContactRequestsManager(opts.RootContext, s, acc.metadataStore, opts.IpfsCoreAPI, opts.Logger); err != nil {
			return nil, errcode.TODO.Wrap(err)
		}
	} else {
		opts.Logger.Warn("no tinder driver provided, incoming and outgoing contact requests won't be enabled")
	}

	return &service{
		ctx:            opts.RootContext,
		ipfsCoreAPI:    opts.IpfsCoreAPI,
		logger:         opts.Logger,
		odb:            odb,
		deviceKeystore: opts.DeviceKeystore,
		close:          opts.close,
		accountGroup:   acc,
		groups: map[string]*bertytypes.Group{
			string(acc.Group().PublicKey): acc.Group(),
		},
		openedGroups: map[string]*groupContext{
			string(acc.Group().PublicKey): acc,
		},
	}, nil
}

func (s *service) Close() error {
	s.odb.Close()
	if s.close != nil {
		s.close()
	}

	return nil
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
