package dna17

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// codex.go reimplements the DNA-16 Codex App Server protocol adapter over DNA-17's
// Surface boundary. It is a faithful crib of skyra-dna-16/codex, NOT an import: no
// Codex identifier or protocol type (thread/turn/JSON-RPC/native-item) escapes into
// the core loop or the state contracts — they live entirely inside the opaque
// DriverState payload and this file.
//
// The Codex session is inherently long-lived and asynchronous: one Ingress returns
// after issuing a request, but the driver's receive goroutine keeps emitting
// (turn/started, item/tool/call, turn/completed) through the persistent output
// callback bound to the invocation's fd. See DECISIONS.md for how this shapes the
// fd model.

const codexStateVersion uint32 = 1
const codexGenericInputSchema = `{"type":"object","additionalProperties":true}`

type codexPhase string

const (
	codexStartingThread codexPhase = "startingThread"
	codexStartingTurn   codexPhase = "startingTurn"
	codexRunning        codexPhase = "running"
	codexWaitingChild   codexPhase = "waitingForChild"
	codexCompleted      codexPhase = "completed"
	codexFaulted        codexPhase = "faulted"
)

type codexPendingTool struct {
	RequestID json.RawMessage `json:"requestId"`
	CallID    string          `json:"callId"`
	Tool      string          `json:"tool"`
}

// codexInvocationState is private provider state encoded inside the opaque
// DriverState. The core loop never reads these fields; tests decode a copy.
type codexInvocationState struct {
	Phase        codexPhase            `json:"phase"`
	ThreadID     string                `json:"threadId,omitempty"`
	TurnID       string                `json:"turnId,omitempty"`
	Input        string                `json:"input,omitempty"`
	ToolTargets  map[string]Capability `json:"toolTargets,omitempty"`
	PendingTool  *codexPendingTool     `json:"pendingTool,omitempty"`
	FinalText    string                `json:"finalText,omitempty"`
	Contaminated bool                  `json:"contaminated,omitempty"`
}

type codexLiveInvocation struct {
	state  codexInvocationState
	output func(Emission, error) error
}

// CodexDriver owns App Server semantics. Transport remains opaque mechanics.
type CodexDriver struct {
	ctx       context.Context
	cancel    context.CancelFunc
	transport Transport

	mu        sync.Mutex
	nextID    int64
	live      map[InvocationID]*codexLiveInvocation
	threads   map[string]InvocationID
	requests  map[int64]InvocationID
	closed    bool
	done      chan struct{}
	ready     chan struct{}
	readyOnce sync.Once
}

func NewCodexDriver(ctx context.Context, transport Transport) (*CodexDriver, error) {
	if ctx == nil || transport == nil {
		return nil, errors.New("codex driver requires context and transport")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	owned, cancel := context.WithCancel(ctx)
	d := &CodexDriver{
		ctx: owned, cancel: cancel, transport: transport, nextID: 100,
		live:     make(map[InvocationID]*codexLiveInvocation),
		threads:  make(map[string]InvocationID),
		requests: make(map[int64]InvocationID),
		done:     make(chan struct{}),
		ready:    make(chan struct{}),
	}
	initialize, _ := json.Marshal(codexRPCRequest{Method: "initialize", ID: 1, Params: codexInitializeParams{
		ClientInfo:   codexClientInfo{Name: "skyra", Title: "Skyra", Version: "dna17"},
		Capabilities: codexClientCapabilities{ExperimentalAPI: true},
	}})
	if err := transport.Send(owned, initialize); err != nil {
		cancel()
		return nil, err
	}
	go d.receive()
	return d, nil
}

func (d *CodexDriver) Close() error {
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return nil
	}
	d.closed = true
	d.mu.Unlock()
	d.cancel()
	err := d.transport.Close()
	<-d.done
	return err
}

func (d *CodexDriver) Done() <-chan struct{} { return d.done }

// Ingress pushes one execution step. Version-zero state starts a thread; a
// waiting-child step answers the pending tool call with the child result.
func (d *CodexDriver) Ingress(ctx context.Context, input DriverInput, output func(Emission, error) error) error {
	if ctx == nil || output == nil {
		return errors.New("codex ingress requires context and output")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case <-d.ready:
	case <-d.done:
		return errors.New("Codex App Server stopped before initialization")
	case <-ctx.Done():
		return ctx.Err()
	}
	invocation := input.Move.Invoke
	if invocation == nil || invocation.InvocationID == "" || strings.TrimSpace(invocation.Action) == "" {
		return errors.New("codex requires an identified Invoke with an action")
	}

	if input.State.Version == 0 {
		tools, targets := compileCodexTools(invocation.Focus)
		state := codexInvocationState{Phase: codexStartingThread, Input: renderCodexInput(*invocation), ToolTargets: targets}
		requestID := d.allocate(invocation.InvocationID)
		command, err := marshalCodexCommand(requestID, "thread/start", codexThreadStartParams{DynamicTools: tools})
		if err != nil {
			return err
		}
		d.bind(invocation.InvocationID, state, output)
		return d.emit(invocation.InvocationID, state, Emission{Commands: []HostCommand{command}})
	}

	state, err := decodeCodexState(input.State)
	if err != nil {
		return err
	}
	d.mu.Lock()
	_, alreadyLive := d.live[invocation.InvocationID]
	d.mu.Unlock()
	if !alreadyLive {
		state.Phase = codexFaulted
		d.bind(invocation.InvocationID, state, output)
		return d.emit(invocation.InvocationID, state, Emission{Failure: errors.New("codex host state restored without a safely reconcilable live request")})
	}
	d.bind(invocation.InvocationID, state, output)
	if state.Phase != codexWaitingChild || state.PendingTool == nil {
		return errors.New("codex invocation is not waiting for a child result")
	}
	if input.ChildResult == nil {
		return errors.New("codex resume requires a child result")
	}
	response := codexRPCResponse{ID: append(json.RawMessage(nil), state.PendingTool.RequestID...), Result: codexDynamicToolResponse{
		Success: true, ContentItems: []codexDynamicToolContent{{Type: "inputText", Text: string(input.ChildResult.Bytes)}},
	}}
	payload, err := json.Marshal(response)
	if err != nil {
		return err
	}
	state.Phase = codexRunning
	state.PendingTool = nil
	return d.emit(invocation.InvocationID, state, Emission{Commands: []HostCommand{{ID: "tool-response", Payload: payload}}})
}

// Execute sends an opaque host command frame (post-commit).
func (d *CodexDriver) Execute(ctx context.Context, command HostCommand) error {
	if command.ID == "" || len(command.Payload) == 0 {
		return errors.New("codex host command requires id and payload")
	}
	return d.transport.Send(ctx, command.Payload)
}

// Egress normalizes an opaque emission payload into protocol Moves.
func (*CodexDriver) Egress(raw string) ([]Move, error) {
	var outputs []codexNormalizedOutput
	if err := json.Unmarshal([]byte(raw), &outputs); err != nil {
		return nil, err
	}
	if len(outputs) != 1 {
		return nil, errors.New("codex egress requires exactly one output")
	}
	out := outputs[0]
	switch out.Kind {
	case "invoke":
		return []Move{{Target: out.Target, Invoke: &Invoke{Action: out.Action, Params: cloneObservations(out.Params)}}}, nil
	case "return":
		if out.Value == nil {
			return nil, errors.New("codex return missing observation")
		}
		return []Move{{Return: &Return{Observation: cloneObservation(*out.Value)}}}, nil
	default:
		return nil, fmt.Errorf("unknown codex output kind %q", out.Kind)
	}
}

func (d *CodexDriver) allocate(invocation InvocationID) int64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.nextID++
	d.requests[d.nextID] = invocation
	return d.nextID
}

func (d *CodexDriver) bind(id InvocationID, state codexInvocationState, output func(Emission, error) error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.live[id] = &codexLiveInvocation{state: cloneCodexState(state), output: output}
	if state.ThreadID != "" {
		d.threads[state.ThreadID] = id
	}
}

func (d *CodexDriver) emit(id InvocationID, state codexInvocationState, emission Emission) error {
	encoded, err := encodeCodexState(state)
	if err != nil {
		return err
	}
	d.mu.Lock()
	live := d.live[id]
	if live == nil || live.output == nil {
		d.mu.Unlock()
		return errors.New("codex invocation has no active output binding")
	}
	live.state = cloneCodexState(state)
	output := live.output
	d.mu.Unlock()
	emission.State = encoded
	return output(emission, nil)
}

func (d *CodexDriver) receive() {
	defer close(d.done)
	for {
		frame, err := d.transport.Receive(d.ctx)
		if err != nil {
			if d.ctx.Err() == nil {
				d.failAll(err)
			}
			return
		}
		if err := d.handle(frame); err != nil {
			d.failAll(err)
			return
		}
	}
}

func (d *CodexDriver) handle(frame []byte) error {
	var message codexRPCMessage
	if err := json.Unmarshal(frame, &message); err != nil {
		return err
	}
	if message.Method != "" && len(message.ID) != 0 {
		return d.handleServerRequest(message)
	}
	if message.Method != "" {
		return d.handleNotification(message)
	}
	if len(message.ID) != 0 {
		return d.handleResponse(message)
	}
	return errors.New("unrecognized Codex message")
}

func (d *CodexDriver) handleResponse(message codexRPCMessage) error {
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
		if err := d.transport.Send(d.ctx, initialized); err != nil {
			return err
		}
		d.readyOnce.Do(func() { close(d.ready) })
		return nil
	}
	d.mu.Lock()
	invocation, exists := d.requests[id]
	delete(d.requests, id)
	live := d.live[invocation]
	d.mu.Unlock()
	if !exists || live == nil {
		return fmt.Errorf("codex response for unknown request %d", id)
	}
	state := cloneCodexState(live.state)
	switch state.Phase {
	case codexStartingThread:
		var result codexThreadStartResult
		if err := json.Unmarshal(message.Result, &result); err != nil || result.Thread.ID == "" {
			return errors.New("invalid thread/start response")
		}
		state.ThreadID = result.Thread.ID
		state.Phase = codexStartingTurn
		d.mu.Lock()
		d.threads[state.ThreadID] = invocation
		d.mu.Unlock()
		requestID := d.allocate(invocation)
		command, err := marshalCodexCommand(requestID, "turn/start", codexTurnStartParams{ThreadID: state.ThreadID, Input: []codexUserInput{{Type: "text", Text: state.Input}}})
		if err != nil {
			return err
		}
		return d.emit(invocation, state, Emission{Commands: []HostCommand{command}})
	case codexStartingTurn:
		var result codexTurnStartResult
		if err := json.Unmarshal(message.Result, &result); err != nil || result.Turn.ID == "" {
			return errors.New("invalid turn/start response")
		}
		state.TurnID = result.Turn.ID
		state.Phase = codexRunning
		return d.emit(invocation, state, Emission{})
	default:
		return fmt.Errorf("unexpected Codex response in phase %q", state.Phase)
	}
}

func (d *CodexDriver) handleNotification(message codexRPCMessage) error {
	switch message.Method {
	case "turn/started":
		var event codexTurnNotification
		if err := json.Unmarshal(message.Params, &event); err != nil {
			return err
		}
		return d.updateTurn(event.ThreadID, event.Turn.ID)
	case "item/started", "item/completed":
		return d.handleItem(message.Method, message.Params)
	case "turn/completed":
		return d.completeTurn(message.Params)
	case "item/agentMessage/delta", "item/reasoning/textDelta", "item/reasoning/summaryTextDelta", "turn/plan/updated", "thread/status/changed", "thread/tokenUsage/updated":
		return nil
	default:
		return d.contaminateUnknown(message.Params, fmt.Errorf("unsupported Codex notification %q", message.Method))
	}
}

func (d *CodexDriver) updateTurn(threadID, turnID string) error {
	id, live, err := d.liveByThread(threadID)
	if err != nil {
		return err
	}
	state := cloneCodexState(live.state)
	if state.TurnID != "" && state.TurnID != turnID {
		return errors.New("Codex turn identity changed")
	}
	state.TurnID = turnID
	state.Phase = codexRunning
	return d.emit(id, state, Emission{})
}

func (d *CodexDriver) handleItem(method string, raw json.RawMessage) error {
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
	id, live, err := d.liveByThread(event.ThreadID)
	if err != nil {
		return err
	}
	state := cloneCodexState(live.state)
	if event.TurnID != "" && state.TurnID != "" && event.TurnID != state.TurnID {
		return errors.New("Codex item belongs to another turn")
	}
	if codexNativeEffect(header.Type) {
		state.Contaminated = true
		state.Phase = codexFaulted
		return d.emit(id, state, Emission{Failure: fmt.Errorf("Codex invoked forbidden native item %q", header.Type)})
	}
	if !codexAllowedItem(header.Type) {
		state.Contaminated = true
		state.Phase = codexFaulted
		return d.emit(id, state, Emission{Failure: fmt.Errorf("unknown Codex item %q", header.Type)})
	}
	if method == "item/completed" && header.Type == "agentMessage" {
		var item codexAgentMessageItem
		if err := json.Unmarshal(event.Item, &item); err != nil {
			return err
		}
		state.FinalText = item.Text
		return d.emit(id, state, Emission{})
	}
	return nil
}

func (d *CodexDriver) completeTurn(raw json.RawMessage) error {
	var event codexTurnNotification
	if err := json.Unmarshal(raw, &event); err != nil {
		return err
	}
	id, live, err := d.liveByThread(event.ThreadID)
	if err != nil {
		return err
	}
	state := cloneCodexState(live.state)
	if event.Turn.ID != state.TurnID {
		return errors.New("Codex completed an unexpected turn")
	}
	if state.Contaminated {
		state.Phase = codexFaulted
		return d.emit(id, state, Emission{Failure: errors.New("contaminated Codex turn cannot return")})
	}
	if event.Turn.Status != "completed" || state.FinalText == "" {
		state.Phase = codexFaulted
		return d.emit(id, state, Emission{Failure: errors.New("Codex turn did not complete with an agent message")})
	}
	observation := Observation{Bytes: []byte(state.FinalText)}
	encoded, _ := json.Marshal([]codexNormalizedOutput{{Kind: "return", Value: &observation}})
	state.Phase = codexCompleted
	return d.emit(id, state, Emission{Raw: string(encoded)})
}

func (d *CodexDriver) handleServerRequest(message codexRPCMessage) error {
	if message.Method != "item/tool/call" {
		return d.contaminateUnknown(message.Params, fmt.Errorf("unsupported Codex server request %q", message.Method))
	}
	var call codexDynamicToolCallParams
	if err := json.Unmarshal(message.Params, &call); err != nil {
		return err
	}
	if call.Namespace != nil {
		return errors.New("namespaced Codex dynamic tools are unsupported")
	}
	id, live, err := d.liveByThread(call.ThreadID)
	if err != nil {
		return err
	}
	state := cloneCodexState(live.state)
	if state.Phase != codexRunning || state.TurnID != call.TurnID || state.PendingTool != nil {
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
	state.Phase = codexWaitingChild
	state.PendingTool = &codexPendingTool{RequestID: append(json.RawMessage(nil), message.ID...), CallID: call.CallID, Tool: call.Tool}
	return d.emit(id, state, Emission{Raw: string(encoded)})
}

func (d *CodexDriver) contaminateUnknown(raw json.RawMessage, cause error) error {
	var correlated struct {
		ThreadID string `json:"threadId"`
	}
	_ = json.Unmarshal(raw, &correlated)
	if correlated.ThreadID == "" {
		return cause
	}
	id, live, err := d.liveByThread(correlated.ThreadID)
	if err != nil {
		return cause
	}
	state := cloneCodexState(live.state)
	state.Contaminated = true
	state.Phase = codexFaulted
	return d.emit(id, state, Emission{Failure: cause})
}

func (d *CodexDriver) liveByThread(threadID string) (InvocationID, *codexLiveInvocation, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	id, exists := d.threads[threadID]
	live := d.live[id]
	if !exists || live == nil {
		return id, nil, errors.New("Codex event has unknown thread")
	}
	return id, &codexLiveInvocation{state: cloneCodexState(live.state), output: live.output}, nil
}

func (d *CodexDriver) failAll(cause error) {
	d.mu.Lock()
	ids := make([]InvocationID, 0, len(d.live))
	for id := range d.live {
		ids = append(ids, id)
	}
	d.mu.Unlock()
	for _, id := range ids {
		d.mu.Lock()
		live := d.live[id]
		state := cloneCodexState(live.state)
		d.mu.Unlock()
		if state.Phase == codexCompleted || state.Phase == codexFaulted {
			continue
		}
		state.Contaminated = true
		state.Phase = codexFaulted
		_ = d.emit(id, state, Emission{Failure: cause})
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

func encodeCodexState(state codexInvocationState) (DriverState, error) {
	payload, err := json.Marshal(state)
	if err != nil {
		return DriverState{}, err
	}
	return DriverState{Version: codexStateVersion, Payload: payload}, nil
}

func decodeCodexState(state DriverState) (codexInvocationState, error) {
	if state.Version != codexStateVersion || len(state.Payload) == 0 {
		return codexInvocationState{}, fmt.Errorf("unsupported Codex driver state version %d", state.Version)
	}
	var decoded codexInvocationState
	if err := json.Unmarshal(state.Payload, &decoded); err != nil {
		return decoded, err
	}
	return decoded, nil
}

func cloneCodexState(state codexInvocationState) codexInvocationState {
	encoded, _ := json.Marshal(state)
	var cloned codexInvocationState
	_ = json.Unmarshal(encoded, &cloned)
	return cloned
}

func compileCodexTools(focus []Capability) ([]codexDynamicTool, map[string]Capability) {
	tools := make([]codexDynamicTool, len(focus))
	targets := make(map[string]Capability, len(focus))
	for index, capability := range focus {
		name := fmt.Sprintf("cap_%03d", index)
		tools[index] = codexDynamicTool{Type: "function", Name: name, Description: capability.Action, InputSchema: json.RawMessage(codexGenericInputSchema)}
		targets[name] = capability
	}
	return tools, targets
}

func renderCodexInput(invocation Invoke) string {
	value := struct {
		Method string        `json:"method"`
		Params []Observation `json:"params,omitempty"`
	}{Method: invocation.Action, Params: cloneObservations(invocation.Params)}
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

func codexArgumentsToObservations(raw json.RawMessage) ([]Observation, error) {
	var values map[string]json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	observations := make([]Observation, len(keys))
	for index, key := range keys {
		observations[index].Key = key
		observations[index].Bytes = append([]byte(nil), values[key]...)
	}
	return observations, nil
}

func marshalCodexCommand(id int64, method string, params any) (HostCommand, error) {
	payload, err := json.Marshal(codexRPCRequest{Method: method, ID: id, Params: params})
	return HostCommand{ID: fmt.Sprintf("rpc-%d", id), Payload: payload}, err
}

type codexNormalizedOutput struct {
	Kind   string        `json:"kind"`
	Target SurfaceID     `json:"target,omitempty"`
	Action string        `json:"action,omitempty"`
	Params []Observation `json:"params,omitempty"`
	Value  *Observation  `json:"value,omitempty"`
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

var _ Surface = (*CodexDriver)(nil)
var _ CommandExecutor = (*CodexDriver)(nil)
