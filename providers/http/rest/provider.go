package rest

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"time"

	"github.com/Eun/nss_http/providers/http/helpers"
	"github.com/Eun/nss_http/types"
	"github.com/pkg/errors"
)

const Name = "http_rest"

const maxResponseSize = 1024 * 1024

var _ types.Provider = &Provider{}

type Provider struct {
	config Config
}

func (f *Provider) Name() string {
	return Name
}

func (f *Provider) GetUser(ctx context.Context, identifier any) (*types.User, error) {
	requestURL, err := f.getRequestURL("user", identifier)
	if err != nil {
		return nil, errors.Wrap(err, "unable to build request url")
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

func (f *Provider) getRequestURL(section string, identifier any) (string, error) {
	switch v := identifier.(type) {
	case types.UIDIdentifier:
		requestURL, err := url.JoinPath(f.config.URL, section, "uid", strconv.FormatUint(uint64(v), 10))
		if err != nil {
			return "", errors.Wrap(err, "unable to build request url")
		}
		return requestURL, nil

	case types.NameIdentifier:
		requestURL, err := url.JoinPath(f.config.URL, section, "name", string(v))
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

	if c.RequestTimeout == 0 {
		c.RequestTimeout = types.Duration(time.Minute)
	}

	return &Provider{config: c}, nil
}
