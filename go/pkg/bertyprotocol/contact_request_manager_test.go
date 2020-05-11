package bertyprotocol

import (
	"context"
	"testing"
	"time"

	"io"

	"berty.tech/berty/v2/go/internal/testutil"
	"berty.tech/berty/v2/go/pkg/bertytypes"
	libp2p_mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/stretchr/testify/require"
)

func TestContactRequestFlow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	opts := TestingOpts{
		Mocknet: libp2p_mocknet.New(ctx),
		Logger:  testutil.Logger(t),
	}

	pts, cleanup := generateTestingProtocol(ctx, t, &opts, 2)
	defer cleanup()

	t.Log("1")

	_, err := pts[0].Client.ContactRequestEnable(ctx, &bertytypes.ContactRequestEnable_Request{})
	require.NoError(t, err)

	_, err = pts[1].Client.ContactRequestEnable(ctx, &bertytypes.ContactRequestEnable_Request{})
	require.NoError(t, err)

	t.Log("2")

	config0, err := pts[0].Client.InstanceGetConfiguration(ctx, &bertytypes.InstanceGetConfiguration_Request{})
	require.NoError(t, err)
	require.NotNil(t, config0)

	config1, err := pts[0].Client.InstanceGetConfiguration(ctx, &bertytypes.InstanceGetConfiguration_Request{})
	require.NoError(t, err)
	require.NotNil(t, config1)

	t.Log("3")

	ref0, err := pts[0].Client.ContactRequestResetReference(ctx, &bertytypes.ContactRequestResetReference_Request{})
	require.NoError(t, err)
	require.NotNil(t, ref0)

	ref1, err := pts[1].Client.ContactRequestResetReference(ctx, &bertytypes.ContactRequestResetReference_Request{})
	require.NoError(t, err)
	require.NotNil(t, ref1)

	t.Log("4")

	subMeta0, err := pts[0].Client.GroupMetadataSubscribe(ctx, &bertytypes.GroupMetadataSubscribe_Request{
		GroupPK: config0.AccountGroupPK,
		Since:   []byte("give me everything"),
	})
	require.NoError(t, err)
	found := false

	t.Log("5")

	_, err = pts[1].Client.ContactRequestSend(ctx, &bertytypes.ContactRequestSend_Request{
		Contact: &bertytypes.ShareableContact{
			PK:                   config0.AccountPK,
			PublicRendezvousSeed: ref0.PublicRendezvousSeed,
		},
	})
	require.NoError(t, err)

	t.Log("6")

	for {
		evt, err := subMeta0.Recv()
		if err == io.EOF {
			break
		}

		require.NoError(t, err)

		if evt == nil || evt.Metadata.EventType != bertytypes.EventTypeAccountContactRequestIncomingReceived {
			continue
		}

		req := bertytypes.AccountContactRequestReceived{}

		require.NoError(t, err)
		require.Equal(t, req.ContactPK, config1.AccountPK)
		found = true
	}

	require.True(t, found)
}
