package main

import (
	"fmt"
	"testing"

	"github.com/Eun/nss_http/types"
	"github.com/Eun/nss_http/utils"
	"github.com/stretchr/testify/require"
)

// buf allocates a C buffer like the one glibc passes to getpwnam_r and
// registers its release with the test.
func buf(t *testing.T, size int) *testBuf {
	t.Helper()
	b := newTestBuf(size)
	t.Cleanup(b.free)
	return b
}

func TestWritePointerVector(t *testing.T) {
	t.Run("empty vector is just a NULL terminator", func(t *testing.T) {
		got, addr, err := testWritePointerVector(buf(t, 128), 0)
		require.NoError(t, err)
		require.NotZero(t, addr, "the vector must never be NULL, only empty")
		require.Empty(t, got)
	})

	t.Run("round trips entries in order", func(t *testing.T) {
		got, _, _, err := testCopyThenWriteVector(buf(t, 256), "joe", "alice", "bob")
		require.NoError(t, err)
		require.Equal(t, []string{"joe", "alice", "bob"}, got)
	})

	t.Run("vector is pointer aligned", func(t *testing.T) {
		// usedBytes values that leave the cursor misaligned: storing a
		// pointer at an unaligned address is undefined behaviour and faults
		// on strict alignment architectures.
		for _, used := range []uint64{1, 2, 3, 5, 7, 9, 15} {
			t.Run(fmt.Sprintf("used=%d", used), func(t *testing.T) {
				_, addr, err := testWritePointerVector(buf(t, 256), used)
				require.NoError(t, err)
				require.Zero(t, uint64(addr%pointerVectorSize),
					"vector must start at a pointer aligned address")
				require.GreaterOrEqual(t, uint64(addr), used,
					"vector must not start before the used region")
			})
		}
	})

	t.Run("does not overlap the string data", func(t *testing.T) {
		b := buf(t, 256)
		got, addr, used, err := testCopyThenWriteVector(b, "joe", "alice")
		require.NoError(t, err)

		// The vector must begin at or after the end of the strings, else it
		// would scribble over the very names it points at.
		require.GreaterOrEqual(t, uint64(addr), uint64(b.base())+used)
		// And the strings must still be readable through the vector.
		require.Equal(t, []string{"joe", "alice"}, got)
	})

	t.Run("reports out of memory instead of overflowing", func(t *testing.T) {
		// Room for the strings but not for the trailing pointer vector.
		_, _, _, err := testCopyThenWriteVector(buf(t, 16), "a", "b")
		require.Error(t, err)
		require.ErrorIs(t, err, utils.OutOfMemoryError{})
	})
}

func TestStoreGroupInGroupStruct(t *testing.T) {
	t.Run("group with members", func(t *testing.T) {
		got, err := testStoreGroup(buf(t, 1024), &types.Group{
			Name:         "admins",
			Passwd:       "x",
			Gid:          6000,
			GroupMembers: []string{"joe", "alice"},
		})
		require.NoError(t, err)

		require.Equal(t, "admins", got.Name)
		require.Equal(t, "x", got.Passwd)
		require.Equal(t, uint(6000), got.Gid)
		// The regression: members used to be dropped entirely, so
		// "getent group admins" printed an empty member list.
		require.Equal(t, []string{"joe", "alice"}, got.Members)
	})

	t.Run("group without members", func(t *testing.T) {
		got, err := testStoreGroup(buf(t, 1024), &types.Group{
			Name: "joe", Passwd: "x", Gid: 3000,
		})
		require.NoError(t, err)

		require.Equal(t, "joe", got.Name)
		require.False(t, got.MemNil, "gr_mem must be an empty vector, not NULL")
		require.Empty(t, got.Members)
	})

	t.Run("single member", func(t *testing.T) {
		got, err := testStoreGroup(buf(t, 1024), &types.Group{
			Name: "admins", Passwd: "x", Gid: 6000,
			GroupMembers: []string{"joe"},
		})
		require.NoError(t, err)
		require.Equal(t, []string{"joe"}, got.Members)
	})

	t.Run("many members", func(t *testing.T) {
		members := make([]string, 64)
		for i := range members {
			members[i] = fmt.Sprintf("user%02d", i)
		}
		got, err := testStoreGroup(buf(t, 4096), &types.Group{
			Name: "big", Passwd: "x", Gid: 7000, GroupMembers: members,
		})
		require.NoError(t, err)
		require.Equal(t, members, got.Members)
	})

	t.Run("too small buffer reports an error", func(t *testing.T) {
		_, err := testStoreGroup(buf(t, 24), &types.Group{
			Name: "admins", Passwd: "x", Gid: 6000,
			GroupMembers: []string{"joe", "alice", "bob", "carol"},
		})
		require.Error(t, err)
	})
}

func TestStoreGroupInGShadowStruct(t *testing.T) {
	t.Run("members are exposed and admins are empty", func(t *testing.T) {
		got, err := testStoreGShadow(buf(t, 1024), &types.Group{
			Name:         "admins",
			Passwd:       "!",
			Gid:          6000,
			GroupMembers: []string{"joe", "alice"},
		})
		require.NoError(t, err)

		require.Equal(t, "admins", got.Name)
		require.Equal(t, "!", got.Passwd)
		require.Equal(t, []string{"joe", "alice"}, got.Members)
		// Both vectors must be valid, since callers iterate them
		// unconditionally rather than checking for NULL.
		require.False(t, got.MemNil, "sg_mem must be an empty vector, not NULL")
		require.False(t, got.AdmNil, "sg_adm must be an empty vector, not NULL")
		require.Empty(t, got.Admins)
	})

	t.Run("group without members", func(t *testing.T) {
		got, err := testStoreGShadow(buf(t, 1024), &types.Group{
			Name: "joe", Passwd: "!", Gid: 3000,
		})
		require.NoError(t, err)
		require.Empty(t, got.Members)
		require.Empty(t, got.Admins)
		require.False(t, got.MemNil)
		require.False(t, got.AdmNil)
	})
}

func TestStoreUserInPasswdStruct(t *testing.T) {
	got, err := testStorePasswd(buf(t, 1024), &types.User{
		User:   "joe",
		Passwd: "x",
		Name:   "Joe Doe",
		Dir:    "/home/joe",
		Shell:  "/bin/bash",
		Uid:    3000,
		Gid:    3000,
	})
	require.NoError(t, err)

	require.Equal(t, "joe", got.Name)
	require.Equal(t, "x", got.Passwd)
	require.Equal(t, "Joe Doe", got.Gecos)
	require.Equal(t, "/home/joe", got.Dir)
	require.Equal(t, "/bin/bash", got.Shell)
	require.Equal(t, uint(3000), got.Uid)
	require.Equal(t, uint(3000), got.Gid)
}

func TestStoreUserInSpwdStruct(t *testing.T) {
	got, err := testStoreSpwd(buf(t, 1024), &types.User{
		User:   "joe",
		Passwd: "$6$salt$hash",
	})
	require.NoError(t, err)

	require.Equal(t, "joe", got.Name)
	require.Equal(t, "$6$salt$hash", got.Passwd)
	// -1 tells sshd there is no password aging information.
	require.EqualValues(t, -1, got.LstChg)
	require.EqualValues(t, -1, got.Expire)
}

func TestCopyToBuffer(t *testing.T) {
	t.Run("NULL terminates every string", func(t *testing.T) {
		got, used, err := testCopyToBuffer(buf(t, 64), "joe", "alice")
		require.NoError(t, err)
		require.Equal(t, []string{"joe", "alice"}, got)
		// "joe\0alice\0"
		require.EqualValues(t, 10, used)
	})

	t.Run("reports out of memory", func(t *testing.T) {
		_, _, err := testCopyToBuffer(buf(t, 4), "joe", "alice")
		require.Error(t, err)
		require.ErrorIs(t, err, utils.OutOfMemoryError{})
	})

	t.Run("handles empty strings", func(t *testing.T) {
		got, _, err := testCopyToBuffer(buf(t, 64), "", "joe")
		require.NoError(t, err)
		require.Equal(t, []string{"", "joe"}, got)
	})
}
