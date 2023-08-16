package mem

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	types "github.com/Eun/nss_http/types"
	"github.com/pkg/errors"
)

const Name = "mem"

var _ types.CacheProvider = &Provider{}

type Provider struct {
	config   Config
	userMap  sync.Map
	groupMap sync.Map
}

type cacheItem[T any] struct {
	Item   T
	Expiry time.Time
}

func (p *Provider) getKey(identifier any) (string, error) {
	switch v := identifier.(type) {
	case types.UIDIdentifier:
		key, err := url.JoinPath("id", strconv.FormatUint(uint64(v), 10))
		if err != nil {
			return "", errors.Wrap(err, "unable to build request url")
		}
		return key, nil
	case types.GIDIdentifier:
		key, err := url.JoinPath("id", strconv.FormatUint(uint64(v), 10))
		if err != nil {
			return "", errors.Wrap(err, "unable to build request url")
		}
		return key, nil

	case types.NameIdentifier:
		key, err := url.JoinPath("name", string(v))
		if err != nil {
			return "", errors.Wrap(err, "unable to build request url")
		}

		return key, nil
	default:
		return "", errors.New("unknown identifier type")
	}
}

func (p *Provider) Name() string {
	return Name
}

func (p *Provider) GetUser(ctx context.Context, identifier any) (*types.User, error) {
	key, err := p.getKey(identifier)
	if err != nil {
		return nil, errors.Wrap(err, "unable to build request url")
	}
	v, ok := p.userMap.Load(key)
	if !ok {
		return nil, nil
	}
	item, ok := v.(cacheItem[types.User])
	if !ok {
		return nil, nil
	}

	if time.Now().After(item.Expiry) {
		return nil, nil
	}
	return &item.Item, nil
}

func (p *Provider) SetUser(ctx context.Context, user *types.User) error {
	for _, i := range []any{types.NameIdentifier(user.User), types.UIDIdentifier(user.Uid)} {
		key, err := p.getKey(i)
		if err != nil {
			return errors.Wrap(err, "unable to build request url")
		}
		p.userMap.Store(key, cacheItem[types.User]{
			Item:   *user,
			Expiry: time.Now().Add(time.Duration(p.config.TTL)),
		})
	}
	return nil
}

func (p *Provider) GetUsers(ctx context.Context) ([]types.User, error) {
	exitedPrematurely := false
	var items []types.User
	p.userMap.Range(func(key, value any) bool {
		s, ok := key.(string)
		if !ok {
			return true
		}
		if !strings.HasPrefix(s, "id/") {
			return true
		}
		item, ok := value.(cacheItem[types.User])
		if !ok {
			exitedPrematurely = true
			return false
		}

		if time.Now().After(item.Expiry) {
			exitedPrematurely = true
			return false
		}
		items = append(items, item.Item)
		return true
	})

	if exitedPrematurely {
		return nil, nil
	}
	return items, nil
}

func (p *Provider) SetUsers(ctx context.Context, users []types.User) error {
	for _, user := range users {
		if err := p.SetUser(ctx, &user); err != nil {
			return errors.Wrap(err, "unable to cache user")
		}
	}
	return nil
}

func (p *Provider) GetGroup(ctx context.Context, identifier any) (*types.Group, error) {
	key, err := p.getKey(identifier)
	if err != nil {
		return nil, errors.Wrap(err, "unable to build request url")
	}
	v, ok := p.groupMap.Load(key)
	if !ok {
		return nil, nil
	}
	item, ok := v.(cacheItem[types.Group])
	if !ok {
		return nil, nil
	}

	if time.Now().After(item.Expiry) {
		return nil, nil
	}
	return &item.Item, nil
}

func (p *Provider) SetGroup(ctx context.Context, group *types.Group) error {
	for _, i := range []any{types.NameIdentifier(group.Name), types.GIDIdentifier(group.Gid)} {
		key, err := p.getKey(i)
		if err != nil {
			return errors.Wrap(err, "unable to build request url")
		}
		p.groupMap.Store(key, cacheItem[types.Group]{
			Item:   *group,
			Expiry: time.Now().Add(time.Duration(p.config.TTL)),
		})
	}
	return nil
}

func (p *Provider) GetGroups(ctx context.Context) ([]types.Group, error) {
	exitedPrematurely := false
	var items []types.Group
	p.groupMap.Range(func(key, value any) bool {
		s, ok := key.(string)
		if !ok {
			return true
		}
		if !strings.HasPrefix(s, "id/") {
			return true
		}
		item, ok := value.(cacheItem[types.Group])
		if !ok {
			exitedPrematurely = true
			return false
		}

		if time.Now().After(item.Expiry) {
			exitedPrematurely = true
			return false
		}
		items = append(items, item.Item)
		return true
	})

	if exitedPrematurely {
		return nil, nil
	}
	return items, nil
}

func (p *Provider) SetGroups(ctx context.Context, groups []types.Group) error {
	for _, group := range groups {
		if err := p.SetGroup(ctx, &group); err != nil {
			return errors.Wrap(err, "unable to cache group")
		}
	}
	return nil
}

func New(config json.RawMessage) (*Provider, error) {
	var c Config
	if len(config) != 0 {
		if err := json.Unmarshal(config, &c); err != nil {
			return nil, errors.Wrap(err, "unable to decode provider config")
		}
	}

	if c.TTL == 0 {
		c.TTL = types.Duration(time.Minute)
	}

	return &Provider{config: c}, nil
}
