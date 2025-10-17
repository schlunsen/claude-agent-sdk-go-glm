package types

import (
	"fmt"
)

// CLINotFoundError is returned when the Claude Code CLI cannot be found
type CLINotFoundError struct {
	Message string
	Cause   error
}

func (e *CLINotFoundError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *CLINotFoundError) Unwrap() error {
	return e.Cause
}

// NewCLINotFoundError creates a new CLINotFoundError
func NewCLINotFoundError(message string, cause error) *CLINotFoundError {
	return &CLINotFoundError{
		Message: message,
		Cause:   cause,
	}
}

// CLIConnectionError is returned when there's an error connecting to the CLI
type CLIConnectionError struct {
	Message string
	Cause   error
}

func (e *CLIConnectionError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *CLIConnectionError) Unwrap() error {
	return e.Cause
}

// NewCLIConnectionError creates a new CLIConnectionError
func NewCLIConnectionError(message string, cause error) *CLIConnectionError {
	return &CLIConnectionError{
		Message: message,
		Cause:   cause,
	}
}

// ProcessError is returned when there's an error with the subprocess
type ProcessError struct {
	Message string
	Cause   error
}

func (e *ProcessError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *ProcessError) Unwrap() error {
	return e.Cause
}

// NewProcessError creates a new ProcessError
func NewProcessError(message string, cause error) *ProcessError {
	return &ProcessError{
		Message: message,
		Cause:   cause,
	}
}

// JSONDecodeError is returned when JSON decoding fails
type JSONDecodeError struct {
	Message string
	Cause   error
}

func (e *JSONDecodeError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *JSONDecodeError) Unwrap() error {
	return e.Cause
}

// NewJSONDecodeError creates a new JSONDecodeError
func NewJSONDecodeError(message string, cause error) *JSONDecodeError {
	return &JSONDecodeError{
		Message: message,
		Cause:   cause,
	}
}

// MessageParseError is returned when message parsing fails
type MessageParseError struct {
	Message string
	Cause   error
}

func (e *MessageParseError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *MessageParseError) Unwrap() error {
	return e.Cause
}

// NewMessageParseError creates a new MessageParseError
func NewMessageParseError(message string, cause error) *MessageParseError {
	return &MessageParseError{
		Message: message,
		Cause:   cause,
	}
}

// ControlProtocolError is returned when there's a control protocol error
type ControlProtocolError struct {
	Message string
	Cause   error
}

func (e *ControlProtocolError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *ControlProtocolError) Unwrap() error {
	return e.Cause
}

// NewControlProtocolError creates a new ControlProtocolError
func NewControlProtocolError(message string, cause error) *ControlProtocolError {
	return &ControlProtocolError{
		Message: message,
		Cause:   cause,
	}
}

// PermissionDeniedError is returned when permission is denied
type PermissionDeniedError struct {
	Message string
	Cause   error
}

func (e *PermissionDeniedError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *PermissionDeniedError) Unwrap() error {
	return e.Cause
}

// NewPermissionDeniedError creates a new PermissionDeniedError
func NewPermissionDeniedError(message string, cause error) *PermissionDeniedError {
	return &PermissionDeniedError{
		Message: message,
		Cause:   cause,
	}
}
