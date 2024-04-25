package types

type Group struct {
	Name         string   `json:"Name"`
	Passwd       string   `json:"Passwd"`
	Gid          uint     `json:"Gid"`
	GroupMembers []string `json:"GroupMembers"`
}

func (g *Group) SetShadowPasswd() {
	g.Passwd = "x"
}

type Groups []Group

func (g Groups) SetShadowPasswd() {
	for i := range g {
		g[i].SetShadowPasswd()
	}
}
