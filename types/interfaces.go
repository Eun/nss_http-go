package types

import (
	"context"
)

type UIDIdentifier uint
type GIDIdentifier uint

type NameIdentifier string

type Provider interface {
	GetUser(ctx context.Context, identifier any) (*User, error)
	GetUsers(ctx context.Context) (Users, error)
	GetGroup(ctx context.Context, identifier any) (*Group, error)
	GetGroups(ctx context.Context) (Groups, error)
	Name() string
	// GetGroup(identifier any) (*Item, error)
}

type CacheProvider interface {
	GetUser(ctx context.Context, identifier any) (*User, error)
	SetUser(ctx context.Context, user *User) error
	GetUsers(ctx context.Context) (Users, error)
	SetUsers(ctx context.Context, users Users) error
	GetGroup(ctx context.Context, identifier any) (*Group, error)
	SetGroup(ctx context.Context, user *Group) error
	GetGroups(ctx context.Context) (Groups, error)
	SetGroups(ctx context.Context, groups Groups) error
	Name() string
}
