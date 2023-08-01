package types

import "context"

type UIDIdentifier uint

type NameIdentifier string

type Provider interface {
	GetUser(ctx context.Context, identifier any) (*User, error)
	Name() string
	// GetGroup(identifier any) (*Group, error)
}

type CacheProvider interface {
	GetUser(ctx context.Context, identifier any) (*User, error)
	SetUser(ctx context.Context, user *User) error
	Name() string
	// GetGroup(identifier any) (*Group, error)
}
