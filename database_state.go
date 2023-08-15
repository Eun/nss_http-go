package main

type DatabaseState[T any] struct {
	IsOpen       bool
	FetchedItems bool
	Items        []T
	ItemIndex    int
}
