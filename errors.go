package memphis

import (
	"errors"
)

// ErrNotDir indicates the proposed location is not a directory
var ErrNotDir = errors.New("ENotDir")

// ErrExists indicates a file already exists at a location
var ErrExists = errors.New("EExists")
