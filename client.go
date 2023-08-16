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
			Any("user", user).
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
		Any("user", user).
		Any("identifer", identifer).
		Msg("found user")
	if err := config.CacheProvider.SetUser(context.Background(), user); err != nil {
		return nil, errors.Wrap(err, "unable to add to cache")

	}
	return user, nil
}

func getUsers() ([]types.User, error) {
	config, err := config.Get()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get config")
	}
	if !config.AllowListingOfUsers {
		return nil, nil
	}
	users, err := config.CacheProvider.GetUsers(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "unable to get users from cache")
	}
	if len(users) > 0 {
		return users, nil
	}
	for _, provider := range config.Providers {
		providerUsers, err := provider.GetUsers(context.Background())
		if err != nil {
			return nil, errors.Wrap(err, "unable to get users")
		}
		log.Debug().
			Str("provider", provider.Name()).
			Any("users", providerUsers).
			Msg("found users")
		users = append(users, providerUsers...)
	}

	log.Debug().
		Any("users", users).
		Msg("found users")

	if err := config.CacheProvider.SetUsers(context.Background(), users); err != nil {
		return nil, errors.Wrap(err, "unable to add to cache")

	}
	return users, nil
}

func getGroup(identifer any) (*types.Group, error) {
	config, err := config.Get()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get config")
	}

	group, err := config.CacheProvider.GetGroup(context.Background(), identifer)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get group from cache")
	}
	if group != nil && group.Name != "" {
		log.Debug().
			Str("provider", config.CacheProvider.Name()).
			Any("group", group).
			Any("identifer", identifer).
			Msg("found group in cache")
	}

	var provider types.Provider
	for _, provider = range config.Providers {
		group, err = provider.GetGroup(context.Background(), identifer)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get group")
		}
		if group != nil && group.Name != "" {
			break
		}
	}

	if group == nil {
		log.Debug().
			Any("identifer", identifer).
			Msg("found no group")
		return nil, nil
	}

	log.Debug().
		Str("provider", provider.Name()).
		Any("group", group).
		Any("identifer", identifer).
		Msg("found group")
	if err := config.CacheProvider.SetGroup(context.Background(), group); err != nil {
		return nil, errors.Wrap(err, "unable to add to cache")

	}
	return group, nil
}

func getGroups() ([]types.Group, error) {
	config, err := config.Get()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get config")
	}
	if !config.AllowListingOfGroups {
		return nil, nil
	}
	groups, err := config.CacheProvider.GetGroups(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "unable to get groups from cache")
	}
	if len(groups) > 0 {
		return groups, nil
	}
	for _, provider := range config.Providers {
		providerGroups, err := provider.GetGroups(context.Background())
		if err != nil {
			return nil, errors.Wrap(err, "unable to get groups")
		}
		log.Debug().
			Str("provider", provider.Name()).
			Any("groups", providerGroups).
			Msg("found groups")
		groups = append(groups, providerGroups...)
	}

	log.Debug().
		Any("groups", groups).
		Msg("found groups")

	if err := config.CacheProvider.SetGroups(context.Background(), groups); err != nil {
		return nil, errors.Wrap(err, "unable to add to cache")

	}
	return groups, nil
}
