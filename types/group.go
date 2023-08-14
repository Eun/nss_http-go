package types

type Group struct {
	Name         string   `json:"Name"`
	Passwd       string   `json:"Passwd"`
	Gid          uint     `json:"Gid"`
	GroupMembers []string `json:"GroupMembers"`
}
