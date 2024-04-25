package utils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatabase_Open(t *testing.T) {
	t.Run("single open", func(t *testing.T) {
		var db Database[string]
		require.NoError(t, db.Open())
	})
	t.Run("double open", func(t *testing.T) {
		var db Database[string]
		require.NoError(t, db.Open())
		require.EqualError(t, db.Open(), "database is not available: database is already open")
	})
}

func TestDatabase_Close(t *testing.T) {
	t.Run("normal close", func(t *testing.T) {
		var db Database[string]
		require.NoError(t, db.Open())
		require.NoError(t, db.Close())
	})
	t.Run("double close", func(t *testing.T) {
		var db Database[string]
		require.NoError(t, db.Open())
		require.NoError(t, db.Close())
		require.EqualError(t, db.Close(), "database is not available: database is not open")
	})
}

func TestDatabase_GetEnt(t *testing.T) {
	t.Run("database is not open", func(t *testing.T) {
		var db Database[string]
		require.EqualError(t, db.GetEnt(func() ([]string, error) {
			return nil, nil
		}, func(*string) error {
			return nil
		}), "database is not available: database is not open")
	})
	t.Run("error during fetching items", func(t *testing.T) {
		var db Database[string]
		require.NoError(t, db.Open())
		require.EqualError(t, db.GetEnt(func() ([]string, error) {
			return nil, errors.New("some error")
		}, func(*string) error {
			return nil
		}), "database is not available: unable to get items: some error")
	})
	t.Run("fetched all items", func(t *testing.T) {
		var db Database[string]
		require.NoError(t, db.Open())
		require.EqualError(t, db.GetEnt(func() ([]string, error) {
			return nil, nil
		}, func(*string) error {
			return nil
		}), "eof")
	})
	t.Run("storing items failed", func(t *testing.T) {
		t.Run("generic error", func(t *testing.T) {
			var db Database[string]
			require.NoError(t, db.Open())
			require.EqualError(t, db.GetEnt(func() ([]string, error) {
				return []string{"joe"}, nil
			}, func(*string) error {
				return errors.New("some error")
			}), "database is not available: unable to store items: some error")
		})
		t.Run("oom", func(t *testing.T) {
			var db Database[string]
			require.NoError(t, db.Open())
			require.EqualError(t, db.GetEnt(func() ([]string, error) {
				return []string{"joe"}, nil
			}, func(*string) error {
				return OutOfMemoryError{}
			}), "out of memory: try again with a bigger buffer")
		})
	})

	t.Run("success", func(t *testing.T) {
		var db Database[string]
		var storedItems []string
		storeFunc := func(s *string) error {
			storedItems = append(storedItems, *s)
			return nil
		}
		require.NoError(t, db.Open())
		require.NoError(t, db.GetEnt(func() ([]string, error) {
			return []string{"joe", "alice"}, nil
		}, storeFunc))
		require.NoError(t, db.GetEnt(func() ([]string, error) {
			return nil, errors.New("should never be called")
		}, storeFunc))
		require.EqualError(t, db.GetEnt(func() ([]string, error) {
			return nil, errors.New("should never be called")
		}, storeFunc), "eof")
		require.Equal(t, []string{"joe", "alice"}, storedItems)

	})

}
