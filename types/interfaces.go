package types

import (
	"context"
)

type UIDIdentifier uint
type GIDIdentifier uint

type NameIdentifier string

type Provider interface {
	GetUser(ctx context.Context, identifier any) (*User, error)
	GetUsers(ctx context.Context) ([]User, error)
	GetGroup(ctx context.Context, identifier any) (*Group, error)
	GetGroups(ctx context.Context) ([]Group, error)
	Name() string
	// GetGroup(identifier any) (*Group, error)
}

type CacheProvider interface {
	GetUser(ctx context.Context, identifier any) (*User, error)
	SetUser(ctx context.Context, user *User) error
	GetUsers(ctx context.Context) ([]User, error)
	SetUsers(ctx context.Context, users []User) error
	GetGroup(ctx context.Context, identifier any) (*Group, error)
	SetGroup(ctx context.Context, user *Group) error
	GetGroups(ctx context.Context) ([]Group, error)
	SetGroups(ctx context.Context, groups []Group) error
	Name() string
}
