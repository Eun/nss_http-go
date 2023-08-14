package types

type User struct {
	User     string   `json:"User"`
	Passwd   string   `json:"Passwd"`
	Name     string   `json:"Name"`
	Dir      string   `json:"Dir"`
	Shell    string   `json:"Shell"`
	Uid      uint     `json:"Uid"`
	Gid      uint     `json:"Gid"`
	AuthKeys []string `json:"AuthKeys"`
}
