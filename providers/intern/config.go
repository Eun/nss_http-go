package intern

import "time"

type Config struct {
	TTL time.Duration `json:"TTL"`
}
