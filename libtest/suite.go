package libtest

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/Eun/nss_http/types"
	"github.com/stretchr/testify/require"
)

const nonExistentUserName = "alice"
const nonExistentUID = "5000"

const nonExistentGroupName = "alice"
const nonExistentGID = "5000"

func GetSpecificUser(t *testing.T, user *types.User, disableShadow bool, container *TestContainer) {
	t.Run("GetSpecificUser", func(t *testing.T) {
		t.Run("passwd", func(t *testing.T) {
			t.Run("known user", func(t *testing.T) {
				t.Run("by name", func(t *testing.T) {
					result, err := container.GetPasswd(user.User)
					require.NoError(t, err)
					passwd := user.Passwd
					if !disableShadow {
						passwd = "x"
					}
					require.Equal(t, fmt.Sprintf("%s:%s:%d:%d:%s:%s:%s", user.User, passwd, user.Uid, user.Gid, user.Name, user.Dir, user.Shell), result)
				})
				t.Run("by uid", func(t *testing.T) {
					result, err := container.GetPasswd(strconv.FormatUint(uint64(user.Uid), 10))
					require.NoError(t, err)
					passwd := user.Passwd
					if !disableShadow {
						passwd = "x"
					}
					require.Equal(t, fmt.Sprintf("%s:%s:%d:%d:%s:%s:%s", user.User, passwd, user.Uid, user.Gid, user.Name, user.Dir, user.Shell), result)
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

}

func UserInUserList(t *testing.T, user *types.User, disableShadow bool, container *TestContainer) {
	t.Run("UserInUserList", func(t *testing.T) {
		t.Run("passwd", func(t *testing.T) {
			result, err := container.GetPasswd("")
			require.NoError(t, err)
			passwd := user.Passwd
			if !disableShadow {
				passwd = "x"
			}
			require.Contains(t, result, fmt.Sprintf("%s:%s:%d:%d:%s:%s:%s", user.User, passwd, user.Uid, user.Gid, user.Name, user.Dir, user.Shell))
		})
		t.Run("shadow", func(t *testing.T) {
			result, err := container.GetShadow("")
			require.NoError(t, err)
			require.Contains(t, result, fmt.Sprintf("%s:%s:::::::0", user.User, user.Passwd))
		})
	})
}

func UserNotInUserList(t *testing.T, user *types.User, disableShadow bool, container *TestContainer) {
	t.Run("UserNotInUserList", func(t *testing.T) {
		t.Run("passwd", func(t *testing.T) {
			result, err := container.GetPasswd("")
			require.NoError(t, err)
			passwd := user.Passwd
			if !disableShadow {
				passwd = "x"
			}
			require.NotContains(t, result, fmt.Sprintf("%s:%s:%d:%d:%s:%s:%s", user.User, passwd, user.Uid, user.Gid, user.Name, user.Dir, user.Shell))
		})
		t.Run("shadow", func(t *testing.T) {
			result, err := container.GetShadow("")
			require.NoError(t, err)
			require.NotContains(t, result, fmt.Sprintf("%s:%s:::::::0", user.User, user.Passwd))
		})
	})
}

// groupLine renders the "getent group" representation of a group, including
// the comma separated member list.
func groupLine(group *types.Group, disableShadow bool) string {
	passwd := group.Passwd
	if !disableShadow {
		passwd = "x"
	}
	return fmt.Sprintf("%s:%s:%d:%s",
		group.Name, passwd, group.Gid, strings.Join(group.GroupMembers, ","))
}

// gshadowLine renders the "getent gshadow" representation of a group. The
// administrator list is always empty, the member list mirrors the group.
func gshadowLine(group *types.Group) string {
	return fmt.Sprintf("%s:%s::%s",
		group.Name, group.Passwd, strings.Join(group.GroupMembers, ","))
}

func GetSpecificGroup(t *testing.T, group *types.Group, disableShadow bool, container *TestContainer) {
	t.Run("GetSpecificGroup", func(t *testing.T) {
		t.Run("group", func(t *testing.T) {
			t.Run("known group", func(t *testing.T) {
				t.Run("by name", func(t *testing.T) {
					result, err := container.GetGroup(group.Name)
					require.NoError(t, err)
					require.Equal(t, groupLine(group, disableShadow), result)
				})
				t.Run("by gid", func(t *testing.T) {
					result, err := container.GetGroup(strconv.FormatUint(uint64(group.Gid), 10))
					require.NoError(t, err)
					require.Equal(t, groupLine(group, disableShadow), result)
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
					require.Equal(t, gshadowLine(group), result)
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

func GroupInGroupList(t *testing.T, group *types.Group, disableShadow bool, container *TestContainer) {
	t.Run("GroupInGroupList", func(t *testing.T) {
		t.Run("group", func(t *testing.T) {
			result, err := container.GetGroup("")
			require.NoError(t, err)
			require.Contains(t, result, groupLine(group, disableShadow))
		})
		t.Run("gshadow", func(t *testing.T) {
			result, err := container.GetGShadow("")
			require.NoError(t, err)
			require.Contains(t, result, gshadowLine(group))
		})
	})
}

func GroupNotInGroupList(t *testing.T, group *types.Group, disableShadow bool, container *TestContainer) {
	t.Run("GroupNotInGroupList", func(t *testing.T) {
		t.Run("group", func(t *testing.T) {
			result, err := container.GetGroup("")
			require.NoError(t, err)
			require.NotContains(t, result, groupLine(group, disableShadow))
		})
		t.Run("gshadow", func(t *testing.T) {
			result, err := container.GetGShadow("")
			require.NoError(t, err)
			require.NotContains(t, result, gshadowLine(group))
		})
	})
}

// IsUserMemberOfGroup asserts that a user shows up in a group's member list
// and that the group shows up among the user's groups.
//
// primaryGroup is the user's own group (matching user.Gid) and extraGroups are
// the supplementary groups the user is a member of via GroupMembers.
func IsUserMemberOfGroup(t *testing.T, user *types.User, primaryGroup *types.Group, container *TestContainer, extraGroups ...*types.Group) {
	t.Run("IsUserMemberOfGroup", func(t *testing.T) {
		t.Run("groups", func(t *testing.T) {
			result, err := container.Groups(user.User)
			require.NoError(t, err)

			prefix := user.User + " : "
			require.True(t, strings.HasPrefix(result, prefix),
				"expected %q to start with %q", result, prefix)

			// The order of supplementary groups is not guaranteed, so
			// compare as a set.
			names := []string{primaryGroup.Name}
			for _, g := range extraGroups {
				names = append(names, g.Name)
			}
			require.ElementsMatch(t, names,
				strings.Fields(strings.TrimPrefix(result, prefix)))
		})
		t.Run("members", func(t *testing.T) {
			// "members -t" also reports users whose *primary* group is the
			// one being queried, so the expected set is the union of the
			// supplementary members and the primary owner.
			for _, g := range append([]*types.Group{primaryGroup}, extraGroups...) {
				g := g
				t.Run(g.Name, func(t *testing.T) {
					result, err := container.Members(g.Name)
					require.NoError(t, err)
					for _, m := range g.GroupMembers {
						require.Contains(t, result, m,
							"%q must be listed as a member of %q", m, g.Name)
					}
				})
			}
		})
	})
}

// UserHasSupplementaryGroups asserts that "id" reports every expected group
// for the user.
//
// This is the regression guard for gr_mem: when the member vector is not
// written into the NSS result buffer the supplementary groups silently
// disappear from "id" while every other lookup still looks healthy.
func UserHasSupplementaryGroups(t *testing.T, user *types.User, primaryGroup *types.Group, container *TestContainer, extraGroups ...*types.Group) {
	t.Run("UserHasSupplementaryGroups", func(t *testing.T) {
		result, err := container.ID(user.User)
		require.NoError(t, err)

		require.Contains(t, result, fmt.Sprintf("uid=%d(%s)", user.Uid, user.User))
		require.Contains(t, result, fmt.Sprintf("gid=%d(%s)", primaryGroup.Gid, primaryGroup.Name))
		for _, g := range extraGroups {
			require.Contains(t, result, fmt.Sprintf("%d(%s)", g.Gid, g.Name),
				"supplementary group %q must be reported by id", g.Name)
		}
	})
}

// GroupHasMembers asserts the member list of a group as rendered by getent.
func GroupHasMembers(t *testing.T, group *types.Group, disableShadow bool, container *TestContainer) {
	t.Run("GroupHasMembers", func(t *testing.T) {
		t.Run("group", func(t *testing.T) {
			result, err := container.GetGroup(group.Name)
			require.NoError(t, err)
			require.Equal(t, groupLine(group, disableShadow), result)
			for _, m := range group.GroupMembers {
				require.Contains(t, result, m)
			}
		})
		t.Run("gshadow", func(t *testing.T) {
			result, err := container.GetGShadow(group.Name)
			require.NoError(t, err)
			require.Equal(t, gshadowLine(group), result)
		})
	})
}
