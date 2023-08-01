package redis

import "time"

type Config struct {
	URL string        `json:"URL"`
	TTL time.Duration `json:"TTL"`
}
