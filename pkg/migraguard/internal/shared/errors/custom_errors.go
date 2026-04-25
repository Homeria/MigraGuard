package errors

import (
	"errors"
	"fmt"
)

// Common errors used across the project.
var (
	ErrDatabaseConn   = errors.New("database connection failed")
	ErrTableNotFound  = errors.New("target table not found")
	ErrColumnNotFound = errors.New("target column not found")
	ErrInvalidSQL     = errors.New("invalid SQL syntax")
	ErrAnalysis       = errors.New("error occurred during analysis engine execution")
)

// Wrap adds context information to an error.
func Wrap(err error, op string, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("[%s] %s: %w", op, message, err)
}

// WrapWithTable adds context information including a table name to an error.
func WrapWithTable(err error, op string, table string, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("[%s] table(%s) - %s: %w", op, table, message, err)
}
