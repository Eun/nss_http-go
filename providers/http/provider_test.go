package http_test

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Eun/nss_http/testhelpers"
	"github.com/Eun/nss_http/types"
	"github.com/stretchr/testify/require"
)

func TestProvider(t *testing.T) {
	mux := http.NewServeMux()

	user := types.User{
		User:   "joe",
		Passwd: "$6$.WdgkoyPbvxIDDKU$mOVy8BlNvGssTojiLDyo37S7/puNMBx53S4VAp1nhxSnV5G7bzZw42QxbcYiq4TJwReY0cBLQGc5Dt6Mnk4lg1",
		Name:   "Joe Doe",
		Dir:    "/home/joe",
		Shell:  "/bin/bash",
		Uid:    3000,
		Gid:    3000,
		AuthKeys: []string{
			"ssh-rsa AAAAB3NzaC1yc2EAAAABJQAAAQEA0iftI+BgPr2F0aX1ajg89TVsGZ/nKvyayWSHpynE53s+5KABK7ns66hZyQWe/Q2g15Npebdz7xXES752ch3CtPiBOe//aehUZTYj/rxY9z7/JoYBYeoGeTHNtZ1D6IWnH3Aw9xeAPx1utHPMNRNkXiJWL7FaUKFMsLSK8irc/nyiJrJ5kIoGZBLMXu9CQY775DXP9xkT23nvV2RQw/9Sr20EFtSATqUrYab0/3V4Sb+W4fIAa4zGo75GC8Xpgb0gO96atJzVD8s3bpOFXl+F52JUZXyCU4B68gO+1Chv/TT6TuLU0jtVBJorvincc573NrmSNdhavmJBbHqJ6CXx5Q== rsa-key-20160901",
		},
	}

	group := types.Group{
		Name:   "joe",
		Passwd: "$6$.WdgkoyPbvxIDDKU$mOVy8BlNvGssTojiLDyo37S7/puNMBx53S4VAp1nhxSnV5G7bzZw42QxbcYiq4TJwReY0cBLQGc5Dt6Mnk4lg1",
		Gid:    3000,
	}

	mux.HandleFunc("/user/uid/3000", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(user)
	})
	mux.HandleFunc("/user/name/joe", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(user)
	})
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]types.User{user})
	})

	mux.HandleFunc("/group/gid/3000", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(group)
	})
	mux.HandleFunc("/group/name/joe", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(group)
	})
	mux.HandleFunc("/groups", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]types.Group{group})
	})

	s := httptest.NewServer(mux)
	defer s.Close()
	_, port, err := net.SplitHostPort(s.Listener.Addr().String())
	require.NoError(t, err)

	container, err := testhelpers.NewTestContainer(fmt.Sprintf(`
{
	"Providers": [{
		"Name": "http",
		"RequestURLs": {
			"UserUID": "http://host.docker.internal:%[1]s/user/uid/",
			"UserName": "http://host.docker.internal:%[1]s/user/name/",
			"Users": "http://host.docker.internal:%[1]s/users",
			"GroupUID": "http://host.docker.internal:%[1]s/group/uid/",
			"GroupName": "http://host.docker.internal:%[1]s/group/name/",
			"Groups": "http://host.docker.internal:%[1]s/groups"
		},
		"Headers": {}
	}],
	"Cache": {
		"Name": "disabled"
	},
	"AllowListingOfUsers": true,
	"AllowListingOfGroups": true
}
`, port))
	require.NoError(t, err)
	defer container.Close()
	defer func() {
		logs, err := container.GetLogs()
		if err == nil && strings.TrimSpace(logs) != "" {
			fmt.Println(logs)
		}
	}()

	testhelpers.RunUserTests(t, &user, container)
	testhelpers.RunGroupTests(t, &group, container)
}
