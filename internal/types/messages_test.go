package types

import (
	"encoding/json"
	"testing"
)

func TestTextBlock(t *testing.T) {
	block := &TextBlock{
		Text: "Hello, world!",
	}

	if block.Type() != "text" {
		t.Errorf("TextBlock.Type() = %v, want %v", block.Type(), "text")
	}

	data, err := MarshalContentBlock(block)
	if err != nil {
		t.Fatalf("MarshalContentBlock() error = %v", err)
	}

	var unmarshaled TextBlock
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Text != block.Text {
		t.Errorf("TextBlock text = %v, want %v", unmarshaled.Text, block.Text)
	}
}

func TestToolUseBlock(t *testing.T) {
	input := map[string]any{
		"param1": "value1",
		"param2": 42,
	}

	block := &ToolUseBlock{
		ID:    "tool_123",
		Name:  "test_tool",
		Input: input,
	}

	if block.Type() != "tool_use" {
		t.Errorf("ToolUseBlock.Type() = %v, want %v", block.Type(), "tool_use")
	}

	data, err := MarshalContentBlock(block)
	if err != nil {
		t.Fatalf("MarshalContentBlock() error = %v", err)
	}

	unmarshaled, err := UnmarshalContentBlock(data)
	if err != nil {
		t.Fatalf("UnmarshalContentBlock() error = %v", err)
	}

	toolBlock, ok := unmarshaled.(*ToolUseBlock)
	if !ok {
		t.Fatalf("Expected *ToolUseBlock, got %T", unmarshaled)
	}

	if toolBlock.ID != block.ID {
		t.Errorf("ToolUseBlock ID = %v, want %v", toolBlock.ID, block.ID)
	}
	if toolBlock.Name != block.Name {
		t.Errorf("ToolUseBlock Name = %v, want %v", toolBlock.Name, block.Name)
	}
}

func TestToolResultBlock(t *testing.T) {
	content := "Tool execution completed"
	isError := false

	block := &ToolResultBlock{
		ToolUseID: "tool_123",
		Content:   content,
		IsError:   &isError,
	}

	if block.Type() != "tool_result" {
		t.Errorf("ToolResultBlock.Type() = %v, want %v", block.Type(), "tool_result")
	}

	data, err := MarshalContentBlock(block)
	if err != nil {
		t.Fatalf("MarshalContentBlock() error = %v", err)
	}

	unmarshaled, err := UnmarshalContentBlock(data)
	if err != nil {
		t.Fatalf("UnmarshalContentBlock() error = %v", err)
	}

	resultBlock, ok := unmarshaled.(*ToolResultBlock)
	if !ok {
		t.Fatalf("Expected *ToolResultBlock, got %T", unmarshaled)
	}

	if resultBlock.ToolUseID != block.ToolUseID {
		t.Errorf("ToolResultBlock ToolUseID = %v, want %v", resultBlock.ToolUseID, block.ToolUseID)
	}
	if resultBlock.Content != block.Content {
		t.Errorf("ToolResultBlock Content = %v, want %v", resultBlock.Content, block.Content)
	}
	if resultBlock.IsError == nil || *resultBlock.IsError != *block.IsError {
		t.Errorf("ToolResultBlock IsError = %v, want %v", resultBlock.IsError, block.IsError)
	}
}

func TestUserMessage(t *testing.T) {
	tests := []struct {
		name    string
		message *UserMessage
	}{
		{
			name: "string content",
			message: &UserMessage{
				Content: "Hello, Claude!",
			},
		},
		{
			name: "content blocks",
			message: &UserMessage{
				Content: []ContentBlock{
					&TextBlock{Text: "Hello, "},
					&TextBlock{Text: "Claude!"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.message.Type() != "user" {
				t.Errorf("UserMessage.Type() = %v, want %v", tt.message.Type(), "user")
			}

			data, err := MarshalMessage(tt.message)
			if err != nil {
				t.Fatalf("MarshalMessage() error = %v", err)
			}

			unmarshaled, err := UnmarshalMessage(data)
			if err != nil {
				t.Fatalf("UnmarshalMessage() error = %v", err)
			}

			userMsg, ok := unmarshaled.(*UserMessage)
			if !ok {
				t.Fatalf("Expected *UserMessage, got %T", unmarshaled)
			}

			// Content comparison depends on type
			switch content := tt.message.Content.(type) {
			case string:
				if userMsg.Content != content {
					t.Errorf("UserMessage Content = %v, want %v", userMsg.Content, content)
				}
			case []ContentBlock:
				contentBlocks, ok := userMsg.Content.([]ContentBlock)
				if !ok {
					t.Fatalf("Expected []ContentBlock, got %T", userMsg.Content)
				}
				if len(contentBlocks) != len(content) {
					t.Errorf("UserMessage Content length = %v, want %v", len(contentBlocks), len(content))
				}
			}
		})
	}
}

func TestAssistantMessage(t *testing.T) {
	blocks := []ContentBlock{
		&TextBlock{Text: "I'll help you with that."},
		&ToolUseBlock{
			ID:   "tool_123",
			Name: "calculator",
			Input: map[string]any{
				"expression": "2 + 2",
			},
		},
	}

	message := &AssistantMessage{
		Content: blocks,
		Model:   "claude-3-sonnet",
	}

	if message.Type() != "assistant" {
		t.Errorf("AssistantMessage.Type() = %v, want %v", message.Type(), "assistant")
	}

	data, err := MarshalMessage(message)
	if err != nil {
		t.Fatalf("MarshalMessage() error = %v", err)
	}

	unmarshaled, err := UnmarshalMessage(data)
	if err != nil {
		t.Fatalf("UnmarshalMessage() error = %v", err)
	}

	assistantMsg, ok := unmarshaled.(*AssistantMessage)
	if !ok {
		t.Fatalf("Expected *AssistantMessage, got %T", unmarshaled)
	}

	if assistantMsg.Model != message.Model {
		t.Errorf("AssistantMessage Model = %v, want %v", assistantMsg.Model, message.Model)
	}
	if len(assistantMsg.Content) != len(message.Content) {
		t.Errorf("AssistantMessage Content length = %v, want %v", len(assistantMsg.Content), len(message.Content))
	}
}

func TestSystemMessage(t *testing.T) {
	data := map[string]any{
		"key1": "value1",
		"key2": 42,
	}

	message := &SystemMessage{
		Subtype: "test",
		Data:    data,
	}

	if message.Type() != "system" {
		t.Errorf("SystemMessage.Type() = %v, want %v", message.Type(), "system")
	}

	dataBytes, err := MarshalMessage(message)
	if err != nil {
		t.Fatalf("MarshalMessage() error = %v", err)
	}

	unmarshaled, err := UnmarshalMessage(dataBytes)
	if err != nil {
		t.Fatalf("UnmarshalMessage() error = %v", err)
	}

	systemMsg, ok := unmarshaled.(*SystemMessage)
	if !ok {
		t.Fatalf("Expected *SystemMessage, got %T", unmarshaled)
	}

	if systemMsg.Subtype != message.Subtype {
		t.Errorf("SystemMessage Subtype = %v, want %v", systemMsg.Subtype, message.Subtype)
	}
	if len(systemMsg.Data) != len(message.Data) {
		t.Errorf("SystemMessage Data length = %v, want %v", len(systemMsg.Data), len(message.Data))
	}
}

func TestResultMessage(t *testing.T) {
	result := "Task completed successfully"
	cost := 0.00123

	message := &ResultMessage{
		Subtype:       "completion",
		DurationMS:    1500,
		DurationAPIMS: 1200,
		IsError:       false,
		NumTurns:      3,
		SessionID:     "session_123",
		TotalCostUSD:  &cost,
		Result:        &result,
	}

	if message.Type() != "result" {
		t.Errorf("ResultMessage.Type() = %v, want %v", message.Type(), "result")
	}

	data, err := MarshalMessage(message)
	if err != nil {
		t.Fatalf("MarshalMessage() error = %v", err)
	}

	unmarshaled, err := UnmarshalMessage(data)
	if err != nil {
		t.Fatalf("UnmarshalMessage() error = %v", err)
	}

	resultMsg, ok := unmarshaled.(*ResultMessage)
	if !ok {
		t.Fatalf("Expected *ResultMessage, got %T", unmarshaled)
	}

	if resultMsg.Subtype != message.Subtype {
		t.Errorf("ResultMessage Subtype = %v, want %v", resultMsg.Subtype, message.Subtype)
	}
	if resultMsg.DurationMS != message.DurationMS {
		t.Errorf("ResultMessage DurationMS = %v, want %v", resultMsg.DurationMS, message.DurationMS)
	}
	if resultMsg.TotalCostUSD == nil || *resultMsg.TotalCostUSD != *message.TotalCostUSD {
		t.Errorf("ResultMessage TotalCostUSD = %v, want %v", resultMsg.TotalCostUSD, message.TotalCostUSD)
	}
}

func TestStreamEvent(t *testing.T) {
	eventData := map[string]any{
		"type": "content_block_delta",
		"delta": map[string]any{
			"type": "text_delta",
			"text": "Hello",
		},
	}

	message := &StreamEvent{
		UUID:      "uuid_123",
		SessionID: "session_123",
		Event:     eventData,
	}

	if message.Type() != "stream_event" {
		t.Errorf("StreamEvent.Type() = %v, want %v", message.Type(), "stream_event")
	}

	data, err := MarshalMessage(message)
	if err != nil {
		t.Fatalf("MarshalMessage() error = %v", err)
	}

	unmarshaled, err := UnmarshalMessage(data)
	if err != nil {
		t.Fatalf("UnmarshalMessage() error = %v", err)
	}

	streamMsg, ok := unmarshaled.(*StreamEvent)
	if !ok {
		t.Fatalf("Expected *StreamEvent, got %T", unmarshaled)
	}

	if streamMsg.UUID != message.UUID {
		t.Errorf("StreamEvent UUID = %v, want %v", streamMsg.UUID, message.UUID)
	}
	if streamMsg.SessionID != message.SessionID {
		t.Errorf("StreamEvent SessionID = %v, want %v", streamMsg.SessionID, message.SessionID)
	}
}

func TestUnknownContentBlockType(t *testing.T) {
	data := []byte(`{"type": "unknown", "data": "test"}`)

	_, err := UnmarshalContentBlock(data)
	if err == nil {
		t.Error("Expected error for unknown content block type")
	}

	if msgErr, ok := err.(*MessageParseError); !ok || msgErr.Message != "unknown content block type: unknown" {
		t.Errorf("Expected MessageParseError with 'unknown content block type: unknown', got %v", err)
	}
}

func TestUnknownMessageType(t *testing.T) {
	data := []byte(`{"type": "unknown", "data": "test"}`)

	_, err := UnmarshalMessage(data)
	if err == nil {
		t.Error("Expected error for unknown message type")
	}

	if msgErr, ok := err.(*MessageParseError); !ok || msgErr.Message != "unknown message type: unknown" {
		t.Errorf("Expected MessageParseError with 'unknown message type: unknown', got %v", err)
	}
}

func TestInvalidJSON(t *testing.T) {
	data := []byte(`{"type": "text", "text":}`) // Invalid JSON

	_, err := UnmarshalContentBlock(data)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	if _, ok := err.(*JSONDecodeError); !ok {
		t.Errorf("Expected JSONDecodeError, got %v", err)
	}
}
