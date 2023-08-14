package rest

import "github.com/Eun/nss_http/types"

type Config struct {
	URL            string            `json:"URL"`
	RequestTimeout types.Duration    `json:"RequestTimeout"`
	Headers        map[string]string `json:"Headers"`
}
