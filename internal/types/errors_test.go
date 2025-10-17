package types

import (
	"errors"
	"testing"
)

func TestCLINotFoundError(t *testing.T) {
	tests := []struct {
		name    string
		message string
		cause   error
		want    string
	}{
		{
			name:    "no cause",
			message: "CLI not found",
			cause:   nil,
			want:    "CLI not found",
		},
		{
			name:    "with cause",
			message: "CLI not found",
			cause:   errors.New("PATH not found"),
			want:    "CLI not found: PATH not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewCLINotFoundError(tt.message, tt.cause)
			if err.Error() != tt.want {
				t.Errorf("CLINotFoundError.Error() = %v, want %v", err.Error(), tt.want)
			}
			if err.Unwrap() != tt.cause {
				t.Errorf("CLINotFoundError.Unwrap() = %v, want %v", err.Unwrap(), tt.cause)
			}
		})
	}
}

func TestCLIConnectionError(t *testing.T) {
	tests := []struct {
		name    string
		message string
		cause   error
		want    string
	}{
		{
			name:    "no cause",
			message: "Connection failed",
			cause:   nil,
			want:    "Connection failed",
		},
		{
			name:    "with cause",
			message: "Connection failed",
			cause:   errors.New("network unreachable"),
			want:    "Connection failed: network unreachable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewCLIConnectionError(tt.message, tt.cause)
			if err.Error() != tt.want {
				t.Errorf("CLIConnectionError.Error() = %v, want %v", err.Error(), tt.want)
			}
			if err.Unwrap() != tt.cause {
				t.Errorf("CLIConnectionError.Unwrap() = %v, want %v", err.Unwrap(), tt.cause)
			}
		})
	}
}

func TestProcessError(t *testing.T) {
	tests := []struct {
		name    string
		message string
		cause   error
		want    string
	}{
		{
			name:    "no cause",
			message: "Process error",
			cause:   nil,
			want:    "Process error",
		},
		{
			name:    "with cause",
			message: "Process error",
			cause:   errors.New("exit status 1"),
			want:    "Process error: exit status 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewProcessError(tt.message, tt.cause)
			if err.Error() != tt.want {
				t.Errorf("ProcessError.Error() = %v, want %v", err.Error(), tt.want)
			}
			if err.Unwrap() != tt.cause {
				t.Errorf("ProcessError.Unwrap() = %v, want %v", err.Unwrap(), tt.cause)
			}
		})
	}
}

func TestJSONDecodeError(t *testing.T) {
	tests := []struct {
		name    string
		message string
		cause   error
		want    string
	}{
		{
			name:    "no cause",
			message: "Invalid JSON",
			cause:   nil,
			want:    "Invalid JSON",
		},
		{
			name:    "with cause",
			message: "Invalid JSON",
			cause:   errors.New("unexpected token"),
			want:    "Invalid JSON: unexpected token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewJSONDecodeError(tt.message, tt.cause)
			if err.Error() != tt.want {
				t.Errorf("JSONDecodeError.Error() = %v, want %v", err.Error(), tt.want)
			}
			if err.Unwrap() != tt.cause {
				t.Errorf("JSONDecodeError.Unwrap() = %v, want %v", err.Unwrap(), tt.cause)
			}
		})
	}
}

func TestMessageParseError(t *testing.T) {
	tests := []struct {
		name    string
		message string
		cause   error
		want    string
	}{
		{
			name:    "no cause",
			message: "Parse error",
			cause:   nil,
			want:    "Parse error",
		},
		{
			name:    "with cause",
			message: "Parse error",
			cause:   errors.New("invalid format"),
			want:    "Parse error: invalid format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewMessageParseError(tt.message, tt.cause)
			if err.Error() != tt.want {
				t.Errorf("MessageParseError.Error() = %v, want %v", err.Error(), tt.want)
			}
			if err.Unwrap() != tt.cause {
				t.Errorf("MessageParseError.Unwrap() = %v, want %v", err.Unwrap(), tt.cause)
			}
		})
	}
}

func TestControlProtocolError(t *testing.T) {
	tests := []struct {
		name    string
		message string
		cause   error
		want    string
	}{
		{
			name:    "no cause",
			message: "Protocol error",
			cause:   nil,
			want:    "Protocol error",
		},
		{
			name:    "with cause",
			message: "Protocol error",
			cause:   errors.New("invalid response"),
			want:    "Protocol error: invalid response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewControlProtocolError(tt.message, tt.cause)
			if err.Error() != tt.want {
				t.Errorf("ControlProtocolError.Error() = %v, want %v", err.Error(), tt.want)
			}
			if err.Unwrap() != tt.cause {
				t.Errorf("ControlProtocolError.Unwrap() = %v, want %v", err.Unwrap(), tt.cause)
			}
		})
	}
}

func TestPermissionDeniedError(t *testing.T) {
	tests := []struct {
		name    string
		message string
		cause   error
		want    string
	}{
		{
			name:    "no cause",
			message: "Permission denied",
			cause:   nil,
			want:    "Permission denied",
		},
		{
			name:    "with cause",
			message: "Permission denied",
			cause:   errors.New("access forbidden"),
			want:    "Permission denied: access forbidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewPermissionDeniedError(tt.message, tt.cause)
			if err.Error() != tt.want {
				t.Errorf("PermissionDeniedError.Error() = %v, want %v", err.Error(), tt.want)
			}
			if err.Unwrap() != tt.cause {
				t.Errorf("PermissionDeniedError.Unwrap() = %v, want %v", err.Unwrap(), tt.cause)
			}
		})
	}
}

func TestErrorTypes(t *testing.T) {
	// Test that all error types implement the error interface
	var _ error = &CLINotFoundError{}
	var _ error = &CLIConnectionError{}
	var _ error = &ProcessError{}
	var _ error = &JSONDecodeError{}
	var _ error = &MessageParseError{}
	var _ error = &ControlProtocolError{}
	var _ error = &PermissionDeniedError{}
}
