package http

import "github.com/Eun/nss_http/types"

type Config struct {
	RequestUrls    RequestUrls       `json:"RequestUrls"`
	RequestTimeout types.Duration    `json:"RequestTimeout"`
	Headers        map[string]string `json:"Headers"`
}

type RequestUrls struct {
	Users    string `json:"Users"`
	UserUID  string `json:"UserUID"`
	UserName string `json:"UserName"`

	Groups    string `json:"Groups"`
	GroupGID  string `json:"GroupGID"`
	GroupName string `json:"GroupName"`
}
