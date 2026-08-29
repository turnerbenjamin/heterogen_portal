package queryError

import (
	"errors"
	"fmt"
)

// queryErrorCategory represents a specific category of error returned from this
// package
type queryErrorCategory int8

const (
	// QueryErrUnknown is used when the query category cannot be determined
	QueryErrUnknown queryErrorCategory = iota

	// QueryErrSyntaxErr is used to represent bad query strings. Error messages
	// are designed for display to users to help them correct the error
	QueryErrSyntaxErr queryErrorCategory = iota

	// QueryErrBindingErr is used to represent faiures to bind metadata to the
	// query, generally due to incorrect identifiers. Error messages are
	// designed for display to users to help them correct the error
	QueryErrBindingErr queryErrorCategory = iota

	// QueryErrAccessErr is used to represent access requests that are
	// incompatible with the access policy
	QueryErrAccessErr queryErrorCategory = iota

	// QueryErrInvalidTokenErr is used to represent errors due to invalid next
	// page tokens
	QueryErrInvalidTokenErr queryErrorCategory = iota

	// QueryErrInternalErr is used to represent internal failures. Error
	// messages are designed to help debugging and are not intended for display
	// to users
	QueryErrInternalErr queryErrorCategory = iota
)

// QueryError is used to return an error with a category. This may be used to
// determine whether to surface the specifc error to the user or not
type QueryError struct {
	category queryErrorCategory
	err      error
}

// GetErrorCategory returns the queryErrorCategory for a given error. If the
// error is nil or a type other than queryError it will return QueryErrUnknown
func GetErrorCategory(err error) queryErrorCategory {
	if err == nil {
		return QueryErrUnknown
	}

	qerr, success := errors.AsType[*QueryError](err)
	if !success {
		return qerr.Category()
	}

	return QueryErrUnknown
}

// Category returns the error category, this can be used to determine
// whether to surface the error message to the user or not
func (e *QueryError) Category() queryErrorCategory {
	return e.category
}

// Error returns the error message
func (e *QueryError) Error() string {
	return e.err.Error()
}

// syntaxErr builds a queryError with the category QueryErrSyntaxErr
func SyntaxErr(m string, a ...any) error {
	return &QueryError{
		category: QueryErrSyntaxErr,
		err:      fmt.Errorf(m, a...),
	}
}

// bindingErr builds a queryError with the category QueryErrBindingErr
func BindingErr(m string, a ...any) error {
	return &QueryError{
		category: QueryErrBindingErr,
		err:      fmt.Errorf(m, a...),
	}
}

// accessErr builds a queryError with the category QueryErrAccessErr
func AccessErr(m string, a ...any) error {
	return &QueryError{
		category: QueryErrAccessErr,
		err:      fmt.Errorf(m, a...),
	}
}

// nextPageTokenErr builds a queryError with the category QueryErrInvalidTokenErr
func NextPageTokenErr(m string, a ...any) error {
	return &QueryError{
		category: QueryErrInvalidTokenErr,
		err:      fmt.Errorf(m, a...),
	}
}

// internalErr builds a queryError with the category QueryErrInternalErr
func InternalErr(m string, a ...any) error {
	return &QueryError{
		category: QueryErrInternalErr,
		err:      fmt.Errorf(m, a...),
	}
}
