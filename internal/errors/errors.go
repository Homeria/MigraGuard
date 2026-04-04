package errors

import (
	"errors"
	"fmt"
)

// Sentinel errors for comparison using errors.Is()
var (
	ErrTableNotFound     = errors.New("table not found")
	ErrColumnNotFound    = errors.New("column not found")
	ErrInsufficientData  = errors.New("insufficient workload data")
	ErrDatabaseConn      = errors.New("database connection error")
	ErrInvalidSQL        = errors.New("invalid SQL syntax")
	ErrConfiguration     = errors.New("invalid configuration")
)

// MigraError is a custom error struct that provides more context.
type MigraError struct {
	Op      string // Operation where the error occurred
	Table   string // Related table name
	Message string // Detailed message
	Err     error  // Underlying error
}

func (e *MigraError) Error() string {
	if e.Table != "" {
		return fmt.Sprintf("[%s] table '%s': %s: %v", e.Op, e.Table, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s: %v", e.Op, e.Message, e.Err)
}

func (e *MigraError) Unwrap() error {
	return e.Err
}

// Wrap creates a new MigraError.
func Wrap(err error, op string, message string) error {
	return &MigraError{
		Op:      op,
		Message: message,
		Err:     err,
	}
}

// WrapWithTable creates a new MigraError with table context.
func WrapWithTable(err error, op string, table string, message string) error {
	return &MigraError{
		Op:      op,
		Table:   table,
		Message: message,
		Err:     err,
	}
}
