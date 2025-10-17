package types

import (
	"encoding/json"
)

// ContentBlock represents a content block in a message
type ContentBlock interface {
	Type() string
}

// TextBlock represents text content
type TextBlock struct {
	Type_ string `json:"type"`
	Text  string `json:"text"`
}

func (t *TextBlock) Type() string { return "text" }

// ThinkingBlock represents thinking content
type ThinkingBlock struct {
	Type_     string `json:"type"`
	Thinking  string `json:"thinking"`
	Signature string `json:"signature"`
}

func (t *ThinkingBlock) Type() string { return "thinking" }

// ToolUseBlock represents a tool use content block
type ToolUseBlock struct {
	Type_ string         `json:"type"`
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Input map[string]any `json:"input"`
}

func (t *ToolUseBlock) Type() string { return "tool_use" }

// ToolResultBlock represents a tool result content block
type ToolResultBlock struct {
	Type_     string      `json:"type"`
	ToolUseID string      `json:"tool_use_id"`
	Content   interface{} `json:"content,omitempty"`
	IsError   *bool       `json:"is_error,omitempty"`
}

func (t *ToolResultBlock) Type() string { return "tool_result" }

// UnmarshalContentBlock unmarshals JSON into the appropriate ContentBlock type
func UnmarshalContentBlock(data []byte) (ContentBlock, error) {
	var typeField struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(data, &typeField); err != nil {
		return nil, NewJSONDecodeError("failed to decode content block type", err)
	}

	switch typeField.Type {
	case "text":
		var block TextBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, NewJSONDecodeError("failed to decode text block", err)
		}
		return &block, nil
	case "thinking":
		var block ThinkingBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, NewJSONDecodeError("failed to decode thinking block", err)
		}
		return &block, nil
	case "tool_use":
		var block ToolUseBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, NewJSONDecodeError("failed to decode tool_use block", err)
		}
		return &block, nil
	case "tool_result":
		var block ToolResultBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, NewJSONDecodeError("failed to decode tool_result block", err)
		}
		return &block, nil
	default:
		return nil, NewMessageParseError("unknown content block type: "+typeField.Type, nil)
	}
}

// MarshalContentBlock marshals a ContentBlock to JSON
func MarshalContentBlock(block ContentBlock) ([]byte, error) {
	switch b := block.(type) {
	case *TextBlock:
		// Set the type field for consistency
		b.Type_ = "text"
		return json.Marshal(b)
	case *ThinkingBlock:
		b.Type_ = "thinking"
		return json.Marshal(b)
	case *ToolUseBlock:
		b.Type_ = "tool_use"
		return json.Marshal(b)
	case *ToolResultBlock:
		b.Type_ = "tool_result"
		return json.Marshal(b)
	default:
		return nil, NewMessageParseError("unknown content block type", nil)
	}
}

// Message represents a message from Claude
type Message interface {
	Type() string
}

// UserMessage represents a user message
type UserMessage struct {
	Type_           string      `json:"type"`
	Content         interface{} `json:"content"` // string or []ContentBlock
	ParentToolUseID *string     `json:"parent_tool_use_id,omitempty"`
}

func (m *UserMessage) Type() string { return "user" }

// AssistantMessage represents an assistant message with content blocks
type AssistantMessage struct {
	Type_           string         `json:"type"`
	Content         []ContentBlock `json:"content"`
	Model           string         `json:"model"`
	ParentToolUseID *string        `json:"parent_tool_use_id,omitempty"`
}

func (m *AssistantMessage) Type() string { return "assistant" }

// SystemMessage represents a system message with metadata
type SystemMessage struct {
	Type_   string         `json:"type"`
	Subtype string         `json:"subtype"`
	Data    map[string]any `json:"data"`
}

func (m *SystemMessage) Type() string { return "system" }

// ResultMessage represents a result message with cost and usage information
type ResultMessage struct {
	Type_         string         `json:"type"`
	Subtype       string         `json:"subtype"`
	DurationMS    int            `json:"duration_ms"`
	DurationAPIMS int            `json:"duration_api_ms"`
	IsError       bool           `json:"is_error"`
	NumTurns      int            `json:"num_turns"`
	SessionID     string         `json:"session_id"`
	TotalCostUSD  *float64       `json:"total_cost_usd,omitempty"`
	Usage         map[string]any `json:"usage,omitempty"`
	Result        *string        `json:"result,omitempty"`
}

func (m *ResultMessage) Type() string { return "result" }

// StreamEvent represents a stream event for partial message updates during streaming
type StreamEvent struct {
	Type_           string         `json:"type"`
	UUID            string         `json:"uuid"`
	SessionID       string         `json:"session_id"`
	Event           map[string]any `json:"event"`
	ParentToolUseID *string        `json:"parent_tool_use_id,omitempty"`
}

func (m *StreamEvent) Type() string { return "stream_event" }

// UnmarshalMessage unmarshals JSON into the appropriate Message type
func UnmarshalMessage(data []byte) (Message, error) {
	var typeField struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(data, &typeField); err != nil {
		return nil, NewJSONDecodeError("failed to decode message type", err)
	}

	switch typeField.Type {
	case "user":
		var msg UserMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, NewJSONDecodeError("failed to decode user message", err)
		}

		// Handle content field - convert to proper type if needed
		if contentStr, ok := msg.Content.(string); ok {
			msg.Content = contentStr
		} else if contentArray, ok := msg.Content.([]interface{}); ok {
			// Convert JSON array to ContentBlock slice
			blocks := make([]ContentBlock, len(contentArray))
			for i, item := range contentArray {
				itemBytes, err := json.Marshal(item)
				if err != nil {
					return nil, NewJSONDecodeError("failed to marshal content block", err)
				}
				block, err := UnmarshalContentBlock(itemBytes)
				if err != nil {
					return nil, err
				}
				blocks[i] = block
			}
			msg.Content = blocks
		}

		return &msg, nil
	case "assistant":
		var rawMsg json.RawMessage
		if err := json.Unmarshal(data, &rawMsg); err != nil {
			return nil, NewJSONDecodeError("failed to decode assistant message", err)
		}

		var assistant struct {
			Type_           string            `json:"type"`
			Content         []json.RawMessage `json:"content"`
			Model           string            `json:"model"`
			ParentToolUseID *string           `json:"parent_tool_use_id,omitempty"`
		}

		if err := json.Unmarshal(rawMsg, &assistant); err != nil {
			return nil, NewJSONDecodeError("failed to decode assistant message structure", err)
		}

		// Convert content blocks
		blocks := make([]ContentBlock, len(assistant.Content))
		for i, blockBytes := range assistant.Content {
			block, err := UnmarshalContentBlock([]byte(blockBytes))
			if err != nil {
				return nil, err
			}
			blocks[i] = block
		}

		return &AssistantMessage{
			Type_:           assistant.Type_,
			Content:         blocks,
			Model:           assistant.Model,
			ParentToolUseID: assistant.ParentToolUseID,
		}, nil
	case "system":
		var msg SystemMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, NewJSONDecodeError("failed to decode system message", err)
		}
		return &msg, nil
	case "result":
		var msg ResultMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, NewJSONDecodeError("failed to decode result message", err)
		}
		return &msg, nil
	case "stream_event":
		var msg StreamEvent
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, NewJSONDecodeError("failed to decode stream event", err)
		}
		return &msg, nil
	default:
		return nil, NewMessageParseError("unknown message type: "+typeField.Type, nil)
	}
}

// MarshalMessage marshals a Message to JSON
func MarshalMessage(msg Message) ([]byte, error) {
	switch m := msg.(type) {
	case *UserMessage:
		m.Type_ = "user"
		// Handle content blocks - if they are ContentBlock types, marshal them properly
		if blocks, ok := m.Content.([]ContentBlock); ok {
			marshaledBlocks := make([]interface{}, len(blocks))
			for i, block := range blocks {
				blockBytes, err := MarshalContentBlock(block)
				if err != nil {
					return nil, err
				}
				var blockObj interface{}
				if err := json.Unmarshal(blockBytes, &blockObj); err != nil {
					return nil, err
				}
				marshaledBlocks[i] = blockObj
			}
			// Create a temporary struct for marshaling
			tempMsg := struct {
				Type_           string      `json:"type"`
				Content         interface{} `json:"content"`
				ParentToolUseID *string     `json:"parent_tool_use_id,omitempty"`
			}{
				Type_:           m.Type_,
				Content:         marshaledBlocks,
				ParentToolUseID: m.ParentToolUseID,
			}
			return json.Marshal(tempMsg)
		}
		return json.Marshal(m)
	case *AssistantMessage:
		m.Type_ = "assistant"
		// Handle content blocks
		marshaledBlocks := make([]interface{}, len(m.Content))
		for i, block := range m.Content {
			blockBytes, err := MarshalContentBlock(block)
			if err != nil {
				return nil, err
			}
			var blockObj interface{}
			if err := json.Unmarshal(blockBytes, &blockObj); err != nil {
				return nil, err
			}
			marshaledBlocks[i] = blockObj
		}
		// Create a temporary struct for marshaling
		tempMsg := struct {
			Type_           string      `json:"type"`
			Content         interface{} `json:"content"`
			Model           string      `json:"model"`
			ParentToolUseID *string     `json:"parent_tool_use_id,omitempty"`
		}{
			Type_:           m.Type_,
			Content:         marshaledBlocks,
			Model:           m.Model,
			ParentToolUseID: m.ParentToolUseID,
		}
		return json.Marshal(tempMsg)
	case *SystemMessage:
		m.Type_ = "system"
		return json.Marshal(m)
	case *ResultMessage:
		m.Type_ = "result"
		return json.Marshal(m)
	case *StreamEvent:
		m.Type_ = "stream_event"
		return json.Marshal(m)
	default:
		return nil, NewMessageParseError("unknown message type", nil)
	}
}
