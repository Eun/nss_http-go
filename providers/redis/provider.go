package redis

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"time"

	"github.com/Eun/nss_http/types"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

const Name = "redis"

var _ types.Provider = &Provider{}
var _ types.CacheProvider = &Provider{}

type Provider struct {
	config Config
	client *redis.Client
}

func (p *Provider) Name() string {
	return Name
}

func (p *Provider) getItem(ctx context.Context, key string, v any) (any, error) {
	val, err := p.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "unable to get value")
	}

	if err := json.Unmarshal(val, v); err != nil {
		return nil, errors.Wrapf(err, "unable to decode cache item: %q", string(val))
	}
	return v, nil
}

func (p *Provider) GetUser(ctx context.Context, identifier any) (*types.User, error) {
	key, err := p.getKey("users", identifier)
	if err != nil {
		return nil, errors.Wrap(err, "unable to build request url")
	}
	var item types.User
	v, err := p.getItem(ctx, key, &item)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	return &item, nil
}

func (p *Provider) SetUser(ctx context.Context, user *types.User) error {
	buf, err := json.Marshal(user)
	if err != nil {
		return errors.Wrapf(err, "unable to encode cache item: %+v", user)
	}

	for _, i := range []any{types.NameIdentifier(user.User), types.UIDIdentifier(user.Uid)} {
		key, err := p.getKey("users", i)
		if err != nil {
			return errors.Wrap(err, "unable to build request url")
		}

		if err := p.client.Set(ctx, key, buf, time.Duration(p.config.TTL)).Err(); err != nil {
			return errors.Wrap(err, "unable to set value")
		}
	}
	return nil
}

func (p *Provider) GetUsers(ctx context.Context) ([]types.User, error) {
	var items []types.User
	iter := p.client.Scan(ctx, 0, "users/id/*", 10).Iterator()
	for iter.Next(ctx) {
		var item types.User
		v, err := p.getItem(ctx, iter.Val(), &item)
		if err != nil {
			return nil, err
		}
		if v == nil {
			return nil, nil
		}
		items = append(items, item)
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
	key, err := p.getKey("groups", identifier)
	if err != nil {
		return nil, errors.Wrap(err, "unable to build request url")
	}
	var item types.Group
	v, err := p.getItem(ctx, key, &item)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	return &item, nil
}

func (p *Provider) SetGroup(ctx context.Context, group *types.Group) error {
	buf, err := json.Marshal(group)
	if err != nil {
		return errors.Wrapf(err, "unable to encode cache item: %+v", group)
	}

	for _, i := range []any{types.NameIdentifier(group.Name), types.UIDIdentifier(group.Gid)} {
		key, err := p.getKey("groups", i)
		if err != nil {
			return errors.Wrap(err, "unable to build request url")
		}

		if err := p.client.Set(ctx, key, buf, time.Duration(p.config.TTL)).Err(); err != nil {
			return errors.Wrap(err, "unable to set value")
		}
	}
	return nil
}

func (p *Provider) GetGroups(ctx context.Context) ([]types.Group, error) {
	var items []types.Group
	iter := p.client.Scan(ctx, 0, "groups/id/*", 10).Iterator()
	for iter.Next(ctx) {
		var item types.Group
		v, err := p.getItem(ctx, iter.Val(), &item)
		if err != nil {
			return nil, err
		}
		if v == nil {
			return nil, nil
		}
		items = append(items, item)
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

func (p *Provider) getKey(section string, identifier any) (string, error) {
	switch v := identifier.(type) {
	case types.UIDIdentifier:
		keyURL, err := url.JoinPath(section, "id", strconv.FormatUint(uint64(v), 10))
		if err != nil {
			return "", errors.Wrap(err, "unable to build key url")
		}
		return keyURL, nil
	case types.GIDIdentifier:
		keyURL, err := url.JoinPath(section, "id", strconv.FormatUint(uint64(v), 10))
		if err != nil {
			return "", errors.Wrap(err, "unable to build key url")
		}
		return keyURL, nil
	case types.NameIdentifier:
		keyURL, err := url.JoinPath(section, "name", string(v))
		if err != nil {
			return "", errors.Wrap(err, "unable to build key url")
		}
		return keyURL, nil
	default:
		return "", errors.New("unknown identifier type")
	}
}

func New(config json.RawMessage) (*Provider, error) {
	var c Config
	if len(config) != 0 {
		if err := json.Unmarshal(config, &c); err != nil {
			return nil, errors.Wrap(err, "unable to decode provider config")
		}
	}

	if c.URL == "" {
		return nil, errors.New("URL is not set in config")
	}

	if c.TTL == 0 {
		c.TTL = types.Duration(time.Minute)
	}

	opts, err := redis.ParseURL(c.URL)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse redis url")
	}
	client := redis.NewClient(opts)

	return &Provider{
		config: c,
		client: client,
	}, nil
}
