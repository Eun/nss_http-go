package file

import (
	"context"
	"encoding/json"
	"os"

	"github.com/Eun/nss_http/types"
	"github.com/pkg/errors"
)

const Name = "file"

var _ types.Provider = &Provider{}

type Provider struct {
	config Config
}

func (p *Provider) Name() string {
	return Name
}

func (p *Provider) GetUser(ctx context.Context, identifier any) (*types.User, error) {
	items, err := p.GetUsers(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get users")
	}
	switch v := identifier.(type) {
	case types.UIDIdentifier:
		for i := range items {
			if items[i].Uid == uint(v) {
				return &items[i], nil
			}
		}
		return nil, nil
	case types.NameIdentifier:
		for i := range items {
			if items[i].User == string(v) {
				return &items[i], nil
			}
		}
		return nil, nil
	}
	return nil, errors.New("unknown identifier type")

}

func (p *Provider) GetUsers(ctx context.Context) (types.Users, error) {
	if p.config.Users == "" {
		return nil, nil
	}
	buf, err := os.ReadFile(p.config.Users)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read group file")
	}
	var items []types.User
	if err := json.Unmarshal(buf, &items); err != nil {
		return nil, errors.Wrapf(err, "unable to decode file: %q", string(buf))
	}
	return items, nil
}

func (p *Provider) GetGroup(ctx context.Context, identifier any) (*types.Group, error) {
	if p.config.Groups == "" {
		return nil, nil
	}
	items, err := p.GetGroups(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get groups")
	}
	switch v := identifier.(type) {
	case types.GIDIdentifier:
		for i := range items {
			if items[i].Gid == uint(v) {
				return &items[i], nil
			}
		}
		return nil, nil
	case types.NameIdentifier:
		for i := range items {
			if items[i].Name == string(v) {
				return &items[i], nil
			}
		}
		return nil, nil
	}
	return nil, errors.New("unknown identifier type")
}

func (p *Provider) GetGroups(ctx context.Context) (types.Groups, error) {
	buf, err := os.ReadFile(p.config.Groups)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read group file")
	}
	var items []types.Group
	if err := json.Unmarshal(buf, &items); err != nil {
		return nil, errors.Wrapf(err, "unable to decode file: %q", string(buf))
	}
	return items, nil
}

func New(config json.RawMessage) (*Provider, error) {
	var c Config
	if len(config) != 0 {
		if err := json.Unmarshal(config, &c); err != nil {
			return nil, errors.Wrap(err, "unable to decode provider config")
		}
	}

	return &Provider{config: c}, nil
}
