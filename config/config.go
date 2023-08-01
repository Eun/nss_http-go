package config

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Eun/nss_http/providers/dummy"
	"github.com/Eun/nss_http/providers/http/dynamic"
	"github.com/Eun/nss_http/providers/http/static"
	"github.com/Eun/nss_http/providers/intern"
	"github.com/Eun/nss_http/providers/redis"
	"github.com/Eun/nss_http/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const configFile = "/etc/nss_http.json"

type Config struct {
	ConfigProviders []ConfigProvider `json:"Providers"`
	ConfigCache     ConfigCache      `json:"Cache"`

	Providers     []types.Provider    `json:"-"`
	CacheProvider types.CacheProvider `json:"-"`
}

type ConfigProvider struct {
	Name        string `json:"Name"`
	ExtraFields []byte `json:"-"`
}

func (f *ConfigProvider) UnmarshalJSON(data []byte) error {
	var allFields map[string]any
	err := json.Unmarshal(data, &allFields)
	if err != nil {
		return err
	}

	type alias ConfigProvider
	config := alias(*f)
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}
	*f = ConfigProvider(config)
	f.ExtraFields, err = json.Marshal(allFields)
	return err
}

type ConfigCache struct {
	ConfigProvider
	TTL time.Duration `json:"TTL"`
}

var configHolder struct {
	mu     sync.Mutex
	config *Config
}

func Get() (*Config, error) {
	configHolder.mu.Lock()
	defer configHolder.mu.Unlock()
	if configHolder.config != nil {
		return configHolder.config, nil
	}

	f, err := os.Open(configFile)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to open config file `%s'", configFile)
	}
	defer f.Close()

	configHolder.config = &Config{}
	if err = json.NewDecoder(f).Decode(configHolder.config); err != nil {
		return nil, errors.Wrapf(err, "unable to decode config file `%s'", configFile)
	}

	if err := initializeInstances(configHolder.config); err != nil {
		return nil, errors.Wrap(err, "unable to initialize config")
	}

	return configHolder.config, nil

}

func initializeInstances(config *Config) error {
	if len(config.ConfigProviders) == 0 {
		return errors.New("no providers defined")
	}

	var err error
	config.Providers = make([]types.Provider, len(config.ConfigProviders))
	for i, provider := range config.ConfigProviders {
		switch strings.ToLower(provider.Name) {
		case static.Name:
			config.Providers[i], err = static.New(provider.ExtraFields)
		case dynamic.Name:
			config.Providers[i], err = dynamic.New(provider.ExtraFields)
		case redis.Name:
			config.Providers[i], err = redis.New(provider.ExtraFields)
		default:
			return errors.Errorf("unknown provider `%s'", provider.Name)
		}
		if err != nil {
			return errors.Wrapf(err, "unable to create provider `%s'", provider.Name)
		}
		log.Debug().Msgf("initialized provider `%s'", config.Providers[i].Name())
	}

	switch strings.ToLower(config.ConfigCache.Name) {
	case intern.Name:
		config.CacheProvider, err = intern.New(config.ConfigCache.ExtraFields)
	case redis.Name:
		config.CacheProvider, err = redis.New(config.ConfigCache.ExtraFields)
	case "disabled":
		config.CacheProvider, err = dummy.New(config.ConfigCache.ExtraFields)
	default:
		config.ConfigCache.Name = "intern"
		config.CacheProvider, err = intern.New(config.ConfigCache.ExtraFields)
	}
	if err != nil {
		return errors.Wrapf(err, "unable to create cache provider `%s'", config.ConfigCache.Name)
	}
	log.Debug().Msgf("initialized cache provider `%s'", config.CacheProvider.Name())

	return nil
}
