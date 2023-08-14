package static

import (
	"github.com/Eun/nss_http/types"
)

type Config struct {
	UsersURL       string            `json:"UsersURL"`
	GroupsURL      string            `json:"GroupsURL"`
	RequestTimeout types.Duration    `json:"RequestTimeout"`
	Headers        map[string]string `json:"Headers"`
}
