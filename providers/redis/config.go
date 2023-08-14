package redis

import "github.com/Eun/nss_http/types"

type Config struct {
	URL string         `json:"URL"`
	TTL types.Duration `json:"TTL"`
}
