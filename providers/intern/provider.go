package intern

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"sync"
	"time"

	types "github.com/Eun/nss_http/types"
	"github.com/pkg/errors"
)

const Name = "intern"

var _ types.CacheProvider = &Provider{}

type cacheUserItem struct {
	User   types.User
	Expiry time.Time
}

type cacheGroupItem struct {
	Group  types.Group
	Expiry time.Time
}

var internalCache struct {
	userMap  sync.Map
	groupMap sync.Map
}

type Provider struct {
	config Config
}

func (f *Provider) Name() string {
	return Name
}

func (f *Provider) GetUser(ctx context.Context, identifier any) (*types.User, error) {
	key, err := f.getKey(identifier)
	if err != nil {
		return nil, errors.Wrap(err, "unable to build request url")
	}
	v, ok := internalCache.userMap.Load(key)
	if !ok {
		return nil, nil
	}
	item, ok := v.(cacheUserItem)
	if !ok {
		return nil, nil
	}

	if time.Now().After(item.Expiry) {
		return nil, nil
	}
	return &item.User, nil
}

func (f *Provider) SetUser(ctx context.Context, user *types.User) error {
	for _, i := range []any{types.NameIdentifier(user.User), types.UIDIdentifier(user.Uid)} {
		key, err := f.getKey(i)
		if err != nil {
			return errors.Wrap(err, "unable to build request url")
		}
		internalCache.userMap.Store(key, cacheUserItem{
			User:   *user,
			Expiry: time.Now().Add(time.Duration(f.config.TTL)),
		})
	}
	return nil
}

func (f *Provider) GetUsers(ctx context.Context) ([]types.User, error) {
	exitedPrematurely := false
	var users []types.User
	internalCache.userMap.Range(func(key, value any) bool {
		item, ok := value.(cacheUserItem)
		if !ok {
			exitedPrematurely = true
			return false
		}

		if time.Now().After(item.Expiry) {
			exitedPrematurely = true
			return false
		}
		users = append(users, item.User)
		return true
	})

	if exitedPrematurely {
		return nil, nil
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
	key, err := f.getKey(identifier)
	if err != nil {
		return nil, errors.Wrap(err, "unable to build request url")
	}
	v, ok := internalCache.groupMap.Load(key)
	if !ok {
		return nil, nil
	}
	item, ok := v.(cacheGroupItem)
	if !ok {
		return nil, nil
	}

	if time.Now().After(item.Expiry) {
		return nil, nil
	}
	return &item.Group, nil
}

func (f *Provider) SetGroup(ctx context.Context, group *types.Group) error {
	for _, i := range []any{types.NameIdentifier(group.Name), types.UIDIdentifier(group.Gid)} {
		key, err := f.getKey(i)
		if err != nil {
			return errors.Wrap(err, "unable to build request url")
		}
		internalCache.groupMap.Store(key, cacheGroupItem{
			Group:  *group,
			Expiry: time.Now().Add(time.Duration(f.config.TTL)),
		})
	}
	return nil
}

func (f *Provider) GetGroups(ctx context.Context) ([]types.Group, error) {
	exitedPrematurely := false
	var groups []types.Group
	internalCache.groupMap.Range(func(key, value any) bool {
		item, ok := value.(cacheGroupItem)
		if !ok {
			exitedPrematurely = true
			return false
		}

		if time.Now().After(item.Expiry) {
			exitedPrematurely = true
			return false
		}
		groups = append(groups, item.Group)
		return true
	})

	if exitedPrematurely {
		return nil, nil
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

func (f *Provider) getKey(identifier any) (string, error) {
	switch v := identifier.(type) {
	case types.UIDIdentifier:
		requestURL, err := url.JoinPath("uid", strconv.FormatUint(uint64(v), 10))
		if err != nil {
			return "", errors.Wrap(err, "unable to build request url")
		}
		return requestURL, nil

	case types.NameIdentifier:
		requestURL, err := url.JoinPath("name", string(v))
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

	if c.TTL == 0 {
		c.TTL = types.Duration(time.Minute)
	}

	return &Provider{config: c}, nil
}
