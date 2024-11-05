package storage

import (
	"errors"
)

var (
	ErrNotFound = errors.New("url not found")
	ErrURLExist = errors.New("url exist")
)
