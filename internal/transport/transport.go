package transport

import (
	"context"

	"github.com/anthropics/claude-agent-sdk-go/internal/types"
)

// Transport defines the interface for Claude communication transports.
//
// WARNING: This internal API is exposed for custom transport implementations
// (e.g., remote Claude Code connections). The Claude Code team may change or
// remove this abstract interface in any future release. Custom implementations
// must be updated to match interface changes.
//
// This is a low-level transport interface that handles raw I/O with the Claude
// process or service. The Query class builds on top of this to implement the
// control protocol and message routing.
type Transport interface {
	// Connect connects the transport and prepares for communication.
	// For subprocess transports, this starts the process.
	// For network transports, this establishes the connection.
	Connect(ctx context.Context) error

	// Close closes the transport connection and cleans up resources.
	Close(ctx context.Context) error

	// Write writes raw data to the transport.
	// Data is typically a JSON string followed by a newline.
	Write(ctx context.Context, data string) error

	// ReadMessages reads and parses messages from the transport.
	// Returns a channel that yields parsed JSON messages.
	ReadMessages(ctx context.Context) <-chan types.Message

	// OnError handles errors from the transport.
	// This allows the transport to communicate errors back to the client.
	OnError(err error)

	// IsReady checks if the transport is ready for communication.
	// Returns true if transport is ready to send/receive messages.
	IsReady() bool

	// EndInput ends the input stream (closes stdin for process transports).
	EndInput(ctx context.Context) error
}
