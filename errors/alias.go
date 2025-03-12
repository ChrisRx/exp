package errors

import (
	"errors"
)

var (
	// NOTE: Use the generic alternative of As instead of the standard library
	// version.
	// As = errors.As
	Is     = errors.Is
	Join   = errors.Join
	New    = errors.New
	Unwrap = errors.Unwrap

	ErrUnsupported = errors.ErrUnsupported
)
