package redact

import "errors"

// ErrEmptyField is returned when a rule has a blank field name.
var ErrEmptyField = errors.New("redact: rule field name must not be empty")
