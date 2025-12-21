package types

import "encoding/json"

// ACP Protocol Types (Zed <-> This Agent)

type ACPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

type ACPResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

type ACPNotification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ACP Initialize

type InitializeParams struct {
	ProtocolVersion    int                `json:"protocolVersion"`
	ClientCapabilities ClientCapabilities `json:"clientCapabilities"`
	ClientInfo         ClientInfo         `json:"clientInfo"`
}

type ClientCapabilities struct {
	Filesystem *FilesystemCapability `json:"filesystem,omitempty"`
	Terminal   bool                  `json:"terminal,omitempty"`
}

type FilesystemCapability struct {
	ReadTextFile  bool `json:"readTextFile,omitempty"`
	WriteTextFile bool `json:"writeTextFile,omitempty"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResult struct {
	ProtocolVersion   int               `json:"protocolVersion"`
	AgentCapabilities AgentCapabilities `json:"agentCapabilities"`
	AgentInfo         AgentInfo         `json:"agentInfo"`
}

type AgentCapabilities struct {
	LoadSession        bool               `json:"loadSession"`
	PromptCapabilities PromptCapabilities `json:"promptCapabilities,omitempty"`
	MCP                McpInfo            `json:"mcp"`
}

type McpInfo struct {
	Http bool `json:"http"`
	Sse  bool `json:"sse"`
}

type PromptCapabilities struct {
	Image           bool `json:"image,omitempty"`
	Audio           bool `json:"audio,omitempty"`
	EmbeddedContext bool `json:"embeddedContext,omitempty"`
}

type AgentInfo struct {
	Name    string `json:"name"`
	Title   string `json:"title"`
	Version string `json:"version"`
}

// ACP Session

type NewSessionParams struct {
	Cwd string `json:"cwd"`
}

type NewSessionResult struct {
	SessionID string `json:"sessionId"`
	Models    Models `json:"models,omitempty"`
}

type Models struct {
	AvailableModels []ModelInfo `json:"availableModels"`
	CurrentModelId  ModelId     `json:"currentModelId"`
}

type PromptParams struct {
	SessionID string         `json:"sessionId"`
	Prompt    []ContentBlock `json:"prompt"`
}

type ContentBlock struct {
	Type            string          `json:"type"`
	Text            string          `json:"text,omitempty"`
	ClientResources ClientResources `json:"resource"`
}

type ClientResources struct {
	Uri      string `json:"uri"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text"`
}

type PromptResult struct {
	StopReason string `json:"stopReason"`
}

type SetModelParams struct {
	SessionID string `json:"sessionId"`
	ModelID   ModelId `json:"modelId"`
}

// ACP Session Update Notification

type ToolCall struct {
	ToolCallId string `json:"toolCallId,omitempty"`
}

type Option struct {
	OptionId string `json:"optionId"`
	Name     string `json:"name"`
	Kind     string `json:"kind"`
}

type SessionUpdateParams struct {
	SessionID string   `json:"sessionId"`
	Update    Update   `json:"update,omitempty"`
	ToolCall  ToolCall `json:"toolCall,omitempty"`
	Options   []Option `json:"options,omitempty"`
}

type SessionRequestPermissionParams struct {
	SessionID string   `json:"sessionId"`
	ToolCall  ToolCall `json:"toolCall"`
	Options   []Option `json:"options"`
}

type Update struct {
	SessionUpdate string          `json:"sessionUpdate,omitempty"`
	ToolCallId    string          `json:"toolCallId,omitempty"`
	Content       json.RawMessage `json:"content,omitempty"`
	Title         string          `json:"title,omitempty"`
	Kind          string          `json:"kind,omitempty"`
	Status        string          `json:"status,omitempty"`
}

// Droid Protocol Types (This Agent <-> Droid)

type DroidMessage struct {
	JSONRPC           string          `json:"jsonrpc"`
	Type              string          `json:"type"`
	FactoryApiVersion string          `json:"factoryApiVersion"`
	ID                string          `json:"id,omitempty"`
	Method            string          `json:"method,omitempty"`
	Params            json.RawMessage `json:"params,omitempty"`
	Result            json.RawMessage `json:"result,omitempty"`
	Error             json.RawMessage `json:"error,omitempty"`
}

type DroidNotification struct {
	Notification DroidNotificationData `json:"notification,omitempty"`
	ToolUses     []ToolUseEntry        `json:"toolUses,omitempty"`
	Options      []ToolUseOption       `json:"options,omitempty"`
}

type ToolUseEntry struct {
	ToolUse          ToolUse        `json:"toolUse"`
	ConfirmationType string         `json:"confirmationType,omitempty"`
	Details          *ToolUseDetail `json:"details,omitempty"`
}

type ToolUse struct {
	Type  string          `json:"type"`
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input,omitempty"`
}

type ToolUseDetail struct {
	Type              string   `json:"type"`
	FullCommand       string   `json:"fullCommand,omitempty"`
	Command           string   `json:"command,omitempty"`
	ExtractedCommands []string `json:"extractedCommands,omitempty"`
	ImpactLevel       string   `json:"impactLevel,omitempty"`
}

type ToolUseOption struct {
	Label          string `json:"label"`
	Value          string `json:"value"`
	SelectedColor  string `json:"selectedColor,omitempty"`
	SelectedPrefix string `json:"selectedPrefix,omitempty"`
}

type DroidNotificationData struct {
	Type      string          `json:"type"`
	Message   Message         `json:"message,omitempty"`
	TextDelta string          `json:"textDelta,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"`
	// Streaming fields for tool_use
	Index       int    `json:"index,omitempty"`
	InputDelta  string `json:"inputDelta,omitempty"`
	ToolUseID   string `json:"toolUseId,omitempty"`
	ToolUseName string `json:"toolUseName,omitempty"`
	NewState    string `json:"newState,omitempty"`
}

type Message struct {
	Id      string    `json:"id"`
	Role    string    `json:"role"`
	Content []Content `json:"content,omitempty"`
}

type Content struct {
	Id      string `json:"id"`
	Type    string `json:"type"`
	Text    string `json:"text,omitempty"`
	Path    string `json:"path,omitempty"`
	OldText string `json:"oldText,omitempty"`
	NewText string `json:"newText,omitempty"`
	Input   *Input `json:"input,omitempty"`
}

type Input struct {
	Input string `json:"input"`
}

type PatchResult struct {
	URI    string
	Before string
	After  string
}

type ResultModel struct {
	SessionID       string           `json:"sessionId"`
	Session         ResultSession    `json:"session"`
	Settings        SessionSettings  `json:"settings"`
	AvailableModels []AvailableModel `json:"availableModels"`
}

type ResultSession struct {
	Messages []Message `json:"messages"`
}

type SessionSettings struct {
	ModelID         string `json:"modelId"`
	ReasoningEffort string `json:"reasoningEffort"`
	AutonomyLevel   string `json:"autonomyLevel"`
}

type AvailableModel struct {
	ID                        string   `json:"id"`
	ModelID                   string   `json:"modelId"`
	ModelProvider             string   `json:"modelProvider"`
	DisplayName               string   `json:"displayName"`
	ShortDisplayName          string   `json:"shortDisplayName"`
	SupportedReasoningEfforts []string `json:"supportedReasoningEfforts"`
	DefaultReasoningEffort    string   `json:"defaultReasoningEffort"`
	IsCustom                  bool     `json:"isCustom"`
}

type ModelInfo struct {
	ModelId     ModelId `json:"modelId"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
}

type ModelId string
