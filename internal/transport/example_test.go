package transport_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/anthropics/claude-agent-sdk-go/internal/transport"
	"github.com/anthropics/claude-agent-sdk-go/internal/types"
)

// ExampleNewSubprocessCLITransport demonstrates how to use the transport layer
func ExampleNewSubprocessCLITransport() {
	// Create options for the Claude agent
	options := types.NewClaudeAgentOptions().
		WithModel("claude-3-haiku-20240307").
		WithMaxTurns(1).
		WithIncludePartialMessages(false).
		WithStderrCallback(func(line string) {
			log.Printf("CLI stderr: %s", line)
		})

	// Create a new transport
	t := transport.NewSubprocessCLITransport("What is 2+2?", options)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to the CLI process
	if err := t.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer func() {
		_ = t.Close(ctx)
	}()

	// Read messages from the transport
	messageChan := t.ReadMessages(ctx)

	// Process messages
	messageCount := 0
	for {
		select {
		case msg, ok := <-messageChan:
			if !ok {
				// Channel closed, transport finished
				fmt.Println("Transport finished")
				return
			}

			messageCount++
			fmt.Printf("Message %d: %s\n", messageCount, msg.Type())

			// Handle different message types
			switch m := msg.(type) {
			case *types.UserMessage:
				fmt.Printf("  User: %v\n", m.Content)

			case *types.AssistantMessage:
				fmt.Printf("  Assistant (%s):\n", m.Model)
				for _, block := range m.Content {
					if textBlock, ok := block.(*types.TextBlock); ok {
						fmt.Printf("    %s\n", textBlock.Text)
					}
				}

			case *types.ResultMessage:
				fmt.Printf("  Result: %s (duration: %dms)\n",
					func() string {
						if m.Result != nil {
							return *m.Result
						}
						return "No result"
					}(), m.DurationMS)

				// End after result message
				return

			case *types.SystemMessage:
				fmt.Printf("  System: %s - %v\n", m.Subtype, m.Data)
			}

		case <-ctx.Done():
			fmt.Println("Context cancelled")
			return

		case <-time.After(10 * time.Second):
			fmt.Println("Timeout waiting for messages")
			return
		}
	}
}

// ExampleNewSubprocessCLITransport_permissions demonstrates transport with tool permissions
func ExampleNewSubprocessCLITransport_permissions() {
	options := types.NewClaudeAgentOptions().
		WithModel("claude-3-haiku-20240307").
		WithAllowedTools("read_file", "write_file").
		WithPermissionMode(types.PermissionModeDefault).
		WithCanUseTool(func(tool string, input map[string]any, ctx interface{}) (types.PermissionResult, error) {
			// Custom permission logic
			if tool == "read_file" {
				return types.PermissionResult{
					Behavior: "allow",
				}, nil
			}

			return types.PermissionResult{
				Behavior: "deny",
				Message:  fmt.Sprintf("Tool %s is not allowed", tool),
			}, nil
		})

	t := transport.NewSubprocessCLITransport(
		"Please read the file README.md and tell me what it contains.",
		options,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := t.Connect(ctx); err != nil {
		log.Printf("Failed to connect: %v", err)
		return
	}
	defer func() {
		_ = t.Close(ctx)
	}()

	// Read and process messages
	for msg := range t.ReadMessages(ctx) {
		fmt.Printf("Received: %s\n", msg.Type())

		// Break after first result message
		if msg.Type() == types.MessageTypeResult {
			break
		}
	}
}

// ExampleNewSubprocessCLITransport_streaming demonstrates streaming transport usage
func ExampleNewSubprocessCLITransport_streaming() {
	options := types.NewClaudeAgentOptions().
		WithModel("claude-3-haiku-20240307").
		WithIncludePartialMessages(true)

	t := transport.NewSubprocessCLITransport("Count from 1 to 5 slowly.", options)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := t.Connect(ctx); err != nil {
		log.Printf("Failed to connect: %v", err)
		return
	}
	defer func() {
		_ = t.Close(ctx)
	}()

	fmt.Println("Starting streaming conversation...")

	for msg := range t.ReadMessages(ctx) {
		switch m := msg.(type) {
		case *types.StreamEvent:
			// Handle streaming events
			fmt.Printf("Stream event: %v\n", m.Event)

		case *types.AssistantMessage:
			// Handle complete assistant messages
			for _, block := range m.Content {
				if textBlock, ok := block.(*types.TextBlock); ok {
					fmt.Printf("Assistant: %s\n", textBlock.Text)
				}
			}

		case *types.ResultMessage:
			fmt.Printf("Conversation completed in %dms\n", m.DurationMS)
			return
		}
	}
}