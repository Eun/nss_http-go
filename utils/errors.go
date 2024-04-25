package utils

type OutOfMemoryError struct{}

func (OutOfMemoryError) Error() string { return "out of memory" }
