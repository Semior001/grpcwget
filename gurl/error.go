package gurl

import (
	"errors"
	"fmt"
)

// predefined errors.
var (
	ErrCast       = errors.New("type cannot be casted to the desired")
	ErrNoResponse = errors.New("No response received. Has the request been ever sent?")
)

// ErrMethodNotFound indicates that the requested method was not found
// in service.
type ErrMethodNotFound string

// Error returns the string representation of the error.
func (e ErrMethodNotFound) Error() string {
	return fmt.Sprintf("method %q is not declared in the service", string(e))
}

// ErrOutputNotSupported indicates that handling the output of the GRPC method
// is currently not supported.
type ErrOutputNotSupported string

// Error returns the string representation of the error.
func (e ErrOutputNotSupported) Error() string {
	return fmt.Sprintf("output of type %q is not supported", string(e))
}
