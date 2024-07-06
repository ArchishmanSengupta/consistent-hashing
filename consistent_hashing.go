package main

import "errors"

// Custom errors
var (
	ErrNoHost       = errors.New("no host added")
	ErrHostNotFound = errors.New("host not found")
)
