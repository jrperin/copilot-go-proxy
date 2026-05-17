package translator

// === Anthropic API types ===

type AnthropicRequest struct {
	Model         string              `json:"model"`
	MaxTokens     int                 `json:"max_tokens"`
	System        interface{}         `json:"system,omitempty"` // string or []SystemBlock
	Messages      []AnthropicMessage  `json:"messages"`
	Tools         []AnthropicTool     `json:"tools,omitempty"`
	ToolChoice    interface{}         `json:"tool_choice,omitempty"`
	Stream        bool                `json:"stream,omitempty"`
	Temperature   *float64            `json:"temperature,omitempty"`
	StopSequences []string            `json:"stop_sequences,omitempty"`
	TopP          *float64            `json:"top_p,omitempty"`
	TopK          *int                `json:"top_k,omitempty"`
	Metadata      *AnthropicMetadata  `json:"metadata,omitempty"`
}

type SystemBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type AnthropicMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // string or []ContentBlock
}

type ContentBlock struct {
	Type      string                 `json:"type"`
	Text      string                 `json:"text,omitempty"`
	ID        string                 `json:"id,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Input     map[string]interface{} `json:"input,omitempty"`
	ToolUseID string                 `json:"tool_use_id,omitempty"`
	Content   interface{}            `json:"content,omitempty"`
	Source    *ImageSource           `json:"source,omitempty"`
}

type ImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type AnthropicTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

type AnthropicMetadata struct {
	UserID string `json:"user_id,omitempty"`
}

type AnthropicResponse struct {
	ID           string          `json:"id"`
	Type         string          `json:"type"`
	Role         string          `json:"role"`
	Content      []ContentBlock  `json:"content"`
	Model        string          `json:"model"`
	StopReason   *string         `json:"stop_reason"`
	StopSequence *string         `json:"stop_sequence"`
	Usage        AnthropicUsage  `json:"usage"`
}

type AnthropicUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
}

type AnthropicError struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// === OpenAI API types ===

type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Tools       []OpenAITool    `json:"tools,omitempty"`
	ToolChoice  interface{}     `json:"tool_choice,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature *float64        `json:"temperature,omitempty"`
	Stop        []string        `json:"stop,omitempty"`
	TopP        *float64        `json:"top_p,omitempty"`
	User        string          `json:"user,omitempty"`
}

type OpenAIMessage struct {
	Role       string          `json:"role"`
	Content    interface{}     `json:"content,omitempty"`
	ToolCalls  []OpenAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string          `json:"tool_call_id,omitempty"`
}

type OpenAIToolCall struct {
	Index     int              `json:"index,omitempty"`
	ID        string           `json:"id,omitempty"`
	Type      string           `json:"type,omitempty"`
	Function  OpenAIFunctionCall `json:"function"`
}

type OpenAIFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type OpenAITool struct {
	Type     string           `json:"type"`
	Function OpenAIFunction   `json:"function"`
}

type OpenAIFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type OpenAIResponse struct {
	ID      string           `json:"id"`
	Object  string           `json:"object"`
	Model   string           `json:"model"`
	Choices []OpenAIChoice   `json:"choices"`
	Usage   OpenAIUsage      `json:"usage,omitempty"`
}

type OpenAIChoice struct {
	Index        int            `json:"index"`
	Message      *OpenAIMessage `json:"message,omitempty"`
	Delta        *OpenAIMessage `json:"delta,omitempty"`
	FinishReason *string        `json:"finish_reason"`
}

type OpenAIUsage struct {
	PromptTokens            int                `json:"prompt_tokens"`
	CompletionTokens        int                `json:"completion_tokens"`
	TotalTokens             int                `json:"total_tokens"`
	PromptTokensDetails     *PromptTokenDetails `json:"prompt_tokens_details,omitempty"`
}

type PromptTokenDetails struct {
	CachedTokens int `json:"cached_tokens"`
}

// === SSE streaming types ===

type OpenAIChunk struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Model   string         `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   *OpenAIUsage   `json:"usage,omitempty"`
}
