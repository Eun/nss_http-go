package static

import "time"

type Config struct {
	URL            string            `json:"URL"`
	RequestTimeout time.Duration     `json:"RequestTimeout"`
	Headers        map[string]string `json:"Headers"`
	UsersPath      string            `json:"UsersPath"`
	GroupsPath     string            `json:"GroupsPath"`
}
