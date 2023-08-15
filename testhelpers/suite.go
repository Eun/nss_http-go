package testhelpers

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/Eun/nss_http/types"
	"github.com/stretchr/testify/require"
)

const nonExistentUserName = "alice"
const nonExistentUID = "5000"

const nonExistentGroupName = "alice"
const nonExistentGID = "5000"

func RunUserTests(t *testing.T, user *types.User, container *TestContainer) {
	t.Run("get user", func(t *testing.T) {
		t.Run("passwd", func(t *testing.T) {
			t.Run("known user", func(t *testing.T) {
				t.Run("by name", func(t *testing.T) {
					result, err := container.GetPasswd(user.User)
					require.NoError(t, err)
					require.Equal(t, fmt.Sprintf("%s:%s:%d:%d:%s:%s:%s", user.User, user.Passwd, user.Uid, user.Gid, user.Name, user.Dir, user.Shell), result)
				})
				t.Run("by uid", func(t *testing.T) {
					result, err := container.GetPasswd(strconv.FormatUint(uint64(user.Uid), 10))
					require.NoError(t, err)
					require.Equal(t, fmt.Sprintf("%s:%s:%d:%d:%s:%s:%s", user.User, user.Passwd, user.Uid, user.Gid, user.Name, user.Dir, user.Shell), result)
				})
			})

			t.Run("unknown user", func(t *testing.T) {
				t.Run("by name", func(t *testing.T) {
					result, err := container.GetPasswd(nonExistentUserName)
					require.NoError(t, err)
					require.Empty(t, result)
				})
				t.Run("by uid", func(t *testing.T) {
					result, err := container.GetPasswd(nonExistentUID)
					require.NoError(t, err)
					require.Empty(t, result)
				})
			})
		})
		t.Run("shadow", func(t *testing.T) {
			t.Run("known user", func(t *testing.T) {
				t.Run("by name", func(t *testing.T) {
					result, err := container.GetShadow(user.User)
					require.NoError(t, err)
					require.Equal(t, fmt.Sprintf("%s:%s:::::::0", user.User, user.Passwd), result)
				})
			})

			t.Run("unknown user", func(t *testing.T) {
				t.Run("by name", func(t *testing.T) {
					result, err := container.GetShadow(nonExistentUserName)
					require.NoError(t, err)
					require.Empty(t, result)
				})
			})
		})
	})
	t.Run("get users", func(t *testing.T) {
		t.Run("passwd", func(t *testing.T) {
			result, err := container.GetPasswd("")
			require.NoError(t, err)
			require.Empty(t, result)
		})
		t.Run("shadow", func(t *testing.T) {
			result, err := container.GetShadow("")
			require.NoError(t, err)
			require.Equal(t, fmt.Sprintf("%s:%s:::::::0", user.User, user.Passwd), result)

		})
	})
}

func RunGroupTests(t *testing.T, group *types.Group, container *TestContainer) {
	t.Run("get group", func(t *testing.T) {
		t.Run("group", func(t *testing.T) {
			t.Run("known group", func(t *testing.T) {
				t.Run("by name", func(t *testing.T) {
					result, err := container.GetGroup(group.Name)
					require.NoError(t, err)
					require.Equal(t, fmt.Sprintf("%s:%s:%d:", group.Name, group.Passwd, group.Gid), result)
				})
				t.Run("by gid", func(t *testing.T) {
					result, err := container.GetGroup(strconv.FormatUint(uint64(group.Gid), 10))
					require.NoError(t, err)
					require.Equal(t, fmt.Sprintf("%s:%s:%d:", group.Name, group.Passwd, group.Gid), result)
				})
			})

			t.Run("unknown user", func(t *testing.T) {
				t.Run("by name", func(t *testing.T) {
					result, err := container.GetGroup(nonExistentGroupName)
					require.NoError(t, err)
					require.Empty(t, result)
				})
				t.Run("by gid", func(t *testing.T) {
					result, err := container.GetGroup(nonExistentGID)
					require.NoError(t, err)
					require.Empty(t, result)
				})
			})
		})
		t.Run("shadow", func(t *testing.T) {
			t.Run("known user", func(t *testing.T) {
				t.Run("by name", func(t *testing.T) {
					result, err := container.GetGShadow(group.Name)
					require.NoError(t, err)
					require.Equal(t, fmt.Sprintf("%s:%s::", group.Name, group.Passwd), result)
				})
			})

			t.Run("unknown user", func(t *testing.T) {
				t.Run("by name", func(t *testing.T) {
					result, err := container.GetGShadow(nonExistentGroupName)
					require.NoError(t, err)
					require.Empty(t, result)
				})
			})
		})
	})
}
