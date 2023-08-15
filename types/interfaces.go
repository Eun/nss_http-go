package types

import "context"

type UIDIdentifier uint
type GIDIdentifier uint

type NameIdentifier string

type Provider interface {
	GetUser(ctx context.Context, identifier any) (*User, error)
	GetGroup(ctx context.Context, identifier any) (*Group, error)
	Name() string
	// GetGroup(identifier any) (*Group, error)
}

type CacheProvider interface {
	GetUser(ctx context.Context, identifier any) (*User, error)
	SetUser(ctx context.Context, user *User) error
	GetGroup(ctx context.Context, identifier any) (*Group, error)
	SetGroup(ctx context.Context, user *Group) error
	Name() string
	// GetGroup(identifier any) (*Group, error)
}
