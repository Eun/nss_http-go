package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/Eun/nss_http/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const Name = "http"

const maxResponseSize = 1024 * 1024 * 1024

var _ types.Provider = &Provider{}

type Provider struct {
	config Config
}

func (p *Provider) Name() string {
	return Name
}

func (p *Provider) GetUser(ctx context.Context, identifier any) (*types.User, error) {
	requestURL, err := p.getRequestURL(userRequest, identifier)
	if err != nil {
		return nil, errors.Wrap(err, "unable to build request url")
	}
	if requestURL == "" {
		// get user list and lookup user from there
		users, err := p.GetUsers(ctx)
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
		return nil, errors.New("unknown identifier type")
	}
	body, err := doRequest(ctx, requestURL, time.Duration(p.config.RequestTimeout), p.config.Headers, maxResponseSize)
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

func (p *Provider) GetUsers(ctx context.Context) (types.Users, error) {
	if p.config.URLs.Users == "" {
		return nil, nil
	}
	body, err := doRequest(ctx, p.config.URLs.Users, time.Duration(p.config.RequestTimeout), p.config.Headers, maxResponseSize)
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	if body == nil {
		return nil, nil
	}

	var users types.Users
	if err := json.Unmarshal(body, &users); err != nil {
		return nil, errors.Wrapf(err, "unable to decode server response: %q", string(body))
	}
	return users, nil
}

func (p *Provider) GetGroup(ctx context.Context, identifier any) (*types.Group, error) {
	requestURL, err := p.getRequestURL(groupRequest, identifier)
	if err != nil {
		return nil, errors.Wrap(err, "unable to build request url")
	}
	if requestURL == "" {
		// get group list and lookup group from there
		groups, err := p.GetGroups(ctx)
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
		return nil, errors.New("unknown identifier type")
	}
	body, err := doRequest(ctx, requestURL, time.Duration(p.config.RequestTimeout), p.config.Headers, maxResponseSize)
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

func (p *Provider) GetGroups(ctx context.Context) (types.Groups, error) {
	if p.config.URLs.Groups == "" {
		return nil, nil
	}
	body, err := doRequest(ctx, p.config.URLs.Groups, time.Duration(p.config.RequestTimeout), p.config.Headers, maxResponseSize)
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	if body == nil {
		return nil, nil
	}

	var groups types.Groups
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

func (p *Provider) getRequestURL(section requestSection, identifier any) (string, error) {
	switch section {
	case userRequest:
		switch v := identifier.(type) {
		case types.UIDIdentifier:
			if p.config.URLs.UserUID == "" {
				return "", nil
			}
			return p.config.URLs.UserUID + strconv.FormatUint(uint64(v), 10), nil
		case types.NameIdentifier:
			if p.config.URLs.UserName == "" {
				return "", nil
			}
			return p.config.URLs.UserName + string(v), nil
		default:
			return "", errors.New("unknown identifier type")
		}
	case groupRequest:
		switch v := identifier.(type) {
		case types.GIDIdentifier:
			if p.config.URLs.GroupGID == "" {
				return "", nil
			}
			return p.config.URLs.GroupGID + strconv.FormatUint(uint64(v), 10), nil
		case types.NameIdentifier:
			if p.config.URLs.GroupName == "" {
				return "", nil
			}
			return p.config.URLs.GroupName + string(v), nil
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

func doRequest(parentContext context.Context, requestURL string, requestTimeout time.Duration, headers map[string]string, maxResponseSize int64) ([]byte, error) {
	logger := log.With().Str("url", requestURL).Logger()

	logger.Debug().Msg("getting file from server")
	ctx, cancel := context.WithTimeout(parentContext, requestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, http.NoBody)
	if err != nil {
		return nil, errors.Wrap(err, "unable to build request")
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}

	logger.Debug().Msgf("server responded with %d", res.StatusCode)
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.Wrapf(err, "expected status 200, but got %d", res.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(res.Body, maxResponseSize))
	if err != nil {
		return nil, errors.Wrap(err, "unable to read body")
	}

	return body, nil
}
