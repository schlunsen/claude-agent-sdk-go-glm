package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/anthropics/claude-agent-sdk-go/internal/types"
)

const (
	// DefaultMaxBufferSize is the default maximum buffer size for JSON messages
	DefaultMaxBufferSize = 1024 * 1024 // 1MB

	// MinimumClaudeCodeVersion is the minimum supported Claude Code CLI version
	MinimumClaudeCodeVersion = "2.0.0"

	// ClaudeAgentSDKVersion is the version of this SDK
	ClaudeAgentSDKVersion = "0.1.0"

	// CLICodeEntrypoint is the entrypoint identifier for the CLI
	CLICodeEntrypoint = "sdk-go"
)

// SubprocessCLITransport implements Transport using Claude Code CLI subprocess
type SubprocessCLITransport struct {
	// Configuration
	prompt             string                    // The prompt to send
	options            *types.ClaudeAgentOptions // Transport options
	isStreaming        bool                      // Whether we're in streaming mode
	cliPath            string                    // Path to Claude CLI
	cwd                string                    // Working directory
	maxBufferSize      int                       // Maximum buffer size

	// Process management
	cmd                *exec.Cmd          // The subprocess command
	ctx                context.Context    // Context for cancellation
	cancel             context.CancelFunc // Cancellation function
	stdin              io.WriteCloser     // stdin pipe
	stdout             io.ReadCloser      // stdout pipe
	stderr             io.ReadCloser      // stderr pipe

	// Stream management
	stdoutReader       *bufio.Scanner     // Buffered stdout reader
	stdinWriter        *bufio.Writer      // Buffered stdin writer

	// State
	ready              bool               // Whether transport is ready
	mu                 sync.RWMutex       // Mutex for thread safety
	exitError          error              // Error that caused process exit

	// Message handling
	messageChan        chan types.Message // Channel for outgoing messages
	errorChan          chan error         // Channel for errors

	// Stderr handling
	stderrCallback     func(string)       // Callback for stderr output
	stderrDone         chan struct{}      // Channel to signal stderr handling done
}

// NewSubprocessCLITransport creates a new SubprocessCLITransport
func NewSubprocessCLITransport(prompt string, options *types.ClaudeAgentOptions) *SubprocessCLITransport {
	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Determine if we're in streaming mode (prompt could be empty for interactive sessions)
	isStreaming := true // Default to streaming for Go implementation

	// Find CLI path if not provided
	cliPath := ""
	if options.CLIPath != nil {
		cliPath = *options.CLIPath
	} else {
		cliPath = findCLI()
	}

	// Set working directory
	cwd := ""
	if options.CWD != nil {
		cwd = *options.CWD
	} else {
		cwd, _ = os.Getwd()
	}

	// Set max buffer size
	maxBufferSize := DefaultMaxBufferSize
	if options.MaxBufferSize != nil {
		maxBufferSize = *options.MaxBufferSize
	}

	return &SubprocessCLITransport{
		prompt:         prompt,
		options:        options,
		isStreaming:    isStreaming,
		cliPath:        cliPath,
		cwd:            cwd,
		maxBufferSize:  maxBufferSize,
		ctx:            ctx,
		cancel:         cancel,
		messageChan:    make(chan types.Message, 100), // Buffered channel
		errorChan:      make(chan error, 10),          // Buffered channel for errors
		stderrCallback: options.StderrCallback,
		stderrDone:     make(chan struct{}),
	}
}

// findCLI finds the Claude Code CLI binary in common locations
func findCLI() string {
	// First check PATH
	if cli, err := exec.LookPath("claude"); err == nil {
		return cli
	}

	// Check common installation locations
	homeDir, _ := os.UserHomeDir()
	locations := []string{
		filepath.Join(homeDir, ".npm-global", "bin", "claude"),
		"/usr/local/bin/claude",
		filepath.Join(homeDir, ".local", "bin", "claude"),
		filepath.Join(homeDir, "node_modules", ".bin", "claude"),
		filepath.Join(homeDir, ".yarn", "bin", "claude"),
	}

	for _, path := range locations {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// If not found, this will be caught during connect()
	return "claude" // Default to "claude" to trigger proper error during connect
}

// buildCommand builds the CLI command with appropriate arguments
func (t *SubprocessCLITransport) buildCommand() []string {
	cmd := []string{t.cliPath, "--output-format", "stream-json", "--verbose"}

	// System prompt handling
	if t.options.SystemPrompt != nil {
		switch prompt := t.options.SystemPrompt.(type) {
		case string:
			cmd = append(cmd, "--system-prompt", prompt)
		case map[string]interface{}:
			if promptType, ok := prompt["type"].(string); ok && promptType == "preset" {
				if appendText, ok := prompt["append"].(string); ok {
					cmd = append(cmd, "--append-system-prompt", appendText)
				}
			}
		}
	}

	// Allowed tools
	if len(t.options.AllowedTools) > 0 {
		cmd = append(cmd, "--allowedTools", strings.Join(t.options.AllowedTools, ","))
	}

	// Disallowed tools
	if len(t.options.DisallowedTools) > 0 {
		cmd = append(cmd, "--disallowedTools", strings.Join(t.options.DisallowedTools, ","))
	}

	// Max turns
	if t.options.MaxTurns != nil {
		cmd = append(cmd, "--max-turns", strconv.Itoa(*t.options.MaxTurns))
	}

	// Model
	if t.options.Model != nil {
		cmd = append(cmd, "--model", *t.options.Model)
	}

	// Permission prompt tool name
	if t.options.PermissionPromptToolName != nil {
		cmd = append(cmd, "--permission-prompt-tool", *t.options.PermissionPromptToolName)
	}

	// Permission mode
	if t.options.PermissionMode != nil {
		cmd = append(cmd, "--permission-mode", string(*t.options.PermissionMode))
	}

	// Continue conversation
	if t.options.ContinueConversation {
		cmd = append(cmd, "--continue")
	}

	// Resume session
	if t.options.Resume != nil {
		cmd = append(cmd, "--resume", *t.options.Resume)
	}

	// Settings file
	if t.options.Settings != nil {
		cmd = append(cmd, "--settings", *t.options.Settings)
	}

	// Additional directories
	for _, dir := range t.options.AddDirs {
		cmd = append(cmd, "--add-dir", dir)
	}

	// MCP servers
	if len(t.options.MCPServers) > 0 {
		mcpConfig := map[string]interface{}{
			"mcpServers": t.options.MCPServers,
		}
		if configJSON, err := json.Marshal(mcpConfig); err == nil {
			cmd = append(cmd, "--mcp-config", string(configJSON))
		}
	}

	// Include partial messages
	if t.options.IncludePartialMessages {
		cmd = append(cmd, "--include-partial-messages")
	}

	// Fork session
	if t.options.ForkSession {
		cmd = append(cmd, "--fork-session")
	}

	// Agents
	if len(t.options.Agents) > 0 {
		if agentsJSON, err := json.Marshal(t.options.Agents); err == nil {
			cmd = append(cmd, "--agents", string(agentsJSON))
		}
	}

	// Setting sources
	if len(t.options.SettingSources) > 0 {
		sources := make([]string, len(t.options.SettingSources))
		for i, source := range t.options.SettingSources {
			sources[i] = string(source)
		}
		cmd = append(cmd, "--setting-sources", strings.Join(sources, ","))
	}

	// Extra arguments
	for key, value := range t.options.ExtraArgs {
		if value == nil {
			// Boolean flag without value
			cmd = append(cmd, "--"+key)
		} else {
			// Flag with value
			cmd = append(cmd, "--"+key, *value)
		}
	}

	// User
	if t.options.User != nil {
		cmd = append(cmd, "--user", *t.options.User)
	}

	// Prompt handling
	if t.isStreaming {
		// Streaming mode: use stream-json input format
		cmd = append(cmd, "--input-format", "stream-json")
	} else {
		// One-shot mode: use --print with the prompt
		cmd = append(cmd, "--print", "--", t.prompt)
	}

	return cmd
}

// Connect starts the subprocess and prepares for communication
func (t *SubprocessCLITransport) Connect(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cmd != nil {
		return nil // Already connected
	}

	// Validate CLI path exists
	if _, err := os.Stat(t.cliPath); os.IsNotExist(err) {
		return types.NewCLINotFoundError(
			fmt.Sprintf("Claude Code not found at: %s\n\nInstall with: npm install -g @anthropic-ai/claude-code\n\nOr provide the path via ClaudeAgentOptions.WithCLIPath()", t.cliPath),
			err,
		)
	}

	// Check version (skip if environment variable is set)
	if os.Getenv("CLAUDE_AGENT_SDK_SKIP_VERSION_CHECK") == "" {
		if err := t.checkClaudeVersion(ctx); err != nil {
			// Version check failure is not fatal, just log it
			// In a real implementation, you might want to log this
			fmt.Fprintf(os.Stderr, "Warning: Failed to check Claude Code version: %v\n", err)
		}
	}

	// Build command
	cmdArgs := t.buildCommand()
	t.cmd = exec.CommandContext(t.ctx, cmdArgs[0], cmdArgs[1:]...)

	// Set up environment
	processEnv := make([]string, 0, len(os.Environ())+len(t.options.Env)+2)
	processEnv = append(processEnv, os.Environ()...)

	// Add user-provided environment variables
	for k, v := range t.options.Env {
		processEnv = append(processEnv, fmt.Sprintf("%s=%s", k, v))
	}

	// Add SDK-specific environment variables
	processEnv = append(processEnv,
		fmt.Sprintf("CLAUDE_CODE_ENTRYPOINT=%s", CLICodeEntrypoint),
		fmt.Sprintf("CLAUDE_AGENT_SDK_VERSION=%s", ClaudeAgentSDKVersion),
	)

	// Set working directory PWD if different from current
	if t.cwd != "" {
		processEnv = append(processEnv, fmt.Sprintf("PWD=%s", t.cwd))
	}

	t.cmd.Env = processEnv
	t.cmd.Dir = t.cwd

	// Set up pipes
	var err error
	t.stdin, err = t.cmd.StdinPipe()
	if err != nil {
		return types.NewCLIConnectionError("failed to create stdin pipe", err)
	}

	t.stdout, err = t.cmd.StdoutPipe()
	if err != nil {
		_ = t.stdin.Close()
		return types.NewCLIConnectionError("failed to create stdout pipe", err)
	}

	// Pipe stderr if we have a callback or debug mode is enabled
	shouldPipeStderr := t.stderrCallback != nil
	for key := range t.options.ExtraArgs {
		if key == "debug-to-stderr" {
			shouldPipeStderr = true
			break
		}
	}

	if shouldPipeStderr {
		t.stderr, err = t.cmd.StderrPipe()
		if err != nil {
			_ = t.stdin.Close()
			_ = t.stdout.Close()
			return types.NewCLIConnectionError("failed to create stderr pipe", err)
		}
	}

	// Start the process
	if err := t.cmd.Start(); err != nil {
		t.cleanupPipes()
		return types.NewCLIConnectionError(fmt.Sprintf("failed to start Claude Code: %v", err), err)
	}

	// Set up buffered I/O
	t.stdoutReader = bufio.NewScanner(t.stdout)
	t.stdinWriter = bufio.NewWriter(t.stdin)

	// Start message reading loop
	go t.messageReaderLoop()

	// Start stderr handling if needed
	if shouldPipeStderr {
		go t.stderrHandler()
	}

	// Close stdin immediately for non-streaming mode
	if !t.isStreaming {
		_ = t.stdin.Close()
	}

	t.ready = true
	return nil
}

// checkClaudeVersion checks if the Claude Code CLI meets minimum version requirements
func (t *SubprocessCLITransport) checkClaudeVersion(ctx context.Context) error {
	// Create a context with timeout for version check
	versionCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(versionCtx, t.cliPath, "-v")
	output, err := cmd.Output()
	if err != nil {
		// Version check failure is not fatal
		return nil
	}

	// Parse version output
	versionStr := strings.TrimSpace(string(output))
	// Look for version pattern like "2.0.0"
	parts := strings.Split(versionStr, " ")
	for _, part := range parts {
		if strings.Contains(part, ".") {
			versionStr = part
			break
		}
	}

	// Simple version comparison
	if t.compareVersions(versionStr, MinimumClaudeCodeVersion) < 0 {
		// Version is below minimum, log warning
		// In a real implementation, you'd log this properly
		fmt.Fprintf(os.Stderr, "Warning: Claude Code version %s is unsupported in the Agent SDK. Minimum required version is %s. Some features may not work correctly.\n", versionStr, MinimumClaudeCodeVersion)
	}

	return nil
}

// compareVersions compares two version strings (returns -1, 0, or 1)
func (t *SubprocessCLITransport) compareVersions(v1, v2 string) int {
	v1Parts := strings.Split(v1, ".")
	v2Parts := strings.Split(v2, ".")

	maxLen := len(v1Parts)
	if len(v2Parts) > maxLen {
		maxLen = len(v2Parts)
	}

	for i := 0; i < maxLen; i++ {
		var v1Num, v2Num int
		if i < len(v1Parts) {
			if num, err := strconv.Atoi(v1Parts[i]); err == nil {
				v1Num = num
			}
		}
		if i < len(v2Parts) {
			if num, err := strconv.Atoi(v2Parts[i]); err == nil {
				v2Num = num
			}
		}

		if v1Num < v2Num {
			return -1
		} else if v1Num > v2Num {
			return 1
		}
	}

	return 0
}

// messageReaderLoop reads messages from stdout and sends them to the message channel
func (t *SubprocessCLITransport) messageReaderLoop() {
	defer close(t.messageChan)

	t.mu.Lock()
	reader := t.stdoutReader
	t.mu.Unlock()

	if !t.ready || reader == nil {
		return
	}

	jsonBuffer := ""

	// Configure scanner to handle long lines
	buf := make([]byte, 0, 64*1024) // 64KB initial buffer
	reader.Buffer(buf, 10*1024*1024) // 10MB max token size

	for reader.Scan() {
		select {
		case <-t.ctx.Done():
			return
		default:
		}

		line := reader.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Handle potential multiple JSON objects in one line
		jsonLines := strings.Split(line, "\n")
		for _, jsonLine := range jsonLines {
			jsonLine = strings.TrimSpace(jsonLine)
			if jsonLine == "" {
				continue
			}

			// Accumulate partial JSON
			jsonBuffer += jsonLine

			// Check buffer size
			if len(jsonBuffer) > t.maxBufferSize {
				t.OnError(types.NewJSONDecodeError(
					fmt.Sprintf("JSON message exceeded maximum buffer size of %d bytes", t.maxBufferSize),
					fmt.Errorf("buffer size %d exceeds limit %d", len(jsonBuffer), t.maxBufferSize),
				))
				jsonBuffer = ""
				continue
			}

			// Try to parse JSON
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(jsonBuffer), &data); err == nil {
				// Successfully parsed, convert to Message and send
				if message, err := t.parseMessage(data); err == nil {
					select {
					case t.messageChan <- message:
					case <-t.ctx.Done():
						return
					}
				} else {
					t.OnError(err)
				}
				jsonBuffer = ""
			}
			// If JSON parsing fails, continue accumulating (might be partial JSON)
		}
	}

	// Check for scanner errors
	if err := reader.Err(); err != nil {
		t.OnError(types.NewCLIConnectionError("error reading from stdout", err))
	}

	// Wait for process to complete and check exit code
	if t.cmd != nil && t.cmd.Process != nil {
		state, err := t.cmd.Process.Wait()
		if err == nil && state.ExitCode() != 0 {
			t.exitError = types.NewProcessError(
				fmt.Sprintf("Claude Code process exited with code %d", state.ExitCode()),
				fmt.Errorf("exit code %d", state.ExitCode()),
			)
			t.OnError(t.exitError)
		}
	}
}

// parseMessage parses a generic map into a typed Message
func (t *SubprocessCLITransport) parseMessage(data map[string]interface{}) (types.Message, error) {
	// Convert to JSON and use existing unmarshaler
	if jsonData, err := json.Marshal(data); err == nil {
		return types.UnmarshalMessage(jsonData)
	}
	return nil, types.NewMessageParseError("failed to parse message", nil)
}

// stderrHandler handles stderr output
func (t *SubprocessCLITransport) stderrHandler() {
	defer close(t.stderrDone)

	if t.stderr == nil {
		return
	}

	scanner := bufio.NewScanner(t.stderr)
	for scanner.Scan() {
		select {
		case <-t.ctx.Done():
			return
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if line != "" && t.stderrCallback != nil {
			t.stderrCallback(line)
		}
	}
}

// Write writes data to the transport
func (t *SubprocessCLITransport) Write(ctx context.Context, data string) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.ready || t.stdinWriter == nil {
		return types.NewCLIConnectionError("transport is not ready for writing", nil)
	}

	if t.cmd != nil && t.cmd.ProcessState != nil && t.cmd.ProcessState.Exited() {
		return types.NewCLIConnectionError(
			fmt.Sprintf("cannot write to terminated process (exit code: %d)", t.cmd.ProcessState.ExitCode()),
			nil,
		)
	}

	if t.exitError != nil {
		return types.NewCLIConnectionError(
			fmt.Sprintf("cannot write to process that exited with error: %v", t.exitError),
			t.exitError,
		)
	}

	// Write with newline
	if _, err := t.stdinWriter.WriteString(data + "\n"); err != nil {
		t.ready = false
		writeErr := types.NewCLIConnectionError("failed to write to stdin", err)
		t.exitError = writeErr
		return writeErr
	}

	// Flush to ensure data is sent
	if err := t.stdinWriter.Flush(); err != nil {
		t.ready = false
		flushErr := types.NewCLIConnectionError("failed to flush stdin", err)
		t.exitError = flushErr
		return flushErr
	}

	return nil
}

// ReadMessages returns a channel for reading messages
func (t *SubprocessCLITransport) ReadMessages(ctx context.Context) <-chan types.Message {
	return t.messageChan
}

// OnError handles errors from the transport
func (t *SubprocessCLITransport) OnError(err error) {
	select {
	case t.errorChan <- err:
	case <-t.ctx.Done():
	default:
		// Error channel is full, drop the error
	}
}

// IsReady returns whether the transport is ready for communication
func (t *SubprocessCLITransport) IsReady() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.ready
}

// EndInput ends the input stream (closes stdin)
func (t *SubprocessCLITransport) EndInput(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.stdinWriter != nil {
		_ = t.stdinWriter.Flush()
		t.stdinWriter = nil
	}

	if t.stdin != nil {
		if err := t.stdin.Close(); err != nil {
			return types.NewCLIConnectionError("failed to close stdin", err)
		}
		t.stdin = nil
	}

	return nil
}

// Close closes the transport and cleans up resources
func (t *SubprocessCLITransport) Close(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.ready {
		return nil
	}

	t.ready = false

	// Cancel context to stop all goroutines
	t.cancel()

	// Close stdin
	if t.stdinWriter != nil {
		_ = t.stdinWriter.Flush()
		t.stdinWriter = nil
	}
	if t.stdin != nil {
		_ = t.stdin.Close()
		t.stdin = nil
	}

	// Wait for stderr handler to finish
	if t.stderr != nil {
		select {
		case <-t.stderrDone:
		case <-time.After(5 * time.Second):
			// Timeout waiting for stderr handler
		}
		_ = t.stderr.Close()
		t.stderr = nil
	}

	// Terminate process if still running
	if t.cmd != nil && t.cmd.Process != nil {
		if t.cmd.ProcessState == nil || !t.cmd.ProcessState.Exited() {
			// Try graceful termination first
			if err := t.cmd.Process.Signal(syscall.SIGTERM); err == nil {
				// Wait a bit for graceful termination
				done := make(chan error, 1)
				go func() {
					_, err := t.cmd.Process.Wait()
					done <- err
				}()

				select {
				case <-done:
					// Process terminated gracefully
				case <-time.After(5 * time.Second):
					// Force kill if timeout
					_ = t.cmd.Process.Kill()
				}
			}
		}
	}

	// Clean up
	t.cmd = nil
	t.stdoutReader = nil
	t.exitError = nil

	// Close channels
	close(t.errorChan)

	return nil
}

// cleanupPipes cleans up standard I/O pipes
func (t *SubprocessCLITransport) cleanupPipes() {
	if t.stdin != nil {
		_ = t.stdin.Close()
	}
	if t.stdout != nil {
		_ = t.stdout.Close()
	}
	if t.stderr != nil {
		_ = t.stderr.Close()
	}
}