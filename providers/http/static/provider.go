package static

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Eun/nss_http/providers/http/helpers"
	"github.com/Eun/nss_http/types"
	"github.com/pkg/errors"
)

const Name = "http_static"
const maxResponseSize = 1024 * 1024 * 1024

var _ types.Provider = &Provider{}

type Provider struct {
	config Config
}

func (f *Provider) Name() string {
	return Name
}

func (f *Provider) GetUser(ctx context.Context, identifier any) (*types.User, error) {
	body, err := helpers.DoRequest(ctx, f.config.UsersURL, time.Duration(f.config.RequestTimeout), f.config.Headers, maxResponseSize)
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}

	var users []types.User
	if err := json.Unmarshal(body, &users); err != nil {
		return nil, errors.Wrapf(err, "unable to decode server response: %q", string(body))
	}

	switch v := identifier.(type) {
	case types.UIDIdentifier:
		for i := range users {
			if users[i].Uid == uint(v) {
				return &users[i], nil
			}
		}
		return nil, nil
	case types.NameIdentifier:
		for i := range users {
			if users[i].User == string(v) {
				return &users[i], nil
			}
		}
		return nil, nil
	default:
		return nil, errors.New("unknown identifier type")
	}
}

func New(config json.RawMessage) (*Provider, error) {
	var c Config
	if len(config) != 0 {
		if err := json.Unmarshal(config, &c); err != nil {
			return nil, errors.Wrap(err, "unable to decode provider config")
		}
	}

	if c.UsersURL == "" {
		return nil, errors.New("UsersURL is not set in config")
	}
	if c.GroupsURL == "" {
		return nil, errors.New("GroupsURL is not set in config")
	}

	if c.RequestTimeout == 0 {
		c.RequestTimeout = types.Duration(time.Minute)
	}

	return &Provider{config: c}, nil
}
