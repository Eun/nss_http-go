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

func (f *Provider) Name() string {
	return Name
}

func (f *Provider) GetUser(ctx context.Context, identifier any) (*types.User, error) {
	key, err := f.getKey("users", identifier)
	if err != nil {
		return nil, errors.Wrap(err, "unable to build request url")
	}
	val, err := f.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "unable to get value")
	}

	var user types.User
	if err := json.Unmarshal(val, &user); err != nil {
		return nil, errors.Wrapf(err, "unable to decode cache item: %q", string(val))
	}
	return &user, nil
}

func (f *Provider) SetUser(ctx context.Context, user *types.User) error {
	buf, err := json.Marshal(user)
	if err != nil {
		return errors.Wrapf(err, "unable to encode cache item: %+v", user)
	}

	for _, i := range []any{types.NameIdentifier(user.User), types.UIDIdentifier(user.Uid)} {
		key, err := f.getKey("users", i)
		if err != nil {
			return errors.Wrap(err, "unable to build request url")
		}

		if err := f.client.Set(ctx, key, buf, time.Duration(f.config.TTL)).Err(); err != nil {
			return errors.Wrap(err, "unable to set value")
		}
	}
	return nil
}

func (f *Provider) GetUsers(ctx context.Context) ([]types.User, error) {
	var users []types.User
	iter := f.client.HScan(ctx, "users", 0, "", 10).Iterator()
	for iter.Next(ctx) {
		var user types.User
		if err := json.Unmarshal([]byte(iter.Val()), &user); err != nil {
			return nil, errors.Wrapf(err, "unable to decode cache item: %q", iter.Val())
		}
		users = append(users, user)
	}
	return users, nil
}

func (f *Provider) SetUsers(ctx context.Context, users []types.User) error {
	for _, user := range users {
		if err := f.SetUser(ctx, &user); err != nil {
			return errors.Wrap(err, "unable to cache user")
		}
	}
	return nil
}

func (f *Provider) GetGroup(ctx context.Context, identifier any) (*types.Group, error) {
	key, err := f.getKey("groups", identifier)
	if err != nil {
		return nil, errors.Wrap(err, "unable to build request url")
	}
	val, err := f.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "unable to get value")
	}

	var group types.Group
	if err := json.Unmarshal(val, &group); err != nil {
		return nil, errors.Wrapf(err, "unable to decode cache item: %q", string(val))
	}
	return &group, nil
}

func (f *Provider) SetGroup(ctx context.Context, group *types.Group) error {
	buf, err := json.Marshal(group)
	if err != nil {
		return errors.Wrapf(err, "unable to encode cache item: %+v", group)
	}

	for _, i := range []any{types.NameIdentifier(group.Name), types.UIDIdentifier(group.Gid)} {
		key, err := f.getKey("groups", i)
		if err != nil {
			return errors.Wrap(err, "unable to build request url")
		}

		if err := f.client.Set(ctx, key, buf, time.Duration(f.config.TTL)).Err(); err != nil {
			return errors.Wrap(err, "unable to set value")
		}
	}
	return nil
}

func (f *Provider) GetGroups(ctx context.Context) ([]types.Group, error) {
	var groups []types.Group
	iter := f.client.HScan(ctx, "groups", 0, "", 10).Iterator()
	for iter.Next(ctx) {
		var group types.Group
		if err := json.Unmarshal([]byte(iter.Val()), &group); err != nil {
			return nil, errors.Wrapf(err, "unable to decode cache item: %q", iter.Val())
		}
		groups = append(groups, group)
	}
	return groups, nil
}

func (f *Provider) SetGroups(ctx context.Context, groups []types.Group) error {
	for _, group := range groups {
		if err := f.SetGroup(ctx, &group); err != nil {
			return errors.Wrap(err, "unable to cache group")
		}
	}
	return nil
}

func (f *Provider) getKey(section string, identifier any) (string, error) {
	switch v := identifier.(type) {
	case types.UIDIdentifier:
		requestURL, err := url.JoinPath(section, "uid", strconv.FormatUint(uint64(v), 10))
		if err != nil {
			return "", errors.Wrap(err, "unable to build request url")
		}
		return requestURL, nil

	case types.NameIdentifier:
		requestURL, err := url.JoinPath(section, "name", string(v))
		if err != nil {
			return "", errors.Wrap(err, "unable to build request url")
		}
		return requestURL, nil
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
