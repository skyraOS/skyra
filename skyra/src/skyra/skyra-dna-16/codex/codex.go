package codex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"skyra-v1/src/skyra/skyra-dna-16/driver"
	"skyra-v1/src/skyra/skyra-dna-16/internal/support"
	"skyra-v1/src/skyra/skyra-dna-16/protocol"
	"sort"
	"strings"
	"sync"
)

const codexStateVersion uint32 = 1
const codexGenericInputSchema = `{"type":"object","additionalProperties":true}`

type CodexPhase string

const (
	CodexStartingThread CodexPhase = "startingThread"
	CodexStartingTurn   CodexPhase = "startingTurn"
	CodexRunning        CodexPhase = "running"
	CodexWaitingChild   CodexPhase = "waitingForChild"
	CodexCompleted      CodexPhase = "completed"
	CodexFaulted        CodexPhase = "faulted"
)

type CodexPendingTool struct {
	RequestID json.RawMessage `json:"requestId"`
	CallID    string          `json:"callId"`
	Tool      string          `json:"tool"`
}

// CodexInvocationState is private provider state encoded inside protocol.DriverState.
// It is exported only so recovery tests can inspect a decoded copy; the scheduler
// never depends on these fields.
type CodexInvocationState struct {
	Phase        CodexPhase                     `json:"phase"`
	ThreadID     string                         `json:"threadId,omitempty"`
	TurnID       string                         `json:"turnId,omitempty"`
	Input        string                         `json:"input,omitempty"`
	ToolTargets  map[string]protocol.Capability `json:"toolTargets,omitempty"`
	PendingTool  *CodexPendingTool              `json:"pendingTool,omitempty"`
	FinalText    string                         `json:"finalText,omitempty"`
	Contaminated bool                           `json:"contaminated,omitempty"`
}

type codexLiveInvocation struct {
	state  CodexInvocationState
	output func(protocol.DriverEmission, error) error
}

// CodexDriver owns App Server semantics. driver.Transport remains opaque mechanics.
type CodexDriver struct {
	ctx       context.Context
	cancel    context.CancelFunc
	transport driver.Transport

	mu        sync.Mutex
	nextID    int64
	live      map[protocol.InvocationID]*codexLiveInvocation
	threads   map[string]protocol.InvocationID
	requests  map[int64]protocol.InvocationID
	closed    bool
	done      chan struct{}
	ready     chan struct{}
	readyOnce sync.Once
}

func NewCodexDriver(ctx context.Context, transport driver.Transport) (*CodexDriver, error) {
	if ctx == nil || support.Missing(transport) {
		return nil, errors.New("codex driver requires context and transport")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	owned, cancel := context.WithCancel(ctx)
	driver := &CodexDriver{
		ctx: owned, cancel: cancel, transport: transport, nextID: 100,
		live: make(map[protocol.InvocationID]*codexLiveInvocation), threads: make(map[string]protocol.InvocationID),
		requests: make(map[int64]protocol.InvocationID), done: make(chan struct{}), ready: make(chan struct{}),
	}
	initialize, _ := json.Marshal(codexRPCRequest{Method: "initialize", ID: 1, Params: codexInitializeParams{
		ClientInfo:   codexClientInfo{Name: "skyra", Title: "Skyra", Version: "dna16"},
		Capabilities: codexClientCapabilities{ExperimentalAPI: true},
	}})
	if err := transport.Send(owned, initialize); err != nil {
		cancel()
		return nil, err
	}
	go driver.receive()
	return driver, nil
}

func (driver *CodexDriver) Close() error {
	driver.mu.Lock()
	if driver.closed {
		driver.mu.Unlock()
		return nil
	}
	driver.closed = true
	driver.mu.Unlock()
	driver.cancel()
	err := driver.transport.Close()
	<-driver.done
	return err
}

func (driver *CodexDriver) Done() <-chan struct{} { return driver.done }

func (driver *CodexDriver) Ingress(ctx context.Context, input protocol.DriverInput, output func(protocol.DriverEmission, error) error) error {
	if ctx == nil || output == nil {
		return errors.New("codex ingress requires context and output")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case <-driver.ready:
	case <-driver.done:
		return errors.New("Codex App Server stopped before initialization")
	case <-ctx.Done():
		return ctx.Err()
	}
	if err := (CodexInput{}).Validate(input.Move); err != nil {
		return err
	}
	invocation := input.Move.Op.(protocol.Invoke)

	if input.State.Version == 0 {
		tools, targets := compileCodexTools(invocation.Focus)
		state := CodexInvocationState{Phase: CodexStartingThread, Input: renderCodexInput(invocation), ToolTargets: targets}
		requestID := driver.allocate(invocation.InvocationID)
		command, err := marshalCodexCommand(requestID, "thread/start", codexThreadStartParams{DynamicTools: tools})
		if err != nil {
			return err
		}
		driver.bind(invocation.InvocationID, state, output)
		return driver.emit(invocation.InvocationID, state, protocol.DriverEmission{Commands: []protocol.HostCommand{command}})
	}

	state, err := decodeCodexState(input.State)
	if err != nil {
		return err
	}
	driver.mu.Lock()
	_, alreadyLive := driver.live[invocation.InvocationID]
	driver.mu.Unlock()
	if !alreadyLive {
		state.Phase = CodexFaulted
		driver.bind(invocation.InvocationID, state, output)
		return driver.emit(invocation.InvocationID, state, protocol.DriverEmission{Failure: errors.New("codex host state restored without a safely reconcilable live request")})
	}
	driver.bind(invocation.InvocationID, state, output)
	if state.Phase != CodexWaitingChild || state.PendingTool == nil {
		return errors.New("codex invocation is not waiting for a child result")
	}
	observation, err := singleCodexChildResult(invocation.Register)
	if err != nil {
		return err
	}
	response := codexRPCResponse{ID: append(json.RawMessage(nil), state.PendingTool.RequestID...), Result: codexDynamicToolResponse{
		Success: true, ContentItems: []codexDynamicToolContent{{Type: "inputText", Text: string(observation.Content.Bytes)}},
	}}
	payload, err := json.Marshal(response)
	if err != nil {
		return err
	}
	state.Phase = CodexRunning
	state.PendingTool = nil
	return driver.emit(invocation.InvocationID, state, protocol.DriverEmission{Commands: []protocol.HostCommand{{ID: "tool-response", Payload: payload}}})
}

func (driver *CodexDriver) Execute(ctx context.Context, command protocol.HostCommand) error {
	if command.ID == "" || len(command.Payload) == 0 {
		return errors.New("codex host command requires id and payload")
	}
	return driver.transport.Send(ctx, command.Payload)
}

func (*CodexDriver) Egress(raw string) ([]protocol.Move, error) {
	var outputs []codexNormalizedOutput
	if err := json.Unmarshal([]byte(raw), &outputs); err != nil {
		return nil, err
	}
	if len(outputs) != 1 {
		return nil, errors.New("codex egress requires exactly one output")
	}
	output := outputs[0]
	switch output.Kind {
	case "invoke":
		return []protocol.Move{{Target: output.Target, Op: protocol.Invoke{Action: output.Action, Params: cloneObservations(output.Params)}}}, nil
	case "return":
		if output.Value == nil {
			return nil, errors.New("codex return missing observation")
		}
		return []protocol.Move{{Op: protocol.Return{Observation: support.CloneObservation(*output.Value)}}}, nil
	default:
		return nil, fmt.Errorf("unknown codex output kind %q", output.Kind)
	}
}

type CodexInput struct{}

func (CodexInput) Declaration() json.RawMessage { return json.RawMessage(codexGenericInputSchema) }
func (CodexInput) Validate(move protocol.Move) error {
	invocation, ok := move.Op.(protocol.Invoke)
	if !ok || invocation.InvocationID.Value == "" || strings.TrimSpace(invocation.Action.Value) == "" {
		return errors.New("codex requires identified Invoke with action")
	}
	return nil
}

func (driver *CodexDriver) allocate(invocation protocol.InvocationID) int64 {
	driver.mu.Lock()
	defer driver.mu.Unlock()
	driver.nextID++
	driver.requests[driver.nextID] = invocation
	return driver.nextID
}

func (driver *CodexDriver) bind(id protocol.InvocationID, state CodexInvocationState, output func(protocol.DriverEmission, error) error) {
	driver.mu.Lock()
	defer driver.mu.Unlock()
	driver.live[id] = &codexLiveInvocation{state: cloneCodexState(state), output: output}
	if state.ThreadID != "" {
		driver.threads[state.ThreadID] = id
	}
}

func (driver *CodexDriver) emit(id protocol.InvocationID, state CodexInvocationState, emission protocol.DriverEmission) error {
	encoded, err := encodeCodexState(state)
	if err != nil {
		return err
	}
	driver.mu.Lock()
	live := driver.live[id]
	if live == nil || live.output == nil {
		driver.mu.Unlock()
		return errors.New("codex invocation has no active output binding")
	}
	live.state = cloneCodexState(state)
	output := live.output
	driver.mu.Unlock()
	emission.State = encoded
	return output(emission, nil)
}

func (driver *CodexDriver) receive() {
	defer close(driver.done)
	for {
		frame, err := driver.transport.Receive(driver.ctx)
		if err != nil {
			if driver.ctx.Err() == nil {
				driver.failAll(err)
			}
			return
		}
		if err := driver.handle(frame); err != nil {
			driver.failAll(err)
			return
		}
	}
}

func (driver *CodexDriver) handle(frame []byte) error {
	var message codexRPCMessage
	if err := json.Unmarshal(frame, &message); err != nil {
		return err
	}
	if message.Method != "" && len(message.ID) != 0 {
		return driver.handleServerRequest(message)
	}
	if message.Method != "" {
		return driver.handleNotification(message)
	}
	if len(message.ID) != 0 {
		return driver.handleResponse(message)
	}
	return errors.New("unrecognized Codex message")
}

func (driver *CodexDriver) handleResponse(message codexRPCMessage) error {
	var id int64
	if err := json.Unmarshal(message.ID, &id); err != nil {
		return err
	}
	if message.Error != nil {
		return fmt.Errorf("codex response %d: %s", id, message.Error.Message)
	}
	if id == 1 {
		initialized, err := json.Marshal(codexRPCNotification{Method: "initialized", Params: struct{}{}})
		if err != nil {
			return err
		}
		if err := driver.transport.Send(driver.ctx, initialized); err != nil {
			return err
		}
		driver.readyOnce.Do(func() { close(driver.ready) })
		return nil
	}
	driver.mu.Lock()
	invocation, exists := driver.requests[id]
	delete(driver.requests, id)
	live := driver.live[invocation]
	driver.mu.Unlock()
	if !exists || live == nil {
		return fmt.Errorf("codex response for unknown request %d", id)
	}
	state := cloneCodexState(live.state)
	switch state.Phase {
	case CodexStartingThread:
		var result codexThreadStartResult
		if err := json.Unmarshal(message.Result, &result); err != nil || result.Thread.ID == "" {
			return errors.New("invalid thread/start response")
		}
		state.ThreadID = result.Thread.ID
		state.Phase = CodexStartingTurn
		driver.mu.Lock()
		driver.threads[state.ThreadID] = invocation
		driver.mu.Unlock()
		requestID := driver.allocate(invocation)
		command, err := marshalCodexCommand(requestID, "turn/start", codexTurnStartParams{ThreadID: state.ThreadID, Input: []codexUserInput{{Type: "text", Text: state.Input}}})
		if err != nil {
			return err
		}
		return driver.emit(invocation, state, protocol.DriverEmission{Commands: []protocol.HostCommand{command}})
	case CodexStartingTurn:
		var result codexTurnStartResult
		if err := json.Unmarshal(message.Result, &result); err != nil || result.Turn.ID == "" {
			return errors.New("invalid turn/start response")
		}
		state.TurnID = result.Turn.ID
		state.Phase = CodexRunning
		return driver.emit(invocation, state, protocol.DriverEmission{})
	default:
		return fmt.Errorf("unexpected Codex response in phase %q", state.Phase)
	}
}

func (driver *CodexDriver) handleNotification(message codexRPCMessage) error {
	switch message.Method {
	case "turn/started":
		var event codexTurnNotification
		if err := json.Unmarshal(message.Params, &event); err != nil {
			return err
		}
		return driver.updateTurn(event.ThreadID, event.Turn.ID)
	case "item/started", "item/completed":
		return driver.handleItem(message.Method, message.Params)
	case "turn/completed":
		return driver.completeTurn(message.Params)
	case "item/agentMessage/delta", "item/reasoning/textDelta", "item/reasoning/summaryTextDelta", "turn/plan/updated", "thread/status/changed", "thread/tokenUsage/updated":
		return nil
	default:
		return driver.contaminateUnknown(message.Params, fmt.Errorf("unsupported Codex notification %q", message.Method))
	}
}

func (driver *CodexDriver) updateTurn(threadID, turnID string) error {
	id, live, err := driver.liveByThread(threadID)
	if err != nil {
		return err
	}
	state := cloneCodexState(live.state)
	if state.TurnID != "" && state.TurnID != turnID {
		return errors.New("Codex turn identity changed")
	}
	state.TurnID = turnID
	state.Phase = CodexRunning
	return driver.emit(id, state, protocol.DriverEmission{})
}

func (driver *CodexDriver) handleItem(method string, raw json.RawMessage) error {
	var event codexItemParams
	if err := json.Unmarshal(raw, &event); err != nil {
		return err
	}
	var header struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(event.Item, &header); err != nil || header.Type == "" {
		return errors.New("Codex item missing type")
	}
	id, live, err := driver.liveByThread(event.ThreadID)
	if err != nil {
		return err
	}
	state := cloneCodexState(live.state)
	if event.TurnID != "" && state.TurnID != "" && event.TurnID != state.TurnID {
		return errors.New("Codex item belongs to another turn")
	}
	if codexNativeEffect(header.Type) {
		state.Contaminated = true
		state.Phase = CodexFaulted
		return driver.emit(id, state, protocol.DriverEmission{Failure: fmt.Errorf("Codex invoked forbidden native item %q", header.Type)})
	}
	if !codexAllowedItem(header.Type) {
		state.Contaminated = true
		state.Phase = CodexFaulted
		return driver.emit(id, state, protocol.DriverEmission{Failure: fmt.Errorf("unknown Codex item %q", header.Type)})
	}
	if method == "item/completed" && header.Type == "agentMessage" {
		var item codexAgentMessageItem
		if err := json.Unmarshal(event.Item, &item); err != nil {
			return err
		}
		state.FinalText = item.Text
		return driver.emit(id, state, protocol.DriverEmission{})
	}
	return nil
}

func (driver *CodexDriver) completeTurn(raw json.RawMessage) error {
	var event codexTurnNotification
	if err := json.Unmarshal(raw, &event); err != nil {
		return err
	}
	id, live, err := driver.liveByThread(event.ThreadID)
	if err != nil {
		return err
	}
	state := cloneCodexState(live.state)
	if event.Turn.ID != state.TurnID {
		return errors.New("Codex completed an unexpected turn")
	}
	if state.Contaminated {
		state.Phase = CodexFaulted
		return driver.emit(id, state, protocol.DriverEmission{Failure: errors.New("contaminated Codex turn cannot return")})
	}
	if event.Turn.Status != "completed" || state.FinalText == "" {
		state.Phase = CodexFaulted
		return driver.emit(id, state, protocol.DriverEmission{Failure: errors.New("Codex turn did not complete with an agent message")})
	}
	observation := protocol.Observation{}
	observation.Content.Bytes = []byte(state.FinalText)
	encoded, _ := json.Marshal([]codexNormalizedOutput{{Kind: "return", Value: &observation}})
	state.Phase = CodexCompleted
	return driver.emit(id, state, protocol.DriverEmission{Raw: string(encoded)})
}

func (driver *CodexDriver) handleServerRequest(message codexRPCMessage) error {
	if message.Method != "item/tool/call" {
		return driver.contaminateUnknown(message.Params, fmt.Errorf("unsupported Codex server request %q", message.Method))
	}
	var call codexDynamicToolCallParams
	if err := json.Unmarshal(message.Params, &call); err != nil {
		return err
	}
	if call.Namespace != nil {
		return errors.New("namespaced Codex dynamic tools are unsupported")
	}
	id, live, err := driver.liveByThread(call.ThreadID)
	if err != nil {
		return err
	}
	state := cloneCodexState(live.state)
	if state.Phase != CodexRunning || state.TurnID != call.TurnID || state.PendingTool != nil {
		return errors.New("Codex tool call does not match a running invocation")
	}
	capability, exists := state.ToolTargets[call.Tool]
	if !exists {
		return fmt.Errorf("Codex selected unknown dynamic tool %q", call.Tool)
	}
	params, err := codexArgumentsToObservations(call.Arguments)
	if err != nil {
		return err
	}
	encoded, err := json.Marshal([]codexNormalizedOutput{{Kind: "invoke", Target: capability.Target, Action: capability.Action, Params: params}})
	if err != nil {
		return err
	}
	state.Phase = CodexWaitingChild
	state.PendingTool = &CodexPendingTool{RequestID: append(json.RawMessage(nil), message.ID...), CallID: call.CallID, Tool: call.Tool}
	return driver.emit(id, state, protocol.DriverEmission{Raw: string(encoded)})
}

func (driver *CodexDriver) contaminateUnknown(raw json.RawMessage, cause error) error {
	var correlated struct {
		ThreadID string `json:"threadId"`
	}
	_ = json.Unmarshal(raw, &correlated)
	if correlated.ThreadID == "" {
		return cause
	}
	id, live, err := driver.liveByThread(correlated.ThreadID)
	if err != nil {
		return cause
	}
	state := cloneCodexState(live.state)
	state.Contaminated = true
	state.Phase = CodexFaulted
	return driver.emit(id, state, protocol.DriverEmission{Failure: cause})
}

func (driver *CodexDriver) liveByThread(threadID string) (protocol.InvocationID, *codexLiveInvocation, error) {
	driver.mu.Lock()
	defer driver.mu.Unlock()
	id, exists := driver.threads[threadID]
	live := driver.live[id]
	if !exists || live == nil {
		return id, nil, errors.New("Codex event has unknown thread")
	}
	copy := &codexLiveInvocation{state: cloneCodexState(live.state), output: live.output}
	return id, copy, nil
}

func (driver *CodexDriver) failAll(cause error) {
	driver.mu.Lock()
	ids := make([]protocol.InvocationID, 0, len(driver.live))
	for id := range driver.live {
		ids = append(ids, id)
	}
	driver.mu.Unlock()
	for _, id := range ids {
		driver.mu.Lock()
		live := driver.live[id]
		state := cloneCodexState(live.state)
		driver.mu.Unlock()
		if state.Phase == CodexCompleted || state.Phase == CodexFaulted {
			continue
		}
		state.Contaminated = true
		state.Phase = CodexFaulted
		_ = driver.emit(id, state, protocol.DriverEmission{Failure: cause})
	}
}

func codexNativeEffect(itemType string) bool {
	switch itemType {
	case "commandExecution", "fileChange", "mcpToolCall", "webSearch", "imageView", "imageGeneration", "collabAgentToolCall", "collabToolCall", "sleep", "hookPrompt":
		return true
	default:
		return false
	}
}

func codexAllowedItem(itemType string) bool {
	switch itemType {
	case "agentMessage", "userMessage", "reasoning", "plan", "dynamicToolCall", "error":
		return true
	default:
		return false
	}
}

func encodeCodexState(state CodexInvocationState) (protocol.DriverState, error) {
	payload, err := json.Marshal(state)
	if err != nil {
		return protocol.DriverState{}, err
	}
	return protocol.DriverState{Version: codexStateVersion, Payload: payload}, nil
}

func decodeCodexState(state protocol.DriverState) (CodexInvocationState, error) {
	if state.Version != codexStateVersion || len(state.Payload) == 0 {
		return CodexInvocationState{}, fmt.Errorf("unsupported Codex driver state version %d", state.Version)
	}
	var decoded CodexInvocationState
	if err := json.Unmarshal(state.Payload, &decoded); err != nil {
		return decoded, err
	}
	return decoded, nil
}

func cloneCodexState(state CodexInvocationState) CodexInvocationState {
	encoded, _ := json.Marshal(state)
	var cloned CodexInvocationState
	_ = json.Unmarshal(encoded, &cloned)
	return cloned
}

func compileCodexTools(focus []protocol.Capability) ([]codexDynamicTool, map[string]protocol.Capability) {
	tools := make([]codexDynamicTool, len(focus))
	targets := make(map[string]protocol.Capability, len(focus))
	for index, capability := range focus {
		name := fmt.Sprintf("cap_%03d", index)
		tools[index] = codexDynamicTool{Type: "function", Name: name, Description: capability.Action.Value, InputSchema: json.RawMessage(codexGenericInputSchema)}
		targets[name] = capability
	}
	return tools, targets
}

func renderCodexInput(invocation protocol.Invoke) string {
	value := struct {
		Method string                 `json:"method"`
		Params []protocol.Observation `json:"params,omitempty"`
	}{Method: invocation.Action.Value, Params: cloneObservations(invocation.Params)}
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

func cloneObservations(values []protocol.Observation) []protocol.Observation {
	cloned := make([]protocol.Observation, len(values))
	for index, value := range values {
		cloned[index] = support.CloneObservation(value)
	}
	return cloned
}

func codexArgumentsToObservations(raw json.RawMessage) ([]protocol.Observation, error) {
	var values map[string]json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	observations := make([]protocol.Observation, len(keys))
	for index, key := range keys {
		observations[index].Segment.By.Value = key
		observations[index].Content.Bytes = append([]byte(nil), values[key]...)
	}
	return observations, nil
}

func singleCodexChildResult(register protocol.Register) (protocol.Observation, error) {
	if len(register.Order) == 0 {
		return protocol.Observation{}, errors.New("Codex child result missing")
	}
	id := register.Order[len(register.Order)-1]
	entry, exists := register.Children[id]
	if !exists || entry.Observation == nil {
		return protocol.Observation{}, errors.New("Codex child result incomplete")
	}
	return support.CloneObservation(entry.Observation.Observation), nil
}

func marshalCodexCommand(id int64, method string, params any) (protocol.HostCommand, error) {
	payload, err := json.Marshal(codexRPCRequest{Method: method, ID: id, Params: params})
	return protocol.HostCommand{ID: fmt.Sprintf("rpc-%d", id), Payload: payload}, err
}

type codexNormalizedOutput struct {
	Kind   string                 `json:"kind"`
	Target protocol.Target        `json:"target,omitempty"`
	Action protocol.Action        `json:"action,omitempty"`
	Params []protocol.Observation `json:"params,omitempty"`
	Value  *protocol.Observation  `json:"value,omitempty"`
}

type codexRPCMessage struct {
	ID     json.RawMessage `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *codexRPCError  `json:"error,omitempty"`
}

type codexRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
type codexRPCRequest struct {
	Method string `json:"method"`
	ID     int64  `json:"id"`
	Params any    `json:"params"`
}
type codexRPCNotification struct {
	Method string `json:"method"`
	Params any    `json:"params"`
}
type codexRPCResponse struct {
	ID     json.RawMessage `json:"id"`
	Result any             `json:"result,omitempty"`
	Error  *codexRPCError  `json:"error,omitempty"`
}
type codexClientInfo struct {
	Name    string `json:"name"`
	Title   string `json:"title"`
	Version string `json:"version"`
}
type codexClientCapabilities struct {
	ExperimentalAPI bool `json:"experimentalApi"`
}
type codexInitializeParams struct {
	ClientInfo   codexClientInfo         `json:"clientInfo"`
	Capabilities codexClientCapabilities `json:"capabilities"`
}
type codexThreadStartParams struct {
	DynamicTools []codexDynamicTool `json:"dynamicTools"`
}
type codexDynamicTool struct {
	Type        string          `json:"type"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}
type codexThreadStartResult struct {
	Thread struct {
		ID string `json:"id"`
	} `json:"thread"`
}
type codexTurnStartParams struct {
	ThreadID string           `json:"threadId"`
	Input    []codexUserInput `json:"input"`
}
type codexUserInput struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
type codexTurnStartResult struct {
	Turn codexTurn `json:"turn"`
}
type codexTurnNotification struct {
	ThreadID string    `json:"threadId"`
	Turn     codexTurn `json:"turn"`
}
type codexTurn struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}
type codexItemParams struct {
	ThreadID string          `json:"threadId"`
	TurnID   string          `json:"turnId"`
	Item     json.RawMessage `json:"item"`
}
type codexAgentMessageItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
type codexDynamicToolCallParams struct {
	Arguments json.RawMessage `json:"arguments"`
	CallID    string          `json:"callId"`
	Namespace *string         `json:"namespace"`
	ThreadID  string          `json:"threadId"`
	Tool      string          `json:"tool"`
	TurnID    string          `json:"turnId"`
}
type codexDynamicToolResponse struct {
	ContentItems []codexDynamicToolContent `json:"contentItems"`
	Success      bool                      `json:"success"`
}
type codexDynamicToolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

var _ driver.Driver = (*CodexDriver)(nil)
var _ driver.CommandExecutor = (*CodexDriver)(nil)
var _ driver.InputValidator = CodexInput{}
