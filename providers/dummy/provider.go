package dummy

import (
	"context"
	"encoding/json"

	types "github.com/Eun/nss_http/types"
)

const Name = "dummy"

var _ types.CacheProvider = &Provider{}

type Provider struct{}

func (*Provider) Name() string {
	return Name
}

func (*Provider) GetUser(ctx context.Context, identifier any) (*types.User, error) {
	return nil, nil
}

func (*Provider) SetUser(ctx context.Context, user *types.User) error {
	return nil
}

func (*Provider) GetUsers(ctx context.Context) (types.Users, error) {
	return nil, nil
}

func (*Provider) SetUsers(ctx context.Context, users types.Users) error {
	return nil
}

func (*Provider) GetGroup(ctx context.Context, identifier any) (*types.Group, error) {
	return nil, nil
}

func (*Provider) SetGroup(ctx context.Context, group *types.Group) error {
	return nil
}

func (*Provider) GetGroups(ctx context.Context) (types.Groups, error) {
	return nil, nil
}

func (*Provider) SetGroups(ctx context.Context, groups types.Groups) error {
	return nil
}

func New(json.RawMessage) (*Provider, error) {
	return &Provider{}, nil
}
