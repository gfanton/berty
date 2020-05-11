package bertyprotocol

import (
	"context"

	"fmt"

	"sync"

	"bytes"

	"berty.tech/berty/v2/go/internal/handshake"
	"berty.tech/berty/v2/go/internal/ipfsutil"
	"berty.tech/berty/v2/go/pkg/bertytypes"
	"berty.tech/berty/v2/go/pkg/errcode"
	ggio "github.com/gogo/protobuf/io"
	"github.com/libp2p/go-libp2p-core/crypto"
	p2phelpers "github.com/libp2p/go-libp2p-core/helpers"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"go.uber.org/zap"
)

type pendingRequest struct {
	cancel  context.CancelFunc
	contact *bertytypes.ShareableContact
	lock    sync.Mutex
	ch      chan peer.AddrInfo
	swiper  *swiper
}

func (p *pendingRequest) update(ctx context.Context, contact *bertytypes.ShareableContact) {
	p.lock.Lock()

	if p.contact != nil && bytes.Equal(p.contact.PublicRendezvousSeed, contact.PublicRendezvousSeed) {
		p.lock.Unlock()
		return
	}

	if p.cancel != nil {
		p.cancel()
	}

	p.contact = contact
	ctx, p.cancel = context.WithCancel(ctx)

	p.lock.Unlock()

	for addr := range p.swiper.watch(ctx, contact.PK, contact.PublicRendezvousSeed) {
		p.ch <- addr
	}
}

func (p *pendingRequest) Close() error {
	p.lock.Lock()
	if p.cancel != nil {
		p.cancel()
	}

	close(p.ch)
	p.lock.Unlock()

	return nil
}

func newPendingRequest(ctx context.Context, swiper *swiper, contact *bertytypes.ShareableContact) (*pendingRequest, chan peer.AddrInfo) {
	ch := make(chan peer.AddrInfo)
	p := &pendingRequest{
		ch:     ch,
		swiper: swiper,
	}

	go p.update(ctx, contact)

	return p, ch
}

type contactRequestsManager struct {
	enabled        bool
	seed           []byte
	announceCancel context.CancelFunc
	metadataStore  *metadataStore
	lock           sync.Mutex
	ipfs           ipfsutil.ExtendedCoreAPI
	accSK          crypto.PrivKey
	ctx            context.Context
	logger         *zap.Logger
	swiper         *swiper
	toAdd          map[string]*pendingRequest
}

func (c *contactRequestsManager) metadataRequestDisabled(_ *bertytypes.GroupMetadataEvent) error {
	c.enabled = false
	if c.announceCancel != nil {
		c.announceCancel()
		c.ipfs.RemoveStreamHandler(contactRequestV1)
	}

	c.announceCancel = nil

	return nil
}

func (c *contactRequestsManager) metadataRequestEnabled(_ *bertytypes.GroupMetadataEvent) error {
	if err := c.listenToRequests(c.ctx, c.seed); err != nil {
		return err
	}

	c.enabled = true
	return nil
}

func (c *contactRequestsManager) metadataRequestReset(evt *bertytypes.GroupMetadataEvent) error {
	e := &bertytypes.AccountContactRequestReferenceReset{}
	if err := e.Unmarshal(evt.Event); err != nil {
		return errcode.ErrDeserialization.Wrap(err)
	}

	c.seed = e.PublicRendezvousSeed

	pkBytes, err := c.accSK.GetPublic().Raw()
	if err != nil {
		return err
	}

	if !c.enabled {
		return nil
	}

	if c.announceCancel != nil {
		c.announceCancel()
	}

	var ctx context.Context

	ctx, c.announceCancel = context.WithCancel(c.ctx)
	c.swiper.announce(ctx, pkBytes, c.seed)

	return nil
}

func (c *contactRequestsManager) metadataRequestEnqueued(evt *bertytypes.GroupMetadataEvent) error {
	e := &bertytypes.AccountContactRequestEnqueued{}
	if err := e.Unmarshal(evt.Event); err != nil {
		return err
	}

	return c.enqueueRequest(&bertytypes.ShareableContact{
		PK:                   e.Contact.PK,
		PublicRendezvousSeed: e.Contact.PublicRendezvousSeed,
	})
}

func (c *contactRequestsManager) metadataRequestSent(evt *bertytypes.GroupMetadataEvent) error {
	e := &bertytypes.AccountContactRequestSent{}
	if err := e.Unmarshal(evt.Event); err != nil {
		return err
	}

	if request, ok := c.toAdd[string(e.ContactPK)]; ok {
		if err := request.Close(); err != nil {
			c.logger.Warn("error while closing request", zap.Error(err))
		}

		delete(c.toAdd, string(e.ContactPK))
	}

	return nil
}

func (c *contactRequestsManager) metadataRequestReceived(evt *bertytypes.GroupMetadataEvent) error {
	e := &bertytypes.AccountContactRequestReceived{}
	if err := e.Unmarshal(evt.Event); err != nil {
		return err
	}

	if request, ok := c.toAdd[string(e.ContactPK)]; ok {
		if err := request.Close(); err != nil {
			c.logger.Warn("error while closing request", zap.Error(err))
		}

		delete(c.toAdd, string(e.ContactPK))
	}

	return nil
}

func (c *contactRequestsManager) enqueueRequest(contact *bertytypes.ShareableContact) error {
	pk, err := crypto.UnmarshalEd25519PublicKey(contact.PK)
	if err != nil {
		return err
	}

	if pending, ok := c.toAdd[string(contact.PK)]; ok {
		go pending.update(c.ctx, contact)

	} else {
		var ch chan peer.AddrInfo
		c.toAdd[string(contact.PK)], ch = newPendingRequest(c.ctx, c.swiper, contact)

		go func(pk crypto.PubKey, ch chan peer.AddrInfo) {
			for addr := range ch {
				stream, err := c.ipfs.NewStream(context.TODO(), addr.ID, contactRequestV1)
				if err != nil {
					c.logger.Error("error while opening stream with other peer", zap.Error(err))
					continue
				}

				c.performSend(pk, stream)
			}
		}(pk, ch)
	}

	return nil
}

func (c *contactRequestsManager) metadataWatcher(ctx context.Context) {
	handlers := map[bertytypes.EventType]func(*bertytypes.GroupMetadataEvent) error{
		bertytypes.EventTypeAccountContactRequestDisabled:         c.metadataRequestDisabled,
		bertytypes.EventTypeAccountContactRequestEnabled:          c.metadataRequestEnabled,
		bertytypes.EventTypeAccountContactRequestReferenceReset:   c.metadataRequestReset,
		bertytypes.EventTypeAccountContactRequestOutgoingEnqueued: c.metadataRequestEnqueued,
		bertytypes.EventTypeAccountContactRequestOutgoingSent:     c.metadataRequestSent,
		bertytypes.EventTypeAccountContactRequestIncomingReceived: c.metadataRequestReceived,
	}

	c.lock.Lock()
	enabled, contact := c.metadataStore.GetIncomingContactRequestsStatus()
	c.enabled = enabled
	c.seed = contact.PublicRendezvousSeed

	for _, contact := range c.metadataStore.ListContactsByStatus(bertytypes.ContactStateToRequest) {
		if err := c.enqueueRequest(contact); err != nil {
			c.logger.Error("unable to enqueue contact request", zap.Error(err))
		}
	}
	c.lock.Unlock()

	go func() {
		for evt := range c.metadataStore.Subscribe(ctx) {
			e, ok := evt.(*bertytypes.GroupMetadataEvent)
			if !ok {
				continue
			}

			if _, ok := handlers[e.Metadata.EventType]; !ok {
				continue
			}

			c.lock.Lock()
			if err := handlers[e.Metadata.EventType](e); err != nil {
				c.lock.Unlock()
				c.logger.Error("error while handling metadata store event", zap.Error(err))
				continue
			}

			c.lock.Unlock()
		}
	}()
}

const contactRequestV1 = "berty_contact_req_v1"

func (c *contactRequestsManager) incomingHandler(stream network.Stream) {
	defer func() {
		if err := p2phelpers.FullClose(stream); err != nil {
			c.logger.Warn("error while closing stream with other peer", zap.Error(err))
		}
	}()

	otherPK, err := handshake.Response(stream, c.accSK)
	if err != nil {
		c.logger.Error("an error occurred during handshake", zap.Error(err))
		return
	}

	otherPKBytes, err := otherPK.Raw()
	if err != nil {
		c.logger.Error("an error occurred during serialization", zap.Error(err))
		return
	}

	reader := ggio.NewDelimitedReader(stream, 2048)
	contact := &bertytypes.ShareableContact{}

	if err = reader.ReadMsg(contact); err != nil {
		c.logger.Error("an error occurred while retrieving contact information", zap.Error(err))
		return
	}

	if err = contact.CheckFormat(bertytypes.ShareableContactOptionsAllowMissingRDVSeed, bertytypes.ShareableContactOptionsAllowMissingPK); err != nil {
		c.logger.Error("an error occurred while verifying contact information", zap.Error(err))
		return
	}

	if _, err = c.metadataStore.ContactRequestIncomingReceived(c.ctx, &bertytypes.ShareableContact{
		PK:                   otherPKBytes,
		PublicRendezvousSeed: contact.PublicRendezvousSeed,
		Metadata:             contact.Metadata,
	}); err != nil {
		c.logger.Error("an error occurred while adding contact request to received", zap.Error(err))
		return
	}
}

func (c *contactRequestsManager) performSend(otherPK crypto.PubKey, stream network.Stream) {
	defer func() {
		if err := p2phelpers.FullClose(stream); err != nil {
			c.logger.Warn("error while closing stream with other peer", zap.Error(err))
		}
	}()
	_, contact := c.metadataStore.GetIncomingContactRequestsStatus()

	if err := handshake.Request(stream, c.accSK, otherPK); err != nil {
		c.logger.Error("an error occurred during handshake", zap.Error(err))
		return
	}

	writer := ggio.NewDelimitedWriter(stream)

	if err := writer.WriteMsg(contact); err != nil {
		c.logger.Error("an error occurred while sending own contact information", zap.Error(err))
		return
	}

	if _, err := c.metadataStore.ContactRequestOutgoingSent(c.ctx, otherPK); err != nil {
		c.logger.Error("an error occurred while marking contact request as sent", zap.Error(err))
		return
	}
}

func (c *contactRequestsManager) listenToRequests(ctx context.Context, seed []byte) error {
	pkBytes, err := c.accSK.GetPublic().Raw()
	if err != nil {
		return errcode.ErrSerialization.Wrap(err)
	}

	if l := len(seed); l != 32 {
		return errcode.ErrInvalidInput.Wrap(fmt.Errorf("invalid rendezvous seed length %d", l))
	}

	if c.announceCancel != nil {
		c.announceCancel()
	}

	ctx, c.announceCancel = context.WithCancel(ctx)
	c.swiper.announce(ctx, pkBytes, c.seed)

	c.ipfs.SetStreamHandler(contactRequestV1, c.incomingHandler)

	return nil
}

func initContactRequestsManager(ctx context.Context, s *swiper, store *metadataStore, ipfs ipfsutil.ExtendedCoreAPI, logger *zap.Logger) error {
	sk, err := store.devKS.AccountPrivKey()
	if err != nil {
		return err
	}

	cm := &contactRequestsManager{
		metadataStore: store,
		ipfs:          ipfs,
		logger:        logger,
		accSK:         sk,
		ctx:           ctx,
		swiper:        s,
		toAdd:         map[string]*pendingRequest{},
	}

	go cm.metadataWatcher(ctx)

	return nil
}
