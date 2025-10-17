package types

import (
	"fmt"
	"os"
	"path/filepath"
)

// PermissionMode represents the permission mode for Claude
type PermissionMode string

const (
	PermissionModeDefault          PermissionMode = "default"
	PermissionModeAcceptEdits      PermissionMode = "acceptEdits"
	PermissionModePlan             PermissionMode = "plan"
	PermissionModeBypassPermission PermissionMode = "bypassPermissions"
)

// SettingSource represents where settings are loaded from
type SettingSource string

const (
	SettingSourceUser    SettingSource = "user"
	SettingSourceProject SettingSource = "project"
	SettingSourceLocal   SettingSource = "local"
)

// SystemPromptPreset represents a system prompt preset configuration
type SystemPromptPreset struct {
	Type   string `json:"type"`
	Preset string `json:"preset"`
	Append string `json:"append,omitempty"`
}

// AgentDefinition represents an agent definition configuration
type AgentDefinition struct {
	Description string   `json:"description"`
	Prompt      string   `json:"prompt"`
	Tools       []string `json:"tools,omitempty"`
	Model       string   `json:"model,omitempty"`
}

// MCPServerConfig represents MCP server configuration
type MCPServerConfig struct {
	Type     string            `json:"type,omitempty"`
	Command  string            `json:"command,omitempty"`
	Args     []string          `json:"args,omitempty"`
	Env      map[string]string `json:"env,omitempty"`
	URL      string            `json:"url,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
	Name     string            `json:"name,omitempty"`
	Instance interface{}       `json:"instance,omitempty"`
}

// PermissionResult represents the result of a permission check
type PermissionResult struct {
	Behavior           string             `json:"behavior"`
	UpdatedInput       map[string]any     `json:"updated_input,omitempty"`
	UpdatedPermissions []PermissionUpdate `json:"updated_permissions,omitempty"`
	Message            string             `json:"message,omitempty"`
	Interrupt          bool               `json:"interrupt,omitempty"`
}

// PermissionUpdate represents a permission update
type PermissionUpdate struct {
	Type        string           `json:"type"`
	Rules       []PermissionRule `json:"rules,omitempty"`
	Behavior    string           `json:"behavior,omitempty"`
	Mode        PermissionMode   `json:"mode,omitempty"`
	Directories []string         `json:"directories,omitempty"`
	Destination string           `json:"destination,omitempty"`
}

// PermissionRule represents a permission rule
type PermissionRule struct {
	ToolName    string `json:"toolName"`
	RuleContent string `json:"ruleContent,omitempty"`
}

// HookEvent represents hook event types
type HookEvent string

const (
	HookEventPreToolUse       HookEvent = "PreToolUse"
	HookEventPostToolUse      HookEvent = "PostToolUse"
	HookEventUserPromptSubmit HookEvent = "UserPromptSubmit"
	HookEventStop             HookEvent = "Stop"
	HookEventSubagentStop     HookEvent = "SubagentStop"
	HookEventPreCompact       HookEvent = "PreCompact"
)

// HookMatcher represents hook matcher configuration
type HookMatcher struct {
	Matcher string     `json:"matcher,omitempty"`
	Hooks   []HookFunc `json:"-"`
}

// HookFunc represents a hook function
type HookFunc func(ctx interface{}, input interface{}, toolUseID *string, context interface{}) (map[string]interface{}, error)

// ClaudeAgentOptions represents query options for Claude SDK
type ClaudeAgentOptions struct {
	// Basic options
	AllowedTools         []string                   `json:"allowed_tools,omitempty"`
	SystemPrompt         interface{}                `json:"system_prompt,omitempty"` // string or SystemPromptPreset
	MCPServers           map[string]MCPServerConfig `json:"mcp_servers,omitempty"`
	PermissionMode       *PermissionMode            `json:"permission_mode,omitempty"`
	ContinueConversation bool                       `json:"continue_conversation,omitempty"`
	Resume               *string                    `json:"resume,omitempty"`
	MaxTurns             *int                       `json:"max_turns,omitempty"`
	DisallowedTools      []string                   `json:"disallowed_tools,omitempty"`
	Model                *string                    `json:"model,omitempty"`

	// Advanced options
	PermissionPromptToolName *string            `json:"permission_prompt_tool_name,omitempty"`
	CWD                      *string            `json:"cwd,omitempty"`
	CLIPath                  *string            `json:"cli_path,omitempty"`
	Settings                 *string            `json:"settings,omitempty"`
	AddDirs                  []string           `json:"add_dirs,omitempty"`
	Env                      map[string]string  `json:"env,omitempty"`
	ExtraArgs                map[string]*string `json:"extra_args,omitempty"`
	MaxBufferSize            *int               `json:"max_buffer_size,omitempty"`
	StderrCallback           func(string)       `json:"-"` // Not serialized

	// Callbacks and hooks
	CanUseTool func(string, map[string]any, interface{}) (PermissionResult, error) `json:"-"`
	Hooks      map[HookEvent][]HookMatcher                                         `json:"hooks,omitempty"`

	// User and session options
	User                   *string                    `json:"user,omitempty"`
	IncludePartialMessages bool                       `json:"include_partial_messages,omitempty"`
	ForkSession            bool                       `json:"fork_session,omitempty"`
	Agents                 map[string]AgentDefinition `json:"agents,omitempty"`
	SettingSources         []SettingSource            `json:"setting_sources,omitempty"`
}

// NewClaudeAgentOptions creates a new ClaudeAgentOptions with defaults
func NewClaudeAgentOptions() *ClaudeAgentOptions {
	return &ClaudeAgentOptions{
		AllowedTools:           make([]string, 0),
		MCPServers:             make(map[string]MCPServerConfig),
		DisallowedTools:        make([]string, 0),
		AddDirs:                make([]string, 0),
		Env:                    make(map[string]string),
		ExtraArgs:              make(map[string]*string),
		Hooks:                  make(map[HookEvent][]HookMatcher),
		Agents:                 make(map[string]AgentDefinition),
		SettingSources:         make([]SettingSource, 0),
		ContinueConversation:   false,
		ForkSession:            false,
		IncludePartialMessages: false,
	}
}

// WithAllowedTools sets the allowed tools
func (o *ClaudeAgentOptions) WithAllowedTools(tools ...string) *ClaudeAgentOptions {
	o.AllowedTools = append(o.AllowedTools, tools...)
	return o
}

// WithSystemPrompt sets the system prompt
func (o *ClaudeAgentOptions) WithSystemPrompt(prompt interface{}) *ClaudeAgentOptions {
	o.SystemPrompt = prompt
	return o
}

// WithMCPServer adds an MCP server configuration
func (o *ClaudeAgentOptions) WithMCPServer(name string, config *MCPServerConfig) *ClaudeAgentOptions {
	if o.MCPServers == nil {
		o.MCPServers = make(map[string]MCPServerConfig)
	}
	o.MCPServers[name] = *config
	return o
}

// WithPermissionMode sets the permission mode
func (o *ClaudeAgentOptions) WithPermissionMode(mode PermissionMode) *ClaudeAgentOptions {
	o.PermissionMode = &mode
	return o
}

// WithContinueConversation sets whether to continue conversation
func (o *ClaudeAgentOptions) WithContinueConversation(continueConv bool) *ClaudeAgentOptions {
	o.ContinueConversation = continueConv
	return o
}

// WithResume sets the session to resume from
func (o *ClaudeAgentOptions) WithResume(resume string) *ClaudeAgentOptions {
	o.Resume = &resume
	return o
}

// WithMaxTurns sets the maximum number of turns
func (o *ClaudeAgentOptions) WithMaxTurns(maxTurns int) *ClaudeAgentOptions {
	o.MaxTurns = &maxTurns
	return o
}

// WithDisallowedTools sets the disallowed tools
func (o *ClaudeAgentOptions) WithDisallowedTools(tools ...string) *ClaudeAgentOptions {
	o.DisallowedTools = append(o.DisallowedTools, tools...)
	return o
}

// WithModel sets the model to use
func (o *ClaudeAgentOptions) WithModel(model string) *ClaudeAgentOptions {
	o.Model = &model
	return o
}

// WithPermissionPromptToolName sets the permission prompt tool name
func (o *ClaudeAgentOptions) WithPermissionPromptToolName(toolName string) *ClaudeAgentOptions {
	o.PermissionPromptToolName = &toolName
	return o
}

// WithCWD sets the working directory
func (o *ClaudeAgentOptions) WithCWD(cwd string) *ClaudeAgentOptions {
	// Convert to absolute path
	if !filepath.IsAbs(cwd) {
		if abs, err := filepath.Abs(cwd); err == nil {
			cwd = abs
		}
	}
	o.CWD = &cwd
	return o
}

// WithCLIPath sets the path to the CLI
func (o *ClaudeAgentOptions) WithCLIPath(cliPath string) *ClaudeAgentOptions {
	// Convert to absolute path
	if !filepath.IsAbs(cliPath) {
		if abs, err := filepath.Abs(cliPath); err == nil {
			cliPath = abs
		}
	}
	o.CLIPath = &cliPath
	return o
}

// WithSettings sets the settings file path
func (o *ClaudeAgentOptions) WithSettings(settings string) *ClaudeAgentOptions {
	o.Settings = &settings
	return o
}

// WithAddDirs adds directories to the allowed list
func (o *ClaudeAgentOptions) WithAddDirs(dirs ...string) *ClaudeAgentOptions {
	o.AddDirs = append(o.AddDirs, dirs...)
	return o
}

// WithEnv sets environment variables
func (o *ClaudeAgentOptions) WithEnv(env map[string]string) *ClaudeAgentOptions {
	if o.Env == nil {
		o.Env = make(map[string]string)
	}
	for k, v := range env {
		o.Env[k] = v
	}
	return o
}

// WithExtraArg adds an extra CLI argument
func (o *ClaudeAgentOptions) WithExtraArg(key string, value *string) *ClaudeAgentOptions {
	if o.ExtraArgs == nil {
		o.ExtraArgs = make(map[string]*string)
	}
	o.ExtraArgs[key] = value
	return o
}

// WithMaxBufferSize sets the maximum buffer size
func (o *ClaudeAgentOptions) WithMaxBufferSize(size int) *ClaudeAgentOptions {
	o.MaxBufferSize = &size
	return o
}

// WithStderrCallback sets the stderr callback
func (o *ClaudeAgentOptions) WithStderrCallback(callback func(string)) *ClaudeAgentOptions {
	o.StderrCallback = callback
	return o
}

// WithCanUseTool sets the tool permission callback
func (o *ClaudeAgentOptions) WithCanUseTool(
	callback func(string, map[string]any, interface{}) (PermissionResult, error),
) *ClaudeAgentOptions {
	o.CanUseTool = callback
	return o
}

// WithHook adds a hook for a specific event
func (o *ClaudeAgentOptions) WithHook(event HookEvent, matcher HookMatcher) *ClaudeAgentOptions {
	if o.Hooks == nil {
		o.Hooks = make(map[HookEvent][]HookMatcher)
	}
	o.Hooks[event] = append(o.Hooks[event], matcher)
	return o
}

// WithUser sets the user
func (o *ClaudeAgentOptions) WithUser(user string) *ClaudeAgentOptions {
	o.User = &user
	return o
}

// WithIncludePartialMessages sets whether to include partial messages
func (o *ClaudeAgentOptions) WithIncludePartialMessages(include bool) *ClaudeAgentOptions {
	o.IncludePartialMessages = include
	return o
}

// WithForkSession sets whether to fork the session
func (o *ClaudeAgentOptions) WithForkSession(fork bool) *ClaudeAgentOptions {
	o.ForkSession = fork
	return o
}

// WithAgent adds an agent definition
func (o *ClaudeAgentOptions) WithAgent(name string, definition AgentDefinition) *ClaudeAgentOptions {
	if o.Agents == nil {
		o.Agents = make(map[string]AgentDefinition)
	}
	o.Agents[name] = definition
	return o
}

// WithSettingSources adds setting sources
func (o *ClaudeAgentOptions) WithSettingSources(sources ...SettingSource) *ClaudeAgentOptions {
	o.SettingSources = append(o.SettingSources, sources...)
	return o
}

// Validate validates the options
func (o *ClaudeAgentOptions) Validate() error {
	// Check for conflicting options
	if o.Resume != nil && o.ContinueConversation {
		return fmt.Errorf("cannot use both resume and continue_conversation options")
	}

	// Check if CWD exists
	if o.CWD != nil {
		if _, err := os.Stat(*o.CWD); os.IsNotExist(err) {
			return fmt.Errorf("working directory does not exist: %s", *o.CWD)
		}
	}

	// Check if CLI path exists
	if o.CLIPath != nil {
		if _, err := os.Stat(*o.CLIPath); os.IsNotExist(err) {
			return fmt.Errorf("CLI path does not exist: %s", *o.CLIPath)
		}
	}

	// Validate permission mode
	if o.PermissionMode != nil {
		switch *o.PermissionMode {
		case PermissionModeDefault, PermissionModeAcceptEdits, PermissionModePlan, PermissionModeBypassPermission:
			// Valid modes
		default:
			return fmt.Errorf("invalid permission mode: %s", *o.PermissionMode)
		}
	}

	return nil
}

// GetWorkingDirectory returns the working directory, defaulting to current directory
func (o *ClaudeAgentOptions) GetWorkingDirectory() string {
	if o.CWD != nil {
		return *o.CWD
	}
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}
	return "."
}

// GetCLIPath returns the CLI path
func (o *ClaudeAgentOptions) GetCLIPath() *string {
	return o.CLIPath
}
