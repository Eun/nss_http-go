package http

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/Eun/nss_http/providers/http/helpers"
	"github.com/Eun/nss_http/types"
	"github.com/pkg/errors"
)

const Name = "http"

const maxResponseSize = 1024 * 1024

var _ types.Provider = &Provider{}

type Provider struct {
	config Config
}

func (f *Provider) Name() string {
	return Name
}

func (f *Provider) GetUser(ctx context.Context, identifier any) (*types.User, error) {
	requestURL, err := f.getRequestURL(userRequest, identifier)
	if err != nil {
		return nil, errors.Wrap(err, "unable to build request url")
	}
	if requestURL == "" {
		// get user list and lookup user from there
		users, err := f.GetUsers(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get users")
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
		}
		return nil, nil
	}
	body, err := helpers.DoRequest(ctx, requestURL, time.Duration(f.config.RequestTimeout), f.config.Headers, maxResponseSize)
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	if body == nil {
		return nil, nil
	}

	var user types.User
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, errors.Wrapf(err, "unable to decode server response: %q", string(body))
	}
	return &user, nil
}

func (f *Provider) GetUsers(ctx context.Context) ([]types.User, error) {
	if f.config.RequestUrls.Users == "" {
		return nil, nil
	}
	body, err := helpers.DoRequest(ctx, f.config.RequestUrls.Users, time.Duration(f.config.RequestTimeout), f.config.Headers, maxResponseSize)
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	if body == nil {
		return nil, nil
	}

	var users []types.User
	if err := json.Unmarshal(body, &users); err != nil {
		return nil, errors.Wrapf(err, "unable to decode server response: %q", string(body))
	}
	return users, nil
}

func (f *Provider) GetGroup(ctx context.Context, identifier any) (*types.Group, error) {
	requestURL, err := f.getRequestURL(groupRequest, identifier)
	if err != nil {
		return nil, errors.Wrap(err, "unable to build request url")
	}
	if requestURL == "" {
		// get group list and lookup group from there
		groups, err := f.GetGroups(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to get groups")
		}
		switch v := identifier.(type) {
		case types.GIDIdentifier:
			for i := range groups {
				if groups[i].Gid == uint(v) {
					return &groups[i], nil
				}
			}
			return nil, nil
		case types.NameIdentifier:
			for i := range groups {
				if groups[i].Name == string(v) {
					return &groups[i], nil
				}
			}
			return nil, nil
		}
		return nil, nil
	}
	body, err := helpers.DoRequest(ctx, requestURL, time.Duration(f.config.RequestTimeout), f.config.Headers, maxResponseSize)
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	if body == nil {
		return nil, nil
	}

	var group types.Group
	if err := json.Unmarshal(body, &group); err != nil {
		return nil, errors.Wrapf(err, "unable to decode server response: %q", string(body))
	}
	return &group, nil
}

func (f *Provider) GetGroups(ctx context.Context) ([]types.Group, error) {
	if f.config.RequestUrls.Groups == "" {
		return nil, nil
	}
	body, err := helpers.DoRequest(ctx, f.config.RequestUrls.Groups, time.Duration(f.config.RequestTimeout), f.config.Headers, maxResponseSize)
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	if body == nil {
		return nil, nil
	}

	var groups []types.Group
	if err := json.Unmarshal(body, &groups); err != nil {
		return nil, errors.Wrapf(err, "unable to decode server response: %q", string(body))
	}
	return groups, nil
}

type requestSection int

const (
	userRequest  requestSection = iota
	groupRequest requestSection = iota
)

func (f *Provider) getRequestURL(section requestSection, identifier any) (string, error) {
	switch section {
	case userRequest:
		switch v := identifier.(type) {
		case types.UIDIdentifier:
			if f.config.RequestUrls.UserUID == "" {
				return "", nil
			}
			return f.config.RequestUrls.UserUID + strconv.FormatUint(uint64(v), 10), nil
		case types.NameIdentifier:
			if f.config.RequestUrls.UserName == "" {
				return "", nil
			}
			return f.config.RequestUrls.UserName + string(v), nil
		default:
			return "", errors.New("unknown identifier type")
		}
	case groupRequest:
		switch v := identifier.(type) {
		case types.GIDIdentifier:
			if f.config.RequestUrls.GroupGID == "" {
				return "", nil
			}
			return f.config.RequestUrls.GroupGID + strconv.FormatUint(uint64(v), 10), nil
		case types.NameIdentifier:
			if f.config.RequestUrls.GroupName == "" {
				return "", nil
			}
			return f.config.RequestUrls.GroupName + string(v), nil
		default:
			return "", errors.New("unknown identifier type")
		}
	default:
		return "", errors.New("unknown request section")
	}
}

func New(config json.RawMessage) (*Provider, error) {
	var c Config
	if len(config) != 0 {
		if err := json.Unmarshal(config, &c); err != nil {
			return nil, errors.Wrap(err, "unable to decode provider config")
		}
	}

	if c.RequestTimeout == 0 {
		c.RequestTimeout = types.Duration(time.Minute)
	}

	return &Provider{config: c}, nil
}
