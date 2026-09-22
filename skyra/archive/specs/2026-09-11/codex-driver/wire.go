package codexdriver

import "encoding/json"

type rpcMessage struct {
	ID     json.RawMessage `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type rpcRequest struct {
	Method string `json:"method"`
	ID     int64  `json:"id"`
	Params any    `json:"params"`
}

type rpcNotification struct {
	Method string `json:"method"`
	Params any    `json:"params"`
}

type rpcResponse struct {
	ID     json.RawMessage `json:"id"`
	Result any             `json:"result,omitempty"`
	Error  *rpcError       `json:"error,omitempty"`
}

type initializeParams struct {
	ClientInfo   clientInfo         `json:"clientInfo"`
	Capabilities clientCapabilities `json:"capabilities"`
}

type clientInfo struct {
	Name    string `json:"name"`
	Title   string `json:"title"`
	Version string `json:"version"`
}

type clientCapabilities struct {
	ExperimentalAPI bool `json:"experimentalApi"`
}

type threadStartParams struct {
	DynamicTools []dynamicTool `json:"dynamicTools"`
}

type dynamicTool struct {
	Type        string          `json:"type"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

type threadStartResult struct {
	Thread struct {
		ID string `json:"id"`
	} `json:"thread"`
}

type turnStartParams struct {
	ThreadID string      `json:"threadId"`
	Input    []userInput `json:"input"`
}

type userInput struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type turnStartResult struct {
	Turn turn `json:"turn"`
}

type turnNotification struct {
	ThreadID string `json:"threadId"`
	Turn     turn   `json:"turn"`
}

type turn struct {
	ID     string     `json:"id"`
	Status string     `json:"status"`
	Error  *turnError `json:"error,omitempty"`
}

type turnError struct {
	Message           string `json:"message"`
	AdditionalDetails string `json:"additionalDetails,omitempty"`
}

type itemCompletedParams struct {
	ThreadID string          `json:"threadId"`
	TurnID   string          `json:"turnId"`
	Item     json.RawMessage `json:"item"`
}

type agentMessageItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type dynamicToolCallParams struct {
	Arguments json.RawMessage `json:"arguments"`
	CallID    string          `json:"callId"`
	Namespace *string         `json:"namespace"`
	ThreadID  string          `json:"threadId"`
	Tool      string          `json:"tool"`
	TurnID    string          `json:"turnId"`
}

type dynamicToolCallResponse struct {
	ContentItems []dynamicToolContent `json:"contentItems"`
	Success      bool                 `json:"success"`
}

type dynamicToolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
