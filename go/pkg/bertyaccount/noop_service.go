package bertyaccount

import (
	context "context"

	"berty.tech/berty/v2/go/pkg/errcode"
)

// `NoopService` is a `Service`
var _ Service = (*NoopService)(nil)

type NoopService struct{}

var noopAccountMeta = &AccountMetadata{
	AccountID: "0",
	Name:      "iPhone Noop",
	PublicKey: "PRLLCXHsVodcmRrAMzN2-Upeyym1MIGqeaWFgiQY7zA",
}

func NewNoopService() Service {
	return &NoopService{}
}

// OpenAccount starts a Berty node.
func (*NoopService) OpenAccount(context.Context, *OpenAccount_Request) (*OpenAccount_Reply, error) {
	return &OpenAccount_Reply{}, nil
}

// OpenAccountWithProgress is similar to OpenAccount, but also streams the progress.
func (*NoopService) OpenAccountWithProgress(*OpenAccountWithProgress_Request, AccountService_OpenAccountWithProgressServer) error {
	return errcode.ErrNotOperational
}

// CloseAccount closes the currently opened account.
func (*NoopService) CloseAccount(context.Context, *CloseAccount_Request) (*CloseAccount_Reply, error) {
	return &CloseAccount_Reply{}, nil
}

// CloseAccountWithProgress is similar to CloseAccount, but also streams the progress.
func (*NoopService) CloseAccountWithProgress(*CloseAccountWithProgress_Request, AccountService_CloseAccountWithProgressServer) error {
	return nil
}

// ListAccounts retrieves a list of local accounts.
func (*NoopService) ListAccounts(context.Context, *ListAccounts_Request) (*ListAccounts_Reply, error) {
	return &ListAccounts_Reply{
		Accounts: []*AccountMetadata{noopAccountMeta},
	}, nil
}

// DeleteAccount deletes an account.
func (*NoopService) DeleteAccount(context.Context, *DeleteAccount_Request) (*DeleteAccount_Reply, error) {
	return &DeleteAccount_Reply{}, nil
}

// ImportAccount imports existing data.
func (*NoopService) ImportAccount(context.Context, *ImportAccount_Request) (*ImportAccount_Reply, error) {
	return &ImportAccount_Reply{}, nil
}

// CreateAccount creates a new account.
func (*NoopService) CreateAccount(context.Context, *CreateAccount_Request) (*CreateAccount_Reply, error) {
	return &CreateAccount_Reply{
		AccountMetadata: noopAccountMeta,
	}, nil
}

// UpdateAccount update account's metadata.
func (*NoopService) UpdateAccount(context.Context, *UpdateAccount_Request) (*UpdateAccount_Reply, error) {
	return &UpdateAccount_Reply{}, nil
}

// GetGRPCListenerAddrs return current listeners addrs available on this bridge.
func (*NoopService) GetGRPCListenerAddrs(context.Context, *GetGRPCListenerAddrs_Request) (*GetGRPCListenerAddrs_Reply, error) {
	return &GetGRPCListenerAddrs_Reply{}, nil
}

// WakeUp should be used for background task or similar task
func (*NoopService) WakeUp(ctx context.Context) error {
	return nil
}

// Close the service
func (*NoopService) Close() error {
	return nil
}
