package libtest

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Eun/nss_http/types"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"golang.org/x/crypto/ssh"
)

func TestLib(t *testing.T) {
	privateKeyBytes, err := os.ReadFile("id_rsa")
	require.NoError(t, err)

	privateSSHKey, err := ssh.ParsePrivateKey(privateKeyBytes)
	require.NoError(t, err)

	mux := http.NewServeMux()

	user := types.User{
		User:     "joe",
		Passwd:   "$6$.WdgkoyPbvxIDDKU$mOVy8BlNvGssTojiLDyo37S7/puNMBx53S4VAp1nhxSnV5G7bzZw42QxbcYiq4TJwReY0cBLQGc5Dt6Mnk4lg1", // joe
		Name:     "Joe Doe",
		Dir:      "/home/joe",
		Shell:    "/bin/bash",
		Uid:      3000,
		Gid:      3000,
		AuthKeys: []string{strings.TrimSpace(string(ssh.MarshalAuthorizedKey(privateSSHKey.PublicKey())))},
	}

	group := types.Group{
		Name:         "joe",
		Passwd:       "",
		Gid:          3000,
		GroupMembers: []string{},
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
	_, httpServerPort, err := net.SplitHostPort(s.Listener.Addr().String())
	require.NoError(t, err)

	t.Run("full feature set", func(t *testing.T) {
		redisContainer, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        "redis:7",
				ExposedPorts: []string{"6379/tcp"},
				HostConfigModifier: func(config *container.HostConfig) {
					config.AutoRemove = true
				},
			},
			Started: true,
		})
		require.NoError(t, err)
		defer redisContainer.Terminate(context.Background())

		redisPort, err := redisContainer.MappedPort(context.Background(), "6379/tcp")
		require.NoError(t, err)

		container, err := NewTestContainer(fmt.Sprintf(`
{
	"Providers": [{
		"Name": "http",
		"URLs": {
			"UserUID": "http://host.docker.internal:%[1]s/user/uid/",
			"UserName": "http://host.docker.internal:%[1]s/user/name/",
			"Users": "http://host.docker.internal:%[1]s/users",
			"GroupGID": "http://host.docker.internal:%[1]s/group/gid/",
			"GroupName": "http://host.docker.internal:%[1]s/group/name/",
			"Groups": "http://host.docker.internal:%[1]s/groups"
		},
		"Headers": {}
	}],
	"Cache": {
		"Name": "redis",
		"URL":  "redis://host.docker.internal:%[2]s"
	},
	"AllowListingOfUsers": true,
	"AllowListingOfGroups": true,
	"DisableShadow": false
}
`, httpServerPort, redisPort.Port()))
		require.NoError(t, err)
		defer container.Close()
		defer func() {
			logs, err := container.GetLogs()
			if err == nil && strings.TrimSpace(logs) != "" {
				fmt.Println(logs)
			}
		}()

		GetSpecificUser(t, &user, false, container)
		UserInUserList(t, &user, false, container)
		GetSpecificGroup(t, &group, false, container)
		GroupInGroupList(t, &group, false, container)
	})

	t.Run("only users and groups url specified", func(t *testing.T) {
		container, err := NewTestContainer(fmt.Sprintf(`
{
	"Providers": [{
		"Name": "http",
		"URLs": {
			"Users": "http://host.docker.internal:%[1]s/users",
			"Groups": "http://host.docker.internal:%[1]s/groups"
		},
		"Headers": {}
	}],
	"Cache": {
		"Name": "disabled"
	},
	"AllowListingOfUsers": true,
	"AllowListingOfGroups": true,
	"DisableShadow": false
}
`, httpServerPort))
		require.NoError(t, err)
		defer container.Close()
		defer func() {
			logs, err := container.GetLogs()
			if err == nil && strings.TrimSpace(logs) != "" {
				fmt.Println(logs)
			}
		}()

		GetSpecificUser(t, &user, false, container)
		UserInUserList(t, &user, false, container)
		GetSpecificGroup(t, &group, false, container)
		GroupInGroupList(t, &group, false, container)
	})

	t.Run("only individual users and groups url specified", func(t *testing.T) {
		container, err := NewTestContainer(fmt.Sprintf(`
{
	"Providers": [{
		"Name": "http",
		"URLs": {
			"UserUID": "http://host.docker.internal:%[1]s/user/uid/",
			"UserName": "http://host.docker.internal:%[1]s/user/name/",
			"GroupGID": "http://host.docker.internal:%[1]s/group/gid/",
			"GroupName": "http://host.docker.internal:%[1]s/group/name/"
		},
		"Headers": {}
	}],
	"Cache": {
		"Name": "disabled"
	},
	"AllowListingOfUsers": true,
	"AllowListingOfGroups": true,
	"DisableShadow": false
}
`, httpServerPort))
		require.NoError(t, err)
		defer container.Close()
		defer func() {
			logs, err := container.GetLogs()
			if err == nil && strings.TrimSpace(logs) != "" {
				fmt.Println(logs)
			}
		}()

		GetSpecificUser(t, &user, false, container)
		UserNotInUserList(t, &user, false, container)
		GetSpecificGroup(t, &group, false, container)
		GroupNotInGroupList(t, &group, false, container)
	})

	t.Run("disallow listing of users & groups", func(t *testing.T) {
		container, err := NewTestContainer(fmt.Sprintf(`
{
	"Providers": [{
		"Name": "http",
		"URLs": {
			"Users": "http://host.docker.internal:%[1]s/users",
			"Groups": "http://host.docker.internal:%[1]s/groups"
		},
		"Headers": {}
	}],
	"Cache": {
		"Name": "disabled"
	},
	"AllowListingOfUsers": false,
	"AllowListingOfGroups": false,
	"DisableShadow": false
}
`, httpServerPort))
		require.NoError(t, err)
		defer container.Close()
		defer func() {
			logs, err := container.GetLogs()
			if err == nil && strings.TrimSpace(logs) != "" {
				fmt.Println(logs)
			}
		}()

		GetSpecificUser(t, &user, false, container)
		UserNotInUserList(t, &user, false, container)
		GetSpecificGroup(t, &group, false, container)
		GroupNotInGroupList(t, &group, false, container)
	})

	t.Run("get members of group", func(t *testing.T) {
		container, err := NewTestContainer(fmt.Sprintf(`
{
	"Providers": [{
		"Name": "http",
		"URLs": {
			"Users": "http://host.docker.internal:%[1]s/users",
			"Groups": "http://host.docker.internal:%[1]s/groups"
		},
		"Headers": {}
	}],
	"Cache": {
		"Name": "disabled"
	},
	"AllowListingOfUsers": true,
	"AllowListingOfGroups": true,
	"DisableShadow": false
}
`, httpServerPort))
		require.NoError(t, err)
		defer container.Close()
		defer func() {
			logs, err := container.GetLogs()
			if err == nil && strings.TrimSpace(logs) != "" {
				fmt.Println(logs)
			}
		}()

		IsUserMemberOfGroup(t, &user, &group, container)
	})

	t.Run("test ssh login", func(t *testing.T) {
		t.Run("private key", func(t *testing.T) {
			container, err := NewTestContainer("")
			require.NoError(t, err)
			defer container.Close()
			defer func() {
				logs, err := container.GetLogs()
				if err == nil && strings.TrimSpace(logs) != "" {
					fmt.Println(logs)
				}
			}()

			addr, err := container.SSHAddr()
			require.NoError(t, err)

			client, err := ssh.Dial("tcp", addr, &ssh.ClientConfig{
				Config: ssh.Config{
					Rand:           nil,
					RekeyThreshold: 0,
					KeyExchanges:   nil,
					Ciphers:        nil,
					MACs:           nil,
				},
				User:            user.User,
				Auth:            []ssh.AuthMethod{ssh.PublicKeys(privateSSHKey)},
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				BannerCallback:  nil,
				ClientVersion:   "",
				Timeout:         time.Minute,
			})
			require.NoError(t, err)
			defer client.Close()
			sess, err := client.NewSession()
			require.NoError(t, err)
			defer sess.Close()
			buf, err := sess.CombinedOutput("whoami")
			require.NoError(t, err)
			require.Equal(t, user.User, strings.TrimSpace(string(buf)))
		})
		t.Run("password", func(t *testing.T) {
			container, err := NewTestContainer("")
			require.NoError(t, err)
			defer container.Close()
			defer func() {
				logs, err := container.GetLogs()
				if err == nil && strings.TrimSpace(logs) != "" {
					fmt.Println(logs)
				}
			}()

			addr, err := container.SSHAddr()
			require.NoError(t, err)

			client, err := ssh.Dial("tcp", addr, &ssh.ClientConfig{
				Config: ssh.Config{
					Rand:           nil,
					RekeyThreshold: 0,
					KeyExchanges:   nil,
					Ciphers:        nil,
					MACs:           nil,
				},
				User:            user.User,
				Auth:            []ssh.AuthMethod{ssh.Password("joe")},
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				BannerCallback:  nil,
				ClientVersion:   "",
				Timeout:         time.Minute,
			})
			require.NoError(t, err)
			defer client.Close()
			sess, err := client.NewSession()
			require.NoError(t, err)
			defer sess.Close()
			buf, err := sess.CombinedOutput("whoami")
			require.NoError(t, err)
			require.Equal(t, user.User, strings.TrimSpace(string(buf)))
		})

	})
}
