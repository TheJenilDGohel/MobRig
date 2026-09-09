package database

import "errors"

// ErrNotFound is returned when a requested record does not exist in the database.
var ErrNotFound = errors.New("record not found")
