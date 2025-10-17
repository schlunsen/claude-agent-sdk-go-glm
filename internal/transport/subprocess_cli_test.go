package transport

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/anthropics/claude-agent-sdk-go/internal/types"
)

func TestFindCLI(t *testing.T) {
	// Test finding CLI in PATH
	cliPath := findCLI()
	if cliPath == "" {
		t.Skip("Claude CLI not found, skipping test")
	}

	// Verify the path exists
	if _, err := os.Stat(cliPath); os.IsNotExist(err) {
		t.Errorf("CLI path %s does not exist", cliPath)
	}
}

func TestSubprocessCLITransport_Creation(t *testing.T) {
	options := types.NewClaudeAgentOptions().
		WithModel("claude-3-sonnet-20240229").
		WithMaxTurns(10)

	transport := NewSubprocessCLITransport("test prompt", options)

	if transport.prompt != "test prompt" {
		t.Errorf("Expected prompt 'test prompt', got '%s'", transport.prompt)
	}

	if transport.options.Model == nil || *transport.options.Model != "claude-3-sonnet-20240229" {
		t.Errorf("Expected model 'claude-3-sonnet-20240229', got '%v'", transport.options.Model)
	}

	if !transport.isStreaming {
		t.Error("Expected streaming mode to be true")
	}

	if transport.maxBufferSize != DefaultMaxBufferSize {
		t.Errorf("Expected default buffer size %d, got %d", DefaultMaxBufferSize, transport.maxBufferSize)
	}
}

func TestSubprocessCLITransport_BuildCommand(t *testing.T) {
	options := types.NewClaudeAgentOptions().
		WithModel("claude-3-sonnet-20240229").
		WithAllowedTools("tool1", "tool2").
		WithDisallowedTools("bad_tool").
		WithMaxTurns(5).
		WithPermissionMode(types.PermissionModeDefault).
		WithContinueConversation(true).
		WithCWD("/tmp").
		WithExtraArg("debug-to-stderr", nil)

	transport := NewSubprocessCLITransport("test prompt", options)
	transport.cliPath = "/path/to/claude"

	cmd := transport.buildCommand()

	// Check basic arguments
	if cmd[0] != "/path/to/claude" {
		t.Errorf("Expected CLI path '/path/to/claude', got '%s'", cmd[0])
	}

	// Check for required flags
	cmdStr := strings.Join(cmd, " ")
	requiredFlags := []string{
		"--output-format", "stream-json",
		"--verbose",
		"--model", "claude-3-sonnet-20240229",
		"--allowedTools", "tool1,tool2",
		"--disallowedTools", "bad_tool",
		"--max-turns", "5",
		"--permission-mode", "default",
		"--continue",
		"--input-format", "stream-json",
	}

	for _, flag := range requiredFlags {
		if !strings.Contains(cmdStr, flag) {
			t.Errorf("Command should contain flag '%s', got: %s", flag, cmdStr)
		}
	}
}

func TestSubprocessCLITransport_BuildCommand_WithSystemPrompt(t *testing.T) {
	// Test with string system prompt
	options1 := types.NewClaudeAgentOptions().WithSystemPrompt("You are a helpful assistant")
	transport1 := NewSubprocessCLITransport("test", options1)
	transport1.cliPath = "claude"
	cmd1 := transport1.buildCommand()

	cmd1Str := strings.Join(cmd1, " ")
	if !strings.Contains(cmd1Str, "--system-prompt") {
		t.Error("Command should contain --system-prompt flag")
	}
	if !strings.Contains(cmd1Str, "You are a helpful assistant") {
		t.Error("Command should contain system prompt text")
	}

	// Test with preset system prompt
	preset := map[string]interface{}{
		"type":   "preset",
		"append": "Additional instructions",
	}
	options2 := types.NewClaudeAgentOptions().WithSystemPrompt(preset)
	transport2 := NewSubprocessCLITransport("test", options2)
	transport2.cliPath = "claude"
	cmd2 := transport2.buildCommand()

	cmd2Str := strings.Join(cmd2, " ")
	if !strings.Contains(cmd2Str, "--append-system-prompt") {
		t.Error("Command should contain --append-system-prompt flag")
	}
	if !strings.Contains(cmd2Str, "Additional instructions") {
		t.Error("Command should contain preset append text")
	}
}

func TestSubprocessCLITransport_BuildCommand_WithMCPServers(t *testing.T) {
	mcpConfig := types.MCPServerConfig{
		Type:    "command",
		Command: "node",
		Args:    []string{"server.js"},
	}

	options := types.NewClaudeAgentOptions().WithMCPServer("test-server", &mcpConfig)
	transport := NewSubprocessCLITransport("test", options)
	transport.cliPath = "claude"
	cmd := transport.buildCommand()

	cmdStr := strings.Join(cmd, " ")
	if !strings.Contains(cmdStr, "--mcp-config") {
		t.Error("Command should contain --mcp-config flag")
	}
}

func TestSubprocessCLITransport_BuildCommand_WithAgents(t *testing.T) {
	agent := types.AgentDefinition{
		Description: "Test agent",
		Prompt:      "You are a test agent",
		Tools:       []string{"tool1"},
	}

	options := types.NewClaudeAgentOptions().WithAgent("test-agent", agent)
	transport := NewSubprocessCLITransport("test", options)
	transport.cliPath = "claude"
	cmd := transport.buildCommand()

	cmdStr := strings.Join(cmd, " ")
	if !strings.Contains(cmdStr, "--agents") {
		t.Error("Command should contain --agents flag")
	}
}

func TestSubprocessCLITransport_CompareVersions(t *testing.T) {
	transport := &SubprocessCLITransport{}

	testCases := []struct {
		v1     string
		v2     string
		expect int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.1", "1.0.0", 1},
		{"1.0.0", "1.0.1", -1},
		{"2.0.0", "1.9.9", 1},
		{"1.9.9", "2.0.0", -1},
		{"1.2.0", "1.2.0", 0},
		{"1.2.3", "1.2.10", -1},
		{"1.2.10", "1.2.3", 1},
	}

	for _, tc := range testCases {
		result := transport.compareVersions(tc.v1, tc.v2)
		if result != tc.expect {
			t.Errorf("compareVersions(%s, %s) = %d, expected %d", tc.v1, tc.v2, result, tc.expect)
		}
	}
}

func TestSubprocessCLITransport_Connect_InvalidCLI(t *testing.T) {
	options := types.NewClaudeAgentOptions()
	options = options.WithCLIPath("/nonexistent/path/to/claude")

	transport := NewSubprocessCLITransport("test", options)

	err := transport.Connect(context.Background())
	if err == nil {
		t.Error("Expected error when connecting with invalid CLI path")
	}

	var cliErr *types.CLINotFoundError
	if !errors.As(err, &cliErr) {
		t.Errorf("Expected CLINotFoundError, got %T", err)
	}
}

func TestSubprocessCLITransport_Write_NotReady(t *testing.T) {
	options := types.NewClaudeAgentOptions()
	transport := NewSubprocessCLITransport("test", options)

	// Don't connect, should not be ready
	err := transport.Write(context.Background(), "test message")
	if err == nil {
		t.Error("Expected error when writing to not-ready transport")
	}

	var connErr *types.CLIConnectionError
	if !errors.As(err, &connErr) {
		t.Errorf("Expected CLIConnectionError, got %T", err)
	}
}

func TestSubprocessCLITransport_ParseMessage(t *testing.T) {
	transport := &SubprocessCLITransport{}

	// Test valid user message
	userMsg := map[string]interface{}{
		"type":    "user",
		"content": "Hello, world!",
	}

	msg, err := transport.parseMessage(userMsg)
	if err != nil {
		t.Fatalf("Failed to parse user message: %v", err)
	}

	userMsgTyped, ok := msg.(*types.UserMessage)
	if !ok {
		t.Fatalf("Expected UserMessage, got %T", msg)
	}

	if userMsgTyped.Type() != types.MessageTypeUser {
		t.Errorf("Expected message type '%s', got '%s'", types.MessageTypeUser, userMsgTyped.Type())
	}

	if userMsgTyped.Content != "Hello, world!" {
		t.Errorf("Expected content 'Hello, world!', got '%v'", userMsgTyped.Content)
	}

	// Test valid assistant message
	assistantMsg := map[string]interface{}{
		"type": "assistant",
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": "Hello! How can I help you?",
			},
		},
		"model": "claude-3-sonnet-20240229",
	}

	msg, err = transport.parseMessage(assistantMsg)
	if err != nil {
		t.Fatalf("Failed to parse assistant message: %v", err)
	}

	assistantMsgTyped, ok := msg.(*types.AssistantMessage)
	if !ok {
		t.Fatalf("Expected AssistantMessage, got %T", msg)
	}

	if assistantMsgTyped.Type() != types.MessageTypeAssistant {
		t.Errorf("Expected message type '%s', got '%s'", types.MessageTypeAssistant, assistantMsgTyped.Type())
	}

	if assistantMsgTyped.Model != "claude-3-sonnet-20240229" {
		t.Errorf("Expected model 'claude-3-sonnet-20240229', got '%s'", assistantMsgTyped.Model)
	}

	if len(assistantMsgTyped.Content) != 1 {
		t.Fatalf("Expected 1 content block, got %d", len(assistantMsgTyped.Content))
	}

	textBlock, ok := assistantMsgTyped.Content[0].(*types.TextBlock)
	if !ok {
		t.Fatalf("Expected TextBlock, got %T", assistantMsgTyped.Content[0])
	}

	if textBlock.Text != "Hello! How can I help you?" {
		t.Errorf("Expected text 'Hello! How can I help you?', got '%s'", textBlock.Text)
	}
}

func TestSubprocessCLITransport_ParseMessage_Invalid(t *testing.T) {
	transport := &SubprocessCLITransport{}

	// Test invalid message type
	invalidMsg := map[string]interface{}{
		"type":    "invalid_type",
		"content": "test",
	}

	_, err := transport.parseMessage(invalidMsg)
	if err == nil {
		t.Error("Expected error when parsing invalid message type")
	}

	var parseErr *types.MessageParseError
	if !errors.As(err, &parseErr) {
		t.Errorf("Expected MessageParseError, got %T", err)
	}
}

func TestSubprocessCLITransport_EndInput(t *testing.T) {
	options := types.NewClaudeAgentOptions()
	transport := NewSubprocessCLITransport("test", options)

	// End input without connecting should not error
	err := transport.EndInput(context.Background())
	if err != nil {
		t.Errorf("Unexpected error ending input: %v", err)
	}
}

func TestSubprocessCLITransport_Close(t *testing.T) {
	options := types.NewClaudeAgentOptions()
	transport := NewSubprocessCLITransport("test", options)

	// Close without connecting should not error
	err := transport.Close(context.Background())
	if err != nil {
		t.Errorf("Unexpected error closing transport: %v", err)
	}

	if transport.IsReady() {
		t.Error("Transport should not be ready after close")
	}
}

func TestSubprocessCLITransport_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if Claude CLI is available
	cliPath := findCLI()
	if cliPath == "" {
		t.Skip("Claude CLI not found, skipping integration test")
	}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "claude-sdk-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	options := types.NewClaudeAgentOptions().
		WithModel("claude-3-haiku-20240307").
		WithMaxTurns(1).
		WithCWD(tempDir).
		WithIncludePartialMessages(false)

	transport := NewSubprocessCLITransport("What is 2+2?", options)

	// Connect
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer func() {
		_ = transport.Close(ctx)
	}()

	if !transport.IsReady() {
		t.Error("Transport should be ready after connect")
	}

	// Read messages
	messageChan := transport.ReadMessages(ctx)

	// Wait for messages with timeout
	messageCount := 0
	messages := make([]types.Message, 0)

	for {
		select {
		case msg, ok := <-messageChan:
			if !ok {
				// Channel closed
				goto done
			}
			messages = append(messages, msg)
			messageCount++

			// Stop after a reasonable number of messages
			if messageCount >= 10 {
				goto done
			}

		case <-time.After(10 * time.Second):
			t.Error("Timeout waiting for messages")
			goto done

		case <-ctx.Done():
			t.Error("Context cancelled while waiting for messages")
			goto done
		}
	}

done:
	if messageCount == 0 {
		t.Error("Expected at least one message")
	}

	// Check that we got expected message types
	hasResult := false
	for _, msg := range messages {
		if msg.Type() == types.MessageTypeResult {
			hasResult = true
			break
		}
	}

	if !hasResult {
		t.Error("Expected at least one result message")
	}
}

// Mock CLI process for testing
func createMockCLI(t *testing.T, script string) string {
	tempDir, err := os.MkdirTemp("", "claude-mock-cli-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	scriptPath := filepath.Join(tempDir, "mock-cli.sh")
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("Failed to write mock script: %v", err)
	}

	return scriptPath
}

func TestSubprocessCLITransport_MockCLI(t *testing.T) {
	// Create a mock CLI that outputs simple JSON messages
	mockScript := `#!/bin/bash
echo '{"type":"system","subtype":"start","data":{"session":"test"}}'
echo '{"type":"assistant","content":[{"type":"text","text":"Hello!"}],"model":"claude-3-haiku-20240307"}'
echo '{"type":"result","subtype":"success","duration_ms":1000,"session_id":"test","result":"Complete"}'
`

	cliPath := createMockCLI(t, mockScript)
	defer func() {
		_ = os.RemoveAll(filepath.Dir(cliPath))
	}()

	options := types.NewClaudeAgentOptions().
		WithCWD(filepath.Dir(cliPath))

	transport := NewSubprocessCLITransport("test", options)
	transport.cliPath = cliPath

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to mock CLI: %v", err)
	}
	defer func() {
		_ = transport.Close(ctx)
	}()

	// Read messages
	messageChan := transport.ReadMessages(ctx)
	messageCount := 0
	expectedTypes := []string{
		types.MessageTypeSystem,
		types.MessageTypeAssistant,
		types.MessageTypeResult,
	}

	for messageCount < len(expectedTypes) {
		select {
		case msg, ok := <-messageChan:
			if !ok {
				t.Errorf("Channel closed unexpectedly after %d messages", messageCount)
				return
			}

			expectedType := expectedTypes[messageCount]
			if msg.Type() != expectedType {
				t.Errorf("Message %d: expected type '%s', got '%s'", messageCount, expectedType, msg.Type())
			}

			messageCount++

		case <-time.After(2 * time.Second):
			t.Errorf("Timeout waiting for message %d", messageCount)
			return

		case <-ctx.Done():
			t.Error("Context cancelled")
			return
		}
	}

	if messageCount != len(expectedTypes) {
		t.Errorf("Expected %d messages, got %d", len(expectedTypes), messageCount)
	}
}