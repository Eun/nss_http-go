package file_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	fileprovider "github.com/Eun/nss_http/providers/file"
	"github.com/Eun/nss_http/types"
	"github.com/stretchr/testify/require"
)

func TestProvider(t *testing.T) {
	const nonExistentUserName = types.NameIdentifier("alice")
	const nonExistentUID = types.UIDIdentifier(5000)

	const nonExistentGroupName = types.NameIdentifier("alice")
	const nonExistentGID = types.GIDIdentifier(5000)

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

	usersFile, err := os.CreateTemp("", "nss_http_file_users_")
	require.NoError(t, err)
	require.NoError(t, json.NewEncoder(usersFile).Encode([]types.User{user}))
	require.NoError(t, usersFile.Close())
	defer os.Remove(usersFile.Name())

	groupsFile, err := os.CreateTemp("", "nss_http_file_groups_")
	require.NoError(t, err)
	require.NoError(t, json.NewEncoder(groupsFile).Encode([]types.Group{group}))
	require.NoError(t, groupsFile.Close())
	defer os.Remove(groupsFile.Name())

	client, err := fileprovider.New(json.RawMessage(fmt.Sprintf(`{
"Users": %q,
"Groups": %q
}`, usersFile.Name(), groupsFile.Name())))
	require.NoError(t, err)

	t.Run("get user", func(t *testing.T) {
		t.Run("known user", func(t *testing.T) {
			t.Run("by uid", func(t *testing.T) {
				gotUser, err := client.GetUser(context.Background(), types.UIDIdentifier(user.Uid))
				require.NoError(t, err)
				require.Equal(t, &user, gotUser)
			})
			t.Run("by name", func(t *testing.T) {
				gotUser, err := client.GetUser(context.Background(), types.NameIdentifier(user.User))
				require.NoError(t, err)
				require.Equal(t, &user, gotUser)
			})
		})
		t.Run("unknown user", func(t *testing.T) {
			t.Run("by uid", func(t *testing.T) {
				gotUser, err := client.GetUser(context.Background(), nonExistentUID)
				require.NoError(t, err)
				require.Nil(t, gotUser)
			})
			t.Run("by name", func(t *testing.T) {
				gotUser, err := client.GetUser(context.Background(), nonExistentUserName)
				require.NoError(t, err)
				require.Nil(t, gotUser)
			})
		})
	})

	t.Run("get users", func(t *testing.T) {
		gotUsers, err := client.GetUsers(context.Background())
		require.NoError(t, err)
		require.Equal(t, types.Users{user}, gotUsers)
	})

	t.Run("get group", func(t *testing.T) {
		t.Run("known group", func(t *testing.T) {
			t.Run("by gid", func(t *testing.T) {
				gotGroup, err := client.GetGroup(context.Background(), types.GIDIdentifier(group.Gid))
				require.NoError(t, err)
				require.Equal(t, &group, gotGroup)
			})
			t.Run("by name", func(t *testing.T) {
				gotGroup, err := client.GetGroup(context.Background(), types.NameIdentifier(group.Name))
				require.NoError(t, err)
				require.Equal(t, &group, gotGroup)
			})
		})
		t.Run("unknown group", func(t *testing.T) {
			t.Run("by gid", func(t *testing.T) {
				gotGroup, err := client.GetGroup(context.Background(), nonExistentGID)
				require.NoError(t, err)
				require.Nil(t, gotGroup)
			})
			t.Run("by name", func(t *testing.T) {
				gotGroup, err := client.GetGroup(context.Background(), nonExistentGroupName)
				require.NoError(t, err)
				require.Nil(t, gotGroup)
			})
		})
	})

	t.Run("get groups", func(t *testing.T) {
		gotGroups, err := client.GetGroups(context.Background())
		require.NoError(t, err)
		require.Equal(t, types.Groups{group}, gotGroups)
	})
}
