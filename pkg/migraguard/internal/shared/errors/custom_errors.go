package errors

import (
	"fmt"
)

// ErrorCode is a unique identifier for specific error types.
type ErrorCode string

const (
	ErrCodeInternal    ErrorCode = "MG-INT-001"
	ErrCodeDBConn      ErrorCode = "MG-DB-001"
	ErrCodeTableMissing ErrorCode = "MG-DB-002"
	ErrCodeSQLInvalid  ErrorCode = "MG-SQL-001"
	ErrCodeAnalysis    ErrorCode = "MG-ANL-001"
)

// MigraError is a structured error with a unique code and contextual information.
type MigraError struct {
	Code    ErrorCode
	Op      string
	Message string
	Err     error
}

func (e *MigraError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %s (Inner: %v)", e.Code, e.Op, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Op, e.Message)
}

func (e *MigraError) Unwrap() error {
	return e.Err
}

// Global Error Instances
var (
	ErrDatabaseConn   = &MigraError{Code: ErrCodeDBConn, Message: "database connection failed"}
	ErrTableNotFound  = &MigraError{Code: ErrCodeTableMissing, Message: "target table not found"}
	ErrInvalidSQL     = &MigraError{Code: ErrCodeSQLInvalid, Message: "invalid SQL syntax"}
	ErrAnalysis       = &MigraError{Code: ErrCodeAnalysis, Message: "error occurred during analysis engine execution"}
)

// Wrap adds operation and message context to an error while assigning an appropriate code.
func Wrap(err error, op string, message string) error {
	if err == nil {
		return nil
	}

	// If it's already a MigraError, preserve its code
	if me, ok := err.(*MigraError); ok {
		return &MigraError{
			Code:    me.Code,
			Op:      op,
			Message: message,
			Err:     err,
		}
	}

	return &MigraError{
		Code:    ErrCodeInternal,
		Op:      op,
		Message: message,
		Err:     err,
	}
}

// WrapWithTable adds context information including a table name to an error.
func WrapWithTable(err error, op string, table string, message string) error {
	if err == nil {
		return nil
	}
	return Wrap(err, op, fmt.Sprintf("table(%s) - %s", table, message))
}

// New creates a new MigraError from scratch.
func New(code ErrorCode, op string, message string) error {
	return &MigraError{
		Code:    code,
		Op:      op,
		Message: message,
	}
}
