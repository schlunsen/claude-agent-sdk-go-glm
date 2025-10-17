package types

// Message type constants
const (
	MessageTypeUser        = "user"
	MessageTypeAssistant   = "assistant"
	MessageTypeSystem      = "system"
	MessageTypeResult      = "result"
	MessageTypeStreamEvent = "stream_event"
)

// Content block type constants
const (
	ContentTypeText       = "text"
	ContentTypeThinking   = "thinking"
	ContentTypeToolUse    = "tool_use"
	ContentTypeToolResult = "tool_result"
)

// Control request/response type constants
const (
	ControlTypeRequest         = "control_request"
	ControlTypeResponse        = "control_response"
	ControlResponseTypeSuccess = "success"
	ControlResponseTypeError   = "error"
)

// Control request subtype constants
const (
	SubtypeInterrupt         = "interrupt"
	SubtypeCanUseTool        = "can_use_tool"
	SubtypeInitialize        = "initialize"
	SubtypeSetPermissionMode = "set_permission_mode"
	SubtypeHookCallback      = "hook_callback"
	SubtypeMCPMessage        = "mcp_message"
)
