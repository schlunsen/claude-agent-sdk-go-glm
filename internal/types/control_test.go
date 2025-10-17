package types

import (
	"encoding/json"
	"testing"
)

func TestPermissionRequest(t *testing.T) {
	req := &PermissionRequest{
		Subtype:  "can_use_tool",
		ToolName: "test_tool",
		Input: map[string]any{
			"param1": "value1",
		},
	}

	wrapper := &PermissionRequestWrapper{
		wrapper: &SDKControlRequest{
			Type_:   "control_request",
			ID:      "req_123",
			Request: req,
		},
		request: req,
	}

	if wrapper.Type() != "can_use_tool" {
		t.Errorf("PermissionRequest.Type() = %v, want %v", wrapper.Type(), "can_use_tool")
	}
	if wrapper.RequestID() != "req_123" {
		t.Errorf("PermissionRequest.RequestID() = %v, want %v", wrapper.RequestID(), "req_123")
	}

	data, err := json.Marshal(wrapper.wrapper)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	unmarshaled, err := UnmarshalControlRequest(data)
	if err != nil {
		t.Fatalf("UnmarshalControlRequest() error = %v", err)
	}

	permReq, ok := unmarshaled.(*PermissionRequestWrapper)
	if !ok {
		t.Fatalf("Expected *PermissionRequestWrapper, got %T", unmarshaled)
	}

	if permReq.request.ToolName != req.ToolName {
		t.Errorf("PermissionRequest ToolName = %v, want %v", permReq.request.ToolName, req.ToolName)
	}
}

func TestHookCallbackRequest(t *testing.T) {
	input := map[string]any{
		"event": "test_event",
		"data":  "test_data",
	}

	req := &HookCallbackRequest{
		Subtype:    "hook_callback",
		CallbackID: "callback_123",
		Input:      input,
		ToolUseID:  stringPtr("tool_123"),
	}

	wrapper := &HookCallbackRequestWrapper{
		wrapper: &SDKControlRequest{
			Type_:   "control_request",
			ID:      "req_456",
			Request: req,
		},
		request: req,
	}

	if wrapper.Type() != "hook_callback" {
		t.Errorf("HookCallbackRequest.Type() = %v, want %v", wrapper.Type(), "hook_callback")
	}
	if wrapper.RequestID() != "req_456" {
		t.Errorf("HookCallbackRequest.RequestID() = %v, want %v", wrapper.RequestID(), "req_456")
	}

	data, err := json.Marshal(wrapper.wrapper)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	unmarshaled, err := UnmarshalControlRequest(data)
	if err != nil {
		t.Fatalf("UnmarshalControlRequest() error = %v", err)
	}

	hookReq, ok := unmarshaled.(*HookCallbackRequestWrapper)
	if !ok {
		t.Fatalf("Expected *HookCallbackRequestWrapper, got %T", unmarshaled)
	}

	if hookReq.request.CallbackID != req.CallbackID {
		t.Errorf("HookCallbackRequest CallbackID = %v, want %v", hookReq.request.CallbackID, req.CallbackID)
	}
}

func TestMCPMessageRequest(t *testing.T) {
	message := map[string]any{
		"method": "tools/call",
		"params": map[string]any{
			"name": "test_tool",
		},
	}

	req := &MCPMessageRequest{
		Subtype:    "mcp_message",
		ServerName: "test_server",
		Message:    message,
	}

	wrapper := &MCPMessageRequestWrapper{
		wrapper: &SDKControlRequest{
			Type_:   "control_request",
			ID:      "req_789",
			Request: req,
		},
		request: req,
	}

	if wrapper.Type() != "mcp_message" {
		t.Errorf("MCPMessageRequest.Type() = %v, want %v", wrapper.Type(), "mcp_message")
	}
	if wrapper.RequestID() != "req_789" {
		t.Errorf("MCPMessageRequest.RequestID() = %v, want %v", wrapper.RequestID(), "req_789")
	}

	data, err := json.Marshal(wrapper.wrapper)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	unmarshaled, err := UnmarshalControlRequest(data)
	if err != nil {
		t.Fatalf("UnmarshalControlRequest() error = %v", err)
	}

	mcpReq, ok := unmarshaled.(*MCPMessageRequestWrapper)
	if !ok {
		t.Fatalf("Expected *MCPMessageRequestWrapper, got %T", unmarshaled)
	}

	if mcpReq.request.ServerName != req.ServerName {
		t.Errorf("MCPMessageRequest ServerName = %v, want %v", mcpReq.request.ServerName, req.ServerName)
	}
}

func TestInitializeRequest(t *testing.T) {
	hooks := map[string]interface{}{
		"PreToolUse": []string{"hook1", "hook2"},
	}

	req := &InitializeRequest{
		Subtype: "initialize",
		Hooks:   hooks,
	}

	wrapper := &InitializeRequestWrapper{
		wrapper: &SDKControlRequest{
			Type_:   "control_request",
			ID:      "req_init",
			Request: req,
		},
		request: req,
	}

	if wrapper.Type() != "initialize" {
		t.Errorf("InitializeRequest.Type() = %v, want %v", wrapper.Type(), "initialize")
	}

	data, err := json.Marshal(wrapper.wrapper)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	unmarshaled, err := UnmarshalControlRequest(data)
	if err != nil {
		t.Fatalf("UnmarshalControlRequest() error = %v", err)
	}

	initReq, ok := unmarshaled.(*InitializeRequestWrapper)
	if !ok {
		t.Fatalf("Expected *InitializeRequestWrapper, got %T", unmarshaled)
	}

	if initReq.request.Subtype != req.Subtype {
		t.Errorf("InitializeRequest Subtype = %v, want %v", initReq.request.Subtype, req.Subtype)
	}
}

func TestSuccessResponse(t *testing.T) {
	responseData := map[string]any{
		"result": "success",
		"data":   "test_data",
	}

	resp := NewSuccessResponse("req_123", responseData)

	if resp.RequestID() != "req_123" {
		t.Errorf("SuccessResponse.RequestID() = %v, want %v", resp.RequestID(), "req_123")
	}

	successResp, ok := resp.(*SuccessResponse)
	if !ok {
		t.Fatalf("Expected *SuccessResponse, got %T", resp)
	}

	if successResp.Subtype != "success" {
		t.Errorf("SuccessResponse.Subtype = %v, want %v", successResp.Subtype, "success")
	}

	data, err := MarshalControlResponse(resp)
	if err != nil {
		t.Fatalf("MarshalControlResponse() error = %v", err)
	}

	var wrapper SDKControlResponse
	if err := json.Unmarshal(data, &wrapper); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if wrapper.Type_ != "control_response" {
		t.Errorf("SDKControlResponse.Type_ = %v, want %v", wrapper.Type_, "control_response")
	}
}

func TestErrorResponse(t *testing.T) {
	errorMsg := "Permission denied"
	resp := NewErrorResponse("req_456", errorMsg)

	if resp.RequestID() != "req_456" {
		t.Errorf("ErrorResponse.RequestID() = %v, want %v", resp.RequestID(), "req_456")
	}

	errorResp, ok := resp.(*ErrorResponse)
	if !ok {
		t.Fatalf("Expected *ErrorResponse, got %T", resp)
	}

	if errorResp.Subtype != "error" {
		t.Errorf("ErrorResponse.Subtype = %v, want %v", errorResp.Subtype, "error")
	}
	if errorResp.Error != errorMsg {
		t.Errorf("ErrorResponse.Error = %v, want %v", errorResp.Error, errorMsg)
	}

	data, err := MarshalControlResponse(resp)
	if err != nil {
		t.Fatalf("MarshalControlResponse() error = %v", err)
	}

	var wrapper SDKControlResponse
	if err := json.Unmarshal(data, &wrapper); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if wrapper.Type_ != "control_response" {
		t.Errorf("SDKControlResponse.Type_ = %v, want %v", wrapper.Type_, "control_response")
	}
}

func TestUnknownControlRequestType(t *testing.T) {
	// Create a control request with unknown subtype
	data := []byte(`{
		"type": "control_request",
		"request_id": "req_unknown",
		"request": {
			"subtype": "unknown_type",
			"data": "test"
		}
	}`)

	_, err := UnmarshalControlRequest(data)
	if err == nil {
		t.Error("Expected error for unknown control request subtype")
	}

	if msgErr, ok := err.(*MessageParseError); !ok || msgErr.Message != "unknown control request subtype: unknown_type" {
		t.Errorf("Expected MessageParseError with 'unknown control request subtype: unknown_type', got %v", err)
	}
}

func TestInvalidControlRequestJSON(t *testing.T) {
	data := []byte(`{"type": "control_request", "request_id":}`) // Invalid JSON

	_, err := UnmarshalControlRequest(data)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	if _, ok := err.(*JSONDecodeError); !ok {
		t.Errorf("Expected JSONDecodeError, got %v", err)
	}
}

// unknownControlResponse is a test type that implements ControlResponse interface
type unknownControlResponse struct{}

func (u *unknownControlResponse) Type() string      { return "unknown" }
func (u *unknownControlResponse) RequestID() string { return "test_123" }

func TestUnknownControlResponseType(t *testing.T) {
	// Create a test type that implements ControlResponse interface
	// but won't be handled by the type switch
	testResp := &unknownControlResponse{}

	// This should return an error because the type switch doesn't handle this type
	_, err := MarshalControlResponse(testResp)
	if err == nil {
		t.Error("Expected error for unknown control response type")
	}

	if msgErr, ok := err.(*MessageParseError); !ok || msgErr.Message != "unknown control response type" {
		t.Errorf("Expected MessageParseError with 'unknown control response type', got %v", err)
	}
}

func TestInterruptRequest(t *testing.T) {
	req := &InterruptRequest{
		Subtype: "interrupt",
	}

	wrapper := &InterruptRequestWrapper{
		wrapper: &SDKControlRequest{
			Type_:   "control_request",
			ID:      "req_interrupt",
			Request: req,
		},
		request: req,
	}

	if wrapper.Type() != "interrupt" {
		t.Errorf("InterruptRequest.Type() = %v, want %v", wrapper.Type(), "interrupt")
	}

	data, err := json.Marshal(wrapper.wrapper)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	unmarshaled, err := UnmarshalControlRequest(data)
	if err != nil {
		t.Fatalf("UnmarshalControlRequest() error = %v", err)
	}

	interruptReq, ok := unmarshaled.(*InterruptRequestWrapper)
	if !ok {
		t.Fatalf("Expected *InterruptRequestWrapper, got %T", unmarshaled)
	}

	if interruptReq.request.Subtype != req.Subtype {
		t.Errorf("InterruptRequest Subtype = %v, want %v", interruptReq.request.Subtype, req.Subtype)
	}
}

func TestSetPermissionModeRequest(t *testing.T) {
	req := &SetPermissionModeRequest{
		Subtype: "set_permission_mode",
		Mode:    "default",
	}

	wrapper := &SetPermissionModeRequestWrapper{
		wrapper: &SDKControlRequest{
			Type_:   "control_request",
			ID:      "req_set_mode",
			Request: req,
		},
		request: req,
	}

	if wrapper.Type() != "set_permission_mode" {
		t.Errorf("SetPermissionModeRequest.Type() = %v, want %v", wrapper.Type(), "set_permission_mode")
	}

	data, err := json.Marshal(wrapper.wrapper)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	unmarshaled, err := UnmarshalControlRequest(data)
	if err != nil {
		t.Fatalf("UnmarshalControlRequest() error = %v", err)
	}

	modeReq, ok := unmarshaled.(*SetPermissionModeRequestWrapper)
	if !ok {
		t.Fatalf("Expected *SetPermissionModeRequestWrapper, got %T", unmarshaled)
	}

	if modeReq.request.Mode != req.Mode {
		t.Errorf("SetPermissionModeRequest Mode = %v, want %v", modeReq.request.Mode, req.Mode)
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
