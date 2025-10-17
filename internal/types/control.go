package types

import (
	"encoding/json"
)

// ControlRequest represents a control request from the CLI
type ControlRequest interface {
	Type() string
	RequestID() string
}

// ControlResponse represents a control response to the CLI
type ControlResponse interface {
	Type() string
	RequestID() string
}

// SDKControlRequest represents the wrapper for all control requests
type SDKControlRequest struct {
	Type_   string      `json:"type"`
	ID      string      `json:"request_id"`
	Request interface{} `json:"request"`
}

func (r *SDKControlRequest) Type() string      { return "control_request" }
func (r *SDKControlRequest) RequestID() string { return r.ID }

// SDKControlResponse represents the wrapper for all control responses
type SDKControlResponse struct {
	Type_    string      `json:"type"`
	Response interface{} `json:"response"`
}

func (r *SDKControlResponse) Type() string { return "control_response" }

// SuccessResponse represents a successful control response
type SuccessResponse struct {
	Subtype  string         `json:"subtype"`
	ID       string         `json:"request_id"`
	Response map[string]any `json:"response,omitempty"`
}

func (r *SuccessResponse) Type() string      { return "success" }
func (r *SuccessResponse) RequestID() string { return r.ID }

// ErrorResponse represents an error control response
type ErrorResponse struct {
	Subtype string `json:"subtype"`
	ID      string `json:"request_id"`
	Error   string `json:"error"`
}

func (r *ErrorResponse) Type() string      { return "error" }
func (r *ErrorResponse) RequestID() string { return r.ID }

// InterruptRequest represents an interrupt control request
type InterruptRequest struct {
	Subtype string `json:"subtype"`
}

func (r *InterruptRequest) Type() string { return "interrupt" }

// PermissionRequest represents a permission control request
type PermissionRequest struct {
	Subtype               string         `json:"subtype"`
	ToolName              string         `json:"tool_name"`
	Input                 map[string]any `json:"input"`
	PermissionSuggestions []interface{}  `json:"permission_suggestions,omitempty"`
	BlockedPath           *string        `json:"blocked_path,omitempty"`
}

func (r *PermissionRequest) Type() string { return "can_use_tool" }

// InitializeRequest represents an initialize control request
type InitializeRequest struct {
	Subtype string                 `json:"subtype"`
	Hooks   map[string]interface{} `json:"hooks,omitempty"`
}

func (r *InitializeRequest) Type() string { return "initialize" }

// SetPermissionModeRequest represents a set permission mode control request
type SetPermissionModeRequest struct {
	Subtype string `json:"subtype"`
	Mode    string `json:"mode"`
}

func (r *SetPermissionModeRequest) Type() string { return "set_permission_mode" }

// HookCallbackRequest represents a hook callback control request
type HookCallbackRequest struct {
	Subtype    string      `json:"subtype"`
	CallbackID string      `json:"callback_id"`
	Input      interface{} `json:"input"`
	ToolUseID  *string     `json:"tool_use_id,omitempty"`
}

func (r *HookCallbackRequest) Type() string { return "hook_callback" }

// MCPMessageRequest represents an MCP message control request
type MCPMessageRequest struct {
	Subtype    string      `json:"subtype"`
	ServerName string      `json:"server_name"`
	Message    interface{} `json:"message"`
}

func (r *MCPMessageRequest) Type() string { return "mcp_message" }

// UnmarshalControlRequest unmarshals JSON into the appropriate ControlRequest type
func UnmarshalControlRequest(data []byte) (ControlRequest, error) {
	var wrapper SDKControlRequest
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, NewJSONDecodeError("failed to decode control request wrapper", err)
	}

	// Extract the request part and unmarshal based on subtype
	requestBytes, err := json.Marshal(wrapper.Request)
	if err != nil {
		return nil, NewJSONDecodeError("failed to marshal request data", err)
	}

	var typeField struct {
		Subtype string `json:"subtype"`
	}

	if err := json.Unmarshal(requestBytes, &typeField); err != nil {
		return nil, NewJSONDecodeError("failed to decode control request subtype", err)
	}

	switch typeField.Subtype {
	case "interrupt":
		var req InterruptRequest
		if err := json.Unmarshal(requestBytes, &req); err != nil {
			return nil, NewJSONDecodeError("failed to decode interrupt request", err)
		}
		return &InterruptRequestWrapper{
			wrapper: &wrapper,
			request: &req,
		}, nil
	case "can_use_tool":
		var req PermissionRequest
		if err := json.Unmarshal(requestBytes, &req); err != nil {
			return nil, NewJSONDecodeError("failed to decode permission request", err)
		}
		return &PermissionRequestWrapper{
			wrapper: &wrapper,
			request: &req,
		}, nil
	case "initialize":
		var req InitializeRequest
		if err := json.Unmarshal(requestBytes, &req); err != nil {
			return nil, NewJSONDecodeError("failed to decode initialize request", err)
		}
		return &InitializeRequestWrapper{
			wrapper: &wrapper,
			request: &req,
		}, nil
	case "set_permission_mode":
		var req SetPermissionModeRequest
		if err := json.Unmarshal(requestBytes, &req); err != nil {
			return nil, NewJSONDecodeError("failed to decode set permission mode request", err)
		}
		return &SetPermissionModeRequestWrapper{
			wrapper: &wrapper,
			request: &req,
		}, nil
	case "hook_callback":
		var req HookCallbackRequest
		if err := json.Unmarshal(requestBytes, &req); err != nil {
			return nil, NewJSONDecodeError("failed to decode hook callback request", err)
		}
		return &HookCallbackRequestWrapper{
			wrapper: &wrapper,
			request: &req,
		}, nil
	case "mcp_message":
		var req MCPMessageRequest
		if err := json.Unmarshal(requestBytes, &req); err != nil {
			return nil, NewJSONDecodeError("failed to decode mcp message request", err)
		}
		return &MCPMessageRequestWrapper{
			wrapper: &wrapper,
			request: &req,
		}, nil
	default:
		return nil, NewMessageParseError("unknown control request subtype: "+typeField.Subtype, nil)
	}
}

// Wrapper types that implement ControlRequest interface
type InterruptRequestWrapper struct {
	wrapper *SDKControlRequest
	request *InterruptRequest
}

func (w *InterruptRequestWrapper) Type() string      { return w.request.Type() }
func (w *InterruptRequestWrapper) RequestID() string { return w.wrapper.ID }

type PermissionRequestWrapper struct {
	wrapper *SDKControlRequest
	request *PermissionRequest
}

func (w *PermissionRequestWrapper) Type() string      { return w.request.Type() }
func (w *PermissionRequestWrapper) RequestID() string { return w.wrapper.ID }

type InitializeRequestWrapper struct {
	wrapper *SDKControlRequest
	request *InitializeRequest
}

func (w *InitializeRequestWrapper) Type() string      { return w.request.Type() }
func (w *InitializeRequestWrapper) RequestID() string { return w.wrapper.ID }

type SetPermissionModeRequestWrapper struct {
	wrapper *SDKControlRequest
	request *SetPermissionModeRequest
}

func (w *SetPermissionModeRequestWrapper) Type() string      { return w.request.Type() }
func (w *SetPermissionModeRequestWrapper) RequestID() string { return w.wrapper.ID }

type HookCallbackRequestWrapper struct {
	wrapper *SDKControlRequest
	request *HookCallbackRequest
}

func (w *HookCallbackRequestWrapper) Type() string      { return w.request.Type() }
func (w *HookCallbackRequestWrapper) RequestID() string { return w.wrapper.ID }

type MCPMessageRequestWrapper struct {
	wrapper *SDKControlRequest
	request *MCPMessageRequest
}

func (w *MCPMessageRequestWrapper) Type() string      { return w.request.Type() }
func (w *MCPMessageRequestWrapper) RequestID() string { return w.wrapper.ID }

// MarshalControlResponse marshals a ControlResponse to JSON
func MarshalControlResponse(resp ControlResponse) ([]byte, error) {
	switch r := resp.(type) {
	case *SuccessResponse:
		wrapper := &SDKControlResponse{
			Type_:    "control_response",
			Response: r,
		}
		return json.Marshal(wrapper)
	case *ErrorResponse:
		wrapper := &SDKControlResponse{
			Type_:    "control_response",
			Response: r,
		}
		return json.Marshal(wrapper)
	default:
		return nil, NewMessageParseError("unknown control response type", nil)
	}
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(requestID string, response map[string]any) ControlResponse {
	return &SuccessResponse{
		Subtype:  "success",
		ID:       requestID,
		Response: response,
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(requestID string, errorMsg string) ControlResponse {
	return &ErrorResponse{
		Subtype: "error",
		ID:      requestID,
		Error:   errorMsg,
	}
}
