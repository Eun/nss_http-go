package main

import (
	"context"

	"github.com/Eun/nss_http/config"
	"github.com/Eun/nss_http/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func getUser(identifer any) (*types.User, error) {
	config, err := config.Get()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get config")
	}

	user, err := config.CacheProvider.GetUser(context.Background(), identifer)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get user from cache")
	}
	if user != nil && user.User != "" {
		log.Debug().
			Str("provider", config.CacheProvider.Name()).
			Any("user", "user").
			Any("identifer", identifer).
			Msg("found user in cache")
	}

	var provider types.Provider
	for _, provider = range config.Providers {
		user, err = provider.GetUser(context.Background(), identifer)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get user")
		}
		if user != nil && user.User != "" {
			break
		}
	}

	if user == nil {
		log.Debug().
			Any("identifer", identifer).
			Msg("found no user")
		return nil, nil
	}

	log.Debug().
		Str("provider", provider.Name()).
		Any("user", "user").
		Any("identifer", identifer).
		Msg("found user")
	if err := config.CacheProvider.SetUser(context.Background(), user); err != nil {
		return nil, errors.Wrap(err, "unable to add to cache")

	}
	return user, nil

}
