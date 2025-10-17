package types

import (
	"os"
	"testing"
)

func TestNewClaudeAgentOptions(t *testing.T) {
	opts := NewClaudeAgentOptions()

	if opts.AllowedTools == nil {
		t.Error("AllowedTools should be initialized")
	}
	if opts.MCPServers == nil {
		t.Error("MCPServers should be initialized")
	}
	if opts.DisallowedTools == nil {
		t.Error("DisallowedTools should be initialized")
	}
	if opts.AddDirs == nil {
		t.Error("AddDirs should be initialized")
	}
	if opts.Env == nil {
		t.Error("Env should be initialized")
	}
	if opts.ExtraArgs == nil {
		t.Error("ExtraArgs should be initialized")
	}
	if opts.Hooks == nil {
		t.Error("Hooks should be initialized")
	}
	if opts.Agents == nil {
		t.Error("Agents should be initialized")
	}
	if opts.SettingSources == nil {
		t.Error("SettingSources should be initialized")
	}
	if opts.ContinueConversation {
		t.Error("ContinueConversation should default to false")
	}
	if opts.ForkSession {
		t.Error("ForkSession should default to false")
	}
	if opts.IncludePartialMessages {
		t.Error("IncludePartialMessages should default to false")
	}
}

func TestClaudeAgentOptionsBuilder(t *testing.T) {
	opts := NewClaudeAgentOptions().
		WithAllowedTools("tool1", "tool2").
		WithSystemPrompt("You are a helpful assistant.").
		WithPermissionMode(PermissionModeDefault).
		WithContinueConversation(true).
		WithMaxTurns(5).
		WithModel("claude-3-sonnet").
		WithUser("test_user").
		WithIncludePartialMessages(true).
		WithForkSession(true)

	if len(opts.AllowedTools) != 2 || opts.AllowedTools[0] != "tool1" || opts.AllowedTools[1] != "tool2" {
		t.Errorf("AllowedTools = %v, want [tool1, tool2]", opts.AllowedTools)
	}
	if opts.SystemPrompt != "You are a helpful assistant." {
		t.Errorf("SystemPrompt = %v, want 'You are a helpful assistant.'", opts.SystemPrompt)
	}
	if opts.PermissionMode == nil || *opts.PermissionMode != PermissionModeDefault {
		t.Errorf("PermissionMode = %v, want %v", opts.PermissionMode, PermissionModeDefault)
	}
	if !opts.ContinueConversation {
		t.Error("ContinueConversation should be true")
	}
	if opts.MaxTurns == nil || *opts.MaxTurns != 5 {
		t.Errorf("MaxTurns = %v, want 5", opts.MaxTurns)
	}
	if opts.Model == nil || *opts.Model != "claude-3-sonnet" {
		t.Errorf("Model = %v, want 'claude-3-sonnet'", opts.Model)
	}
	if opts.User == nil || *opts.User != "test_user" {
		t.Errorf("User = %v, want 'test_user'", opts.User)
	}
	if !opts.IncludePartialMessages {
		t.Error("IncludePartialMessages should be true")
	}
	if !opts.ForkSession {
		t.Error("ForkSession should be true")
	}
}

func TestWithMCPServer(t *testing.T) {
	opts := NewClaudeAgentOptions()
	config := MCPServerConfig{
		Type:    "stdio",
		Command: "node",
		Args:    []string{"server.js"},
	}

	opts.WithMCPServer("test_server", config)

	if len(opts.MCPServers) != 1 {
		t.Errorf("MCPServers length = %v, want 1", len(opts.MCPServers))
	}

	serverConfig, exists := opts.MCPServers["test_server"]
	if !exists {
		t.Error("test_server should exist in MCPServers")
	}
	if serverConfig.Command != "node" {
		t.Errorf("MCPServer Command = %v, want 'node'", serverConfig.Command)
	}
}

func TestWithEnv(t *testing.T) {
	opts := NewClaudeAgentOptions()
	env := map[string]string{
		"VAR1": "value1",
		"VAR2": "value2",
	}

	opts.WithEnv(env)

	if len(opts.Env) != 2 {
		t.Errorf("Env length = %v, want 2", len(opts.Env))
	}
	if opts.Env["VAR1"] != "value1" {
		t.Errorf("Env[VAR1] = %v, want 'value1'", opts.Env["VAR1"])
	}
	if opts.Env["VAR2"] != "value2" {
		t.Errorf("Env[VAR2] = %v, want 'value2'", opts.Env["VAR2"])
	}
}

func TestWithHook(t *testing.T) {
	opts := NewClaudeAgentOptions()
	hook := func(ctx interface{}, input interface{}, toolUseID *string, context interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"continue_": true}, nil
	}

	matcher := HookMatcher{
		Matcher: "test_tool",
		Hooks:   []HookFunc{hook},
	}

	opts.WithHook(HookEventPreToolUse, matcher)

	if len(opts.Hooks) != 1 {
		t.Errorf("Hooks length = %v, want 1", len(opts.Hooks))
	}

	hooks, exists := opts.Hooks[HookEventPreToolUse]
	if !exists {
		t.Error("PreToolUse hook should exist")
	}
	if len(hooks) != 1 {
		t.Errorf("PreToolUse hooks length = %v, want 1", len(hooks))
	}
	if hooks[0].Matcher != "test_tool" {
		t.Errorf("Hook matcher = %v, want 'test_tool'", hooks[0].Matcher)
	}
}

func TestWithAgent(t *testing.T) {
	opts := NewClaudeAgentOptions()
	agent := AgentDefinition{
		Description: "Test agent",
		Prompt:      "You are a test agent",
		Tools:       []string{"tool1", "tool2"},
		Model:       "claude-3-haiku",
	}

	opts.WithAgent("test_agent", agent)

	if len(opts.Agents) != 1 {
		t.Errorf("Agents length = %v, want 1", len(opts.Agents))
	}

	agentDef, exists := opts.Agents["test_agent"]
	if !exists {
		t.Error("test_agent should exist in Agents")
	}
	if agentDef.Description != "Test agent" {
		t.Errorf("Agent Description = %v, want 'Test agent'", agentDef.Description)
	}
}

func TestWithSettingSources(t *testing.T) {
	opts := NewClaudeAgentOptions()
	sources := []SettingSource{SettingSourceUser, SettingSourceProject}

	opts.WithSettingSources(sources...)

	if len(opts.SettingSources) != 2 {
		t.Errorf("SettingSources length = %v, want 2", len(opts.SettingSources))
	}
	if opts.SettingSources[0] != SettingSourceUser {
		t.Errorf("SettingSources[0] = %v, want %v", opts.SettingSources[0], SettingSourceUser)
	}
	if opts.SettingSources[1] != SettingSourceProject {
		t.Errorf("SettingSources[1] = %v, want %v", opts.SettingSources[1], SettingSourceProject)
	}
}

func TestValidate(t *testing.T) {
	t.Run("valid options", func(t *testing.T) {
		opts := NewClaudeAgentOptions().
			WithPermissionMode(PermissionModeDefault).
			WithModel("claude-3-sonnet")

		err := opts.Validate()
		if err != nil {
			t.Errorf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("invalid permission mode", func(t *testing.T) {
		opts := NewClaudeAgentOptions()
		invalidMode := PermissionMode("invalid")
		opts.WithPermissionMode(invalidMode)

		err := opts.Validate()
		if err == nil {
			t.Error("Expected error for invalid permission mode")
		}
		if err.Error() != "invalid permission mode: invalid" {
			t.Errorf("Error message = %v, want 'invalid permission mode: invalid'", err.Error())
		}
	})

	t.Run("conflicting resume and continue conversation", func(t *testing.T) {
		opts := NewClaudeAgentOptions().
			WithContinueConversation(true).
			WithResume("session_123")

		err := opts.Validate()
		if err == nil {
			t.Error("Expected error for conflicting resume and continue_conversation")
		}
	})

	t.Run("non-existent CWD", func(t *testing.T) {
		opts := NewClaudeAgentOptions().
			WithCWD("/non/existent/path")

		err := opts.Validate()
		if err == nil {
			t.Error("Expected error for non-existent CWD")
		}
	})

	t.Run("non-existent CLI path", func(t *testing.T) {
		opts := NewClaudeAgentOptions().
			WithCLIPath("/non/existent/cli")

		err := opts.Validate()
		if err == nil {
			t.Error("Expected error for non-existent CLI path")
		}
	})
}

func TestGetWorkingDirectory(t *testing.T) {
	t.Run("with CWD set", func(t *testing.T) {
		cwd := "/tmp"
		opts := NewClaudeAgentOptions().WithCWD(cwd)

		if opts.GetWorkingDirectory() != cwd {
			t.Errorf("GetWorkingDirectory() = %v, want %v", opts.GetWorkingDirectory(), cwd)
		}
	})

	t.Run("without CWD set", func(t *testing.T) {
		opts := NewClaudeAgentOptions()

		wd := opts.GetWorkingDirectory()
		if wd == "" {
			t.Error("GetWorkingDirectory() should not be empty")
		}

		// Should be the current working directory
		currentWD, err := os.Getwd()
		if err != nil {
			t.Fatalf("os.Getwd() error = %v", err)
		}
		if wd != currentWD {
			t.Errorf("GetWorkingDirectory() = %v, want %v", wd, currentWD)
		}
	})
}

func TestGetCLIPath(t *testing.T) {
	opts := NewClaudeAgentOptions()
	cliPath := "/usr/bin/claude"
	opts.WithCLIPath(cliPath)

	if opts.GetCLIPath() == nil {
		t.Error("GetCLIPath() should not return nil")
	}
	if *opts.GetCLIPath() != cliPath {
		t.Errorf("GetCLIPath() = %v, want %v", *opts.GetCLIPath(), cliPath)
	}
}

func TestSystemPromptPreset(t *testing.T) {
	preset := SystemPromptPreset{
		Type:   "preset",
		Preset: "claude_code",
		Append: "Additional instructions",
	}

	if preset.Type != "preset" {
		t.Errorf("SystemPromptPreset.Type = %v, want 'preset'", preset.Type)
	}
	if preset.Preset != "claude_code" {
		t.Errorf("SystemPromptPreset.Preset = %v, want 'claude_code'", preset.Preset)
	}
	if preset.Append != "Additional instructions" {
		t.Errorf("SystemPromptPreset.Append = %v, want 'Additional instructions'", preset.Append)
	}
}

func TestAgentDefinition(t *testing.T) {
	agent := AgentDefinition{
		Description: "Test agent",
		Prompt:      "Test prompt",
		Tools:       []string{"tool1"},
		Model:       "claude-3-haiku",
	}

	if agent.Description != "Test agent" {
		t.Errorf("AgentDefinition.Description = %v, want 'Test agent'", agent.Description)
	}
	if agent.Prompt != "Test prompt" {
		t.Errorf("AgentDefinition.Prompt = %v, want 'Test prompt'", agent.Prompt)
	}
	if len(agent.Tools) != 1 || agent.Tools[0] != "tool1" {
		t.Errorf("AgentDefinition.Tools = %v, want [tool1]", agent.Tools)
	}
	if agent.Model != "claude-3-haiku" {
		t.Errorf("AgentDefinition.Model = %v, want 'claude-3-haiku'", agent.Model)
	}
}

func TestMCPServerConfig(t *testing.T) {
	config := MCPServerConfig{
		Type:    "stdio",
		Command: "node",
		Args:    []string{"server.js", "--port", "3000"},
		Env:     map[string]string{"NODE_ENV": "production"},
		Name:    "test_server",
	}

	if config.Type != "stdio" {
		t.Errorf("MCPServerConfig.Type = %v, want 'stdio'", config.Type)
	}
	if config.Command != "node" {
		t.Errorf("MCPServerConfig.Command = %v, want 'node'", config.Command)
	}
	if len(config.Args) != 3 || config.Args[0] != "server.js" || config.Args[1] != "--port" || config.Args[2] != "3000" {
		t.Errorf("MCPServerConfig.Args = %v, want [server.js --port 3000]", config.Args)
	}
	if config.Env["NODE_ENV"] != "production" {
		t.Errorf("MCPServerConfig.Env[NODE_ENV] = %v, want 'production'", config.Env["NODE_ENV"])
	}
	if config.Name != "test_server" {
		t.Errorf("MCPServerConfig.Name = %v, want 'test_server'", config.Name)
	}
}
