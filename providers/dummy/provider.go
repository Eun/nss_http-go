package dummy

import (
	"context"
	"encoding/json"

	types "github.com/Eun/nss_http/types"
)

const Name = "dummy"

var _ types.Provider = &Provider{}
var _ types.CacheProvider = &Provider{}

type Provider struct{}

func (f *Provider) Name() string {
	return Name
}

func (f *Provider) GetUser(ctx context.Context, identifier any) (*types.User, error) {
	return nil, nil
}

func (f *Provider) SetUser(ctx context.Context, user *types.User) error {
	return nil
}

func (f *Provider) GetGroup(ctx context.Context, identifier any) (*types.Group, error) {
	return nil, nil
}

func (f *Provider) SetGroup(ctx context.Context, user *types.Group) error {
	return nil
}

func New(json.RawMessage) (*Provider, error) {
	return &Provider{}, nil
}
