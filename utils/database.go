package utils

import (
	"strings"
	"sync"

	"github.com/pkg/errors"
)

type DatabaseUnavailableError struct {
	underlyingError error
}

func (e *DatabaseUnavailableError) Error() string {
	var sb strings.Builder
	sb.WriteString("database is not available")
	if e.underlyingError == nil {
		return sb.String()
	}
	sb.WriteString(": ")
	sb.WriteString(e.underlyingError.Error())
	return sb.String()
}

func (e *DatabaseUnavailableError) Is(err error) bool {
	_, ok := err.(*DatabaseUnavailableError)
	return ok
}

func (e *DatabaseUnavailableError) Unwrap() error {
	return e.underlyingError
}

type DatabaseOutOfMemoryError struct{}

func (e *DatabaseOutOfMemoryError) Error() string {
	return "out of memory: try again with a bigger buffer"
}

func (e *DatabaseOutOfMemoryError) Is(err error) bool {
	_, ok := err.(*DatabaseOutOfMemoryError)
	return ok
}

type DatabaseEOF struct{}

func (d *DatabaseEOF) Error() string { return "eof" }

func (e *DatabaseEOF) Is(err error) bool {
	_, ok := err.(*DatabaseEOF)
	return ok
}

type Database[T any] struct {
	mu           sync.Mutex
	isOpen       bool
	fetchedItems bool
	items        []T
	itemIndex    int
}

func (db *Database[T]) Open() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.isOpen {
		return &DatabaseUnavailableError{errors.New("database is already open")}
	}
	db.isOpen = true
	return nil
}

func (db *Database[T]) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if !db.isOpen {
		return &DatabaseUnavailableError{errors.New("database is not open")}
	}
	db.isOpen = false
	db.fetchedItems = false
	db.items = nil
	db.itemIndex = 0
	return nil
}

func (db *Database[T]) GetEnt(getItemsFunc func() ([]T, error), storeItemFunc func(item *T) error) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if !db.isOpen {
		return &DatabaseUnavailableError{errors.New("database is not open")}
	}
	if !db.fetchedItems {
		var err error
		db.items, err = getItemsFunc()
		if err != nil {
			return &DatabaseUnavailableError{errors.Wrap(err, "unable to get items")}
		}
		db.fetchedItems = true
	}

	if db.itemIndex+1 > len(db.items) {
		return &DatabaseEOF{}
	}

	// store in buffer
	if err := storeItemFunc(&db.items[db.itemIndex]); err != nil {
		if errors.Is(err, OutOfMemoryError{}) {
			return &DatabaseOutOfMemoryError{}
		}
		return &DatabaseUnavailableError{errors.Wrap(err, "unable to store items")}
	}
	db.itemIndex++
	return nil
}
