package bertyprotocol

import (
	"context"
	"sync"

	"fmt"

	"berty.tech/berty/v2/go/internal/cryptoutil"
	"berty.tech/berty/v2/go/pkg/bertytypes"
	"berty.tech/berty/v2/go/pkg/errcode"
	cid "github.com/ipfs/go-cid"
	datastore "github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	"github.com/libp2p/go-libp2p-core/crypto"
	"golang.org/x/crypto/nacl/secretbox"
)

type MessageKeystore struct {
	lock                 sync.Mutex
	preComputedKeysCount int
	store                datastore.Datastore
}

type decryptInfo struct {
	NewlyDecrypted bool
	MK             *[32]byte
	Cid            cid.Cid
}

func (m *MessageKeystore) getDeviceChainKey(ctx context.Context, pk crypto.PubKey) (*bertytypes.DeviceSecret, error) {
	pkB, err := pk.Raw()
	if err != nil {
		return nil, errcode.ErrSerialization.Wrap(err)
	}

	key := idForCurrentCK(pkB)

	dsBytes, err := m.store.Get(key)
	if err == datastore.ErrNotFound {
		return nil, errcode.ErrMissingInput
	} else if err != nil {
		return nil, errcode.ErrMessageKeyPersistenceGet.Wrap(err)
	}

	ds := &bertytypes.DeviceSecret{}
	if err := ds.Unmarshal(dsBytes); err != nil {
		return nil, errcode.ErrInvalidInput
	}

	return ds, nil
}

func (m *MessageKeystore) delPrecomputedKey(ctx context.Context, device crypto.PubKey, counter uint64) error {
	deviceRaw, err := device.Raw()
	if err != nil {
		return errcode.ErrSerialization.Wrap(err)
	}

	id := idForCachedKey(deviceRaw, counter)
	if err := m.store.Delete(id); err != nil {
		return errcode.ErrMessageKeyPersistencePut.Wrap(err)
	}

	return nil
}

func (m *MessageKeystore) postDecryptActions(ctx context.Context, di *decryptInfo, g *bertytypes.Group, ownPK crypto.PubKey, headers *bertytypes.MessageHeaders) error {
	// Message was newly decrypted, we can save the message key and derive
	// future keys if necessary.
	if di == nil || !di.NewlyDecrypted {
		return nil
	}

	var (
		ds  *bertytypes.DeviceSecret
		err error
	)

	pk, err := crypto.UnmarshalEd25519PublicKey(headers.DevicePK)
	if err != nil {
		return errcode.ErrDeserialization.Wrap(err)
	}

	if err = m.putKeyForCID(ctx, di.Cid, di.MK); err != nil {
		return errcode.ErrInternal.Wrap(err)
	}

	if err = m.delPrecomputedKey(ctx, pk, headers.Counter); err != nil {
		return errcode.ErrInternal.Wrap(err)
	}

	if ds, err = m.getDeviceChainKey(ctx, pk); err != nil {
		return errcode.ErrInvalidInput.Wrap(err)
	}

	if ds, err = m.preComputeKeys(ctx, pk, g, ds); err != nil {
		return errcode.ErrInternal.Wrap(err)
	}

	// If the message was not emitted by the current device we might need
	// to update the current chain key
	if ownPK == nil || !ownPK.Equals(pk) {
		if err = m.updateCurrentKey(ctx, pk, ds); err != nil {
			return errcode.ErrInternal.Wrap(err)
		}
	}

	return nil
}

func (m *MessageKeystore) GetDeviceSecret(ctx context.Context, g *bertytypes.Group, acc DeviceKeystore) (*bertytypes.DeviceSecret, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	md, err := acc.MemberDeviceForGroup(g)
	if err != nil {
		return nil, errcode.ErrInternal.Wrap(err)
	}

	ds, err := m.getDeviceChainKey(ctx, md.device.GetPublic())

	if err != nil && errcode.Code(err) == errcode.ErrMissingInput.Code() {
		// If secret does not exist, create it
		ds, err := newDeviceSecret()
		if err != nil {
			return nil, errcode.ErrCryptoKeyGeneration.Wrap(err)
		}

		if err = m.registerChainKey(ctx, g, md.device.GetPublic(), ds, true); err != nil {
			return nil, errcode.ErrMessageKeyPersistencePut.Wrap(err)
		}

		return ds, nil
	} else if err != nil {
		return nil, errcode.ErrMessageKeyPersistenceGet.Wrap(err)
	}

	return ds, nil
}

func (m *MessageKeystore) RegisterChainKey(ctx context.Context, g *bertytypes.Group, devicePK crypto.PubKey, ds *bertytypes.DeviceSecret, isOwnPK bool) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.registerChainKey(ctx, g, devicePK, ds, isOwnPK)
}

func (m *MessageKeystore) registerChainKey(ctx context.Context, g *bertytypes.Group, devicePK crypto.PubKey, ds *bertytypes.DeviceSecret, isOwnPK bool) error {
	var err error

	if _, err := m.getDeviceChainKey(ctx, devicePK); err == nil {
		// device is already registered, ignore it
		return nil
	}

	// If own device store key as is, no need to precompute future keys
	if isOwnPK {
		if err := m.putDeviceChainKey(ctx, devicePK, ds); err != nil {
			return errcode.ErrInternal.Wrap(err)
		}

		return nil
	}

	if ds, err = m.preComputeKeys(ctx, devicePK, g, ds); err != nil {
		return errcode.ErrCryptoKeyGeneration.Wrap(err)
	}

	if err := m.putDeviceChainKey(ctx, devicePK, ds); err != nil {
		return errcode.ErrInternal.Wrap(err)
	}

	return nil
}

func (m *MessageKeystore) preComputeKeys(ctx context.Context, device crypto.PubKey, g *bertytypes.Group, ds *bertytypes.DeviceSecret) (*bertytypes.DeviceSecret, error) {
	ck := ds.ChainKey
	counter := ds.Counter

	knownCK, err := m.getDeviceChainKey(ctx, device)
	if err != nil && errcode.Code(err) != errcode.ErrMissingInput.Code() {
		return nil, errcode.ErrInternal.Wrap(err)
	}

	for i := 1; i <= m.getPrecomputedKeyExpectedCount(); i++ {
		counter++

		knownMK, err := m.getPrecomputedKey(ctx, device, counter)
		if err != nil && errcode.Code(err) != errcode.ErrMissingInput.Code() {
			return nil, errcode.ErrInternal.Wrap(err)
		}

		if knownMK != nil && knownCK != nil {
			if knownCK.Counter != counter-1 {
				continue
			}
		}

		// TODO: Salt?
		newCK, mk, err := deriveNextKeys(ck, nil, g.GetPublicKey())
		if err != nil {
			return nil, errcode.TODO.Wrap(err)
		}

		err = m.putPrecomputedKey(ctx, device, counter, &mk)
		if err != nil {
			return nil, errcode.TODO.Wrap(err)
		}

		ck = newCK
	}

	return &bertytypes.DeviceSecret{
		Counter:  counter,
		ChainKey: ck,
	}, nil
}

func (m *MessageKeystore) getPrecomputedKey(ctx context.Context, device crypto.PubKey, counter uint64) (*[32]byte, error) {
	deviceRaw, err := device.Raw()
	if err != nil {
		return nil, errcode.ErrSerialization.Wrap(err)
	}

	id := idForCachedKey(deviceRaw, counter)

	key, err := m.store.Get(id)

	if err == datastore.ErrNotFound {
		return nil, errcode.ErrMissingInput.Wrap(fmt.Errorf("key for message does not exist in datastore"))
	} else if err != nil {
		return nil, errcode.ErrMessageKeyPersistenceGet.Wrap(err)
	}

	keyArray, err := cryptoutil.KeySliceToArray(key)
	if err != nil {
		return nil, errcode.ErrSerialization
	}

	return keyArray, nil
}

func (m *MessageKeystore) putPrecomputedKey(ctx context.Context, device crypto.PubKey, counter uint64, mk *[32]byte) error {
	deviceRaw, err := device.Raw()
	if err != nil {
		return errcode.ErrSerialization.Wrap(err)
	}

	id := idForCachedKey(deviceRaw, counter)

	if err := m.store.Put(id, mk[:]); err != nil {
		return errcode.ErrMessageKeyPersistencePut.Wrap(err)
	}

	return nil
}

func (m *MessageKeystore) putKeyForCID(ctx context.Context, id cid.Cid, key *[32]byte) error {
	if !id.Defined() {
		return nil
	}

	err := m.store.Put(idForCID(id), key[:])
	if err != nil {
		return errcode.ErrMessageKeyPersistencePut.Wrap(err)
	}

	return nil
}

func (m *MessageKeystore) OpenEnvelope(ctx context.Context, g *bertytypes.Group, ownPK crypto.PubKey, data []byte, id cid.Cid) (*bertytypes.MessageHeaders, []byte, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m == nil || g == nil {
		return nil, nil, errcode.ErrInvalidInput
	}

	env, headers, err := openEnvelopeHeaders(data, g)
	if err != nil {
		return nil, nil, errcode.ErrCryptoDecrypt.Wrap(err)
	}

	msg, decryptInfo, err := m.openPayload(ctx, id, env.Message, headers)
	if err != nil {
		return nil, nil, errcode.ErrCryptoDecrypt.Wrap(err)
	}

	if err := m.postDecryptActions(ctx, decryptInfo, g, ownPK, headers); err != nil {
		return nil, nil, errcode.TODO.Wrap(err)
	}

	return headers, msg, nil
}

func (m *MessageKeystore) openPayload(ctx context.Context, id cid.Cid, payload []byte, headers *bertytypes.MessageHeaders) ([]byte, *decryptInfo, error) {
	var (
		err error
		di  = &decryptInfo{
			Cid:            id,
			NewlyDecrypted: true,
		}
	)

	if di.MK, err = m.getKeyForCID(ctx, id); err == nil {
		di.NewlyDecrypted = false
	} else {
		pk, err := crypto.UnmarshalEd25519PublicKey(headers.DevicePK)
		if err != nil {
			return nil, nil, errcode.ErrDeserialization.Wrap(err)
		}

		di.MK, err = m.getPrecomputedKey(ctx, pk, headers.Counter)
		if err != nil {
			return nil, nil, errcode.ErrCryptoDecrypt.Wrap(err)
		}
	}

	msg, ok := secretbox.Open(nil, payload, uint64AsNonce(headers.Counter), di.MK)
	if !ok {
		return nil, nil, errcode.ErrCryptoDecrypt
	}

	// Message was newly decrypted, we can save the message key and derive
	// future keys if necessary.
	return msg, di, nil
}

func (m *MessageKeystore) getKeyForCID(ctx context.Context, id cid.Cid) (*[32]byte, error) {
	if !id.Defined() {
		return nil, errcode.ErrInvalidInput
	}

	key, err := m.store.Get(idForCID(id))
	if err == datastore.ErrNotFound {
		return nil, errcode.ErrInvalidInput
	}

	keyArray, err := cryptoutil.KeySliceToArray(key)
	if err != nil {
		return nil, errcode.ErrSerialization
	}

	return keyArray, nil
}

func (m *MessageKeystore) getPrecomputedKeyExpectedCount() int {
	return m.preComputedKeysCount
}

func (m *MessageKeystore) putDeviceChainKey(ctx context.Context, device crypto.PubKey, ds *bertytypes.DeviceSecret) error {
	deviceRaw, err := device.Raw()
	if err != nil {
		return errcode.ErrSerialization.Wrap(err)
	}

	key := idForCurrentCK(deviceRaw)

	data, err := ds.Marshal()
	if err != nil {
		return errcode.ErrSerialization.Wrap(err)
	}

	err = m.store.Put(key, data)
	if err != nil {
		return errcode.ErrMessageKeyPersistencePut.Wrap(err)
	}

	return nil
}

func (m *MessageKeystore) SealEnvelope(ctx context.Context, g *bertytypes.Group, deviceSK crypto.PrivKey, payload []byte) ([]byte, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if deviceSK == nil || g == nil || m == nil {
		return nil, errcode.ErrInvalidInput
	}

	ds, err := m.getDeviceChainKey(ctx, deviceSK.GetPublic())
	if err != nil {
		return nil, errcode.ErrInternal.Wrap(err)
	}

	env, err := sealEnvelopeInternal(ctx, payload, ds, deviceSK, g)
	if err != nil {
		return nil, errcode.ErrCryptoEncrypt.Wrap(err)
	}

	if err := m.deriveDeviceSecret(ctx, g, deviceSK); err != nil {
		return nil, errcode.ErrCryptoKeyGeneration.Wrap(err)
	}

	return env, nil
}

func (m *MessageKeystore) deriveDeviceSecret(ctx context.Context, g *bertytypes.Group, deviceSK crypto.PrivKey) error {
	if m == nil || deviceSK == nil {
		return errcode.ErrInvalidInput
	}

	ds, err := m.getDeviceChainKey(ctx, deviceSK.GetPublic())
	if err != nil {
		return errcode.ErrInvalidInput.Wrap(err)
	}

	ck, mk, err := deriveNextKeys(ds.ChainKey, nil, g.GetPublicKey())
	if err != nil {
		return errcode.ErrCryptoKeyGeneration.Wrap(err)
	}

	if err = m.putDeviceChainKey(ctx, deviceSK.GetPublic(), &bertytypes.DeviceSecret{
		ChainKey: ck,
		Counter:  ds.Counter + 1,
	}); err != nil {
		return errcode.ErrCryptoKeyGeneration.Wrap(err)
	}

	if err = m.putPrecomputedKey(ctx, deviceSK.GetPublic(), ds.Counter+1, &mk); err != nil {
		return errcode.ErrMessageKeyPersistencePut.Wrap(err)
	}

	return nil
}

func (m *MessageKeystore) updateCurrentKey(ctx context.Context, pk crypto.PubKey, ds *bertytypes.DeviceSecret) error {
	currentCK, err := m.getDeviceChainKey(ctx, pk)
	if err != nil {
		return errcode.ErrInternal.Wrap(err)
	}

	if ds.Counter < currentCK.Counter {
		return nil
	}

	if err = m.putDeviceChainKey(ctx, pk, ds); err != nil {
		return errcode.ErrInternal.Wrap(err)
	}

	return nil
}

// NewMessageKeystore instantiate a new MessageKeystore
func NewMessageKeystore(s datastore.Datastore) *MessageKeystore {
	return &MessageKeystore{
		preComputedKeysCount: 100,
		store:                s,
	}
}

// NewInMemMessageKeystore instantiate a new MessageKeystore, useful for testing
func NewInMemMessageKeystore() *MessageKeystore {
	return NewMessageKeystore(dssync.MutexWrap(datastore.NewMapDatastore()))
}
