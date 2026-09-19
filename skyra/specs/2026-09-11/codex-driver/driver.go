// Package codexdriver is an executable design probe for the DNA15 driver seam.
// It deliberately imports DNA15 protocol and driver contracts instead of defining
// replacement Skyra abstractions.
package codexdriver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"skyra-v1/src/skyra/skyra-dna-15/driver"
	"skyra-v1/src/skyra/skyra-dna-15/protocol"
)

var _ driver.Driver = (*Driver)(nil)
var _ driver.InputValidator = Input{}

const genericInputSchema = `{"type":"object","additionalProperties":true}`

type phase uint8

const (
	phaseStartingThread phase = iota + 1
	phaseStartingTurn
	phaseRunning
	phaseWaitingForChild
	phaseCompleted
	phaseFaulted
)

type requestKind uint8

const (
	requestInitialize requestKind = iota + 1
	requestThreadStart
	requestTurnStart
)

type pendingRequest struct {
	kind       requestKind
	invocation protocol.InvocationID
}

type pendingTool struct {
	requestID json.RawMessage
	callID    string
}

// session is the one deliberate compromise in this DNA15 probe. These fields
// belong in scheduler-owned opaque driver state once that existing design is
// implemented. No Codex field is added to a Skyra runtime type here.
type session struct {
	phase       phase
	invocation  protocol.InvocationID
	input       string
	threadID    string
	turnID      string
	tools       map[string]protocol.Capability
	pendingTool *pendingTool
	finalText   string
	output      func(string, error) error
}

// Driver translates between DNA15 Move values and the Codex App Server protocol.
// The supplied context is the possibility-scoped host lifetime, not an invocation
// lifetime.
type Driver struct {
	ctx       context.Context
	cancel    context.CancelFunc
	transport Transport

	mu       sync.Mutex
	nextID   int64
	pending  map[int64]pendingRequest
	sessions map[protocol.InvocationID]*session
	threads  map[string]protocol.InvocationID
	closed   bool
	done     chan struct{}
}

// Input implements DNA15's existing registered endpoint validation seam.
type Input struct{}

func (Input) Declaration() json.RawMessage {
	return json.RawMessage(`{"type":"Invoke","action":"required","params":"observations"}`)
}

func (Input) Validate(move protocol.Move) error {
	invoke, ok := move.Op.(protocol.Invoke)
	if !ok {
		return errors.New("codex input requires Invoke")
	}
	if invoke.InvocationID.Value == "" || strings.TrimSpace(invoke.Action.Value) == "" {
		return errors.New("codex input requires invocation identity and action")
	}
	return nil
}

func (d *Driver) Registration() driver.DriverRegistration {
	return driver.DriverRegistration{Driver: d, Input: Input{}}
}

func New(ctx context.Context, transport Transport) (*Driver, error) {
	if ctx == nil || transport == nil {
		return nil, errors.New("codex driver requires context and transport")
	}
	hostCtx, cancel := context.WithCancel(ctx)
	d := &Driver{
		ctx:       hostCtx,
		cancel:    cancel,
		transport: transport,
		nextID:    2,
		pending: map[int64]pendingRequest{
			1: {kind: requestInitialize},
		},
		sessions: make(map[protocol.InvocationID]*session),
		threads:  make(map[string]protocol.InvocationID),
		done:     make(chan struct{}),
	}

	initialize := rpcRequest{
		Method: "initialize",
		ID:     1,
		Params: initializeParams{
			ClientInfo: clientInfo{Name: "skyra", Title: "Skyra", Version: "0.1.0"},
			Capabilities: clientCapabilities{
				ExperimentalAPI: true,
			},
		},
	}
	if err := d.send(hostCtx, initialize); err != nil {
		cancel()
		return nil, fmt.Errorf("codex initialize: %w", err)
	}
	if err := d.send(hostCtx, rpcNotification{Method: "initialized", Params: struct{}{}}); err != nil {
		cancel()
		return nil, fmt.Errorf("codex initialized: %w", err)
	}

	go d.receive()
	return d, nil
}

func (d *Driver) Close() error {
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return nil
	}
	d.closed = true
	d.mu.Unlock()
	d.cancel()
	return d.transport.Close()
}

func (d *Driver) Done() <-chan struct{} { return d.done }

// Ingress implements the existing DNA15 driver contract. A first delivery starts
// a thread. A later delivery of the same invocation answers its one pending Codex
// tool request from the scheduler-owned Register snapshot.
func (d *Driver) Ingress(ctx context.Context, move protocol.Move, output func(string, error) error) error {
	if ctx == nil || output == nil {
		return errors.New("codex ingress requires context and output")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := (Input{}).Validate(move); err != nil {
		return err
	}
	invoke := move.Op.(protocol.Invoke)

	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return errors.New("codex driver closed")
	}
	if existing := d.sessions[invoke.InvocationID]; existing != nil {
		message, err := d.resumeLocked(existing, invoke, output)
		d.mu.Unlock()
		if err != nil {
			return err
		}
		return d.transport.Send(ctx, message)
	}

	tools, targets := compileTools(invoke.Focus)
	s := &session{
		phase:      phaseStartingThread,
		invocation: invoke.InvocationID,
		input:      renderInput(invoke),
		tools:      targets,
		output:     output,
	}
	d.sessions[invoke.InvocationID] = s
	id := d.allocateLocked(requestThreadStart, invoke.InvocationID)
	d.mu.Unlock()

	request := rpcRequest{
		Method: "thread/start",
		ID:     id,
		Params: threadStartParams{DynamicTools: tools},
	}
	if err := d.send(ctx, request); err != nil {
		d.drop(invoke.InvocationID, id)
		return err
	}
	return nil
}

// Egress decodes only the private normalized representation emitted by this
// driver. Codex wire messages are consumed by receive and never enter the
// controller directly.
func (*Driver) Egress(raw string) ([]protocol.Move, error) {
	var outputs []normalizedOutput
	if err := json.Unmarshal([]byte(raw), &outputs); err != nil {
		return nil, fmt.Errorf("codex egress: %w", err)
	}
	if len(outputs) != 1 {
		return nil, errors.New("codex egress requires exactly one output")
	}
	o := outputs[0]
	switch o.Kind {
	case "invoke":
		return []protocol.Move{{
			Target: o.Target,
			Op: protocol.Invoke{
				Action: o.Action,
				Params: cloneObservations(o.Params),
			},
		}}, nil
	case "return":
		if o.Value == nil {
			return nil, errors.New("codex egress return missing observation")
		}
		return []protocol.Move{{Op: protocol.Return{Observation: cloneObservation(*o.Value)}}}, nil
	default:
		return nil, fmt.Errorf("codex egress unknown output kind %q", o.Kind)
	}
}

type normalizedOutput struct {
	Kind   string                 `json:"kind"`
	Target protocol.Target        `json:"target,omitempty"`
	Action protocol.Action        `json:"action,omitempty"`
	Params []protocol.Observation `json:"params,omitempty"`
	Value  *protocol.Observation  `json:"value,omitempty"`
}

func (d *Driver) resumeLocked(s *session, invoke protocol.Invoke, output func(string, error) error) ([]byte, error) {
	if s.phase != phaseWaitingForChild || s.pendingTool == nil {
		return nil, fmt.Errorf("codex invocation %q is not waiting for a child", invoke.InvocationID.Value)
	}
	observation, err := singleReturnedObservation(invoke.Register)
	if err != nil {
		return nil, err
	}
	response := rpcResponse{
		ID: append(json.RawMessage(nil), s.pendingTool.requestID...),
		Result: dynamicToolCallResponse{
			ContentItems: []dynamicToolContent{{Type: "inputText", Text: string(observation.Content.Bytes)}},
			Success:      true,
		},
	}
	message, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	s.pendingTool = nil
	s.output = output
	s.phase = phaseRunning
	return message, nil
}

func (d *Driver) receive() {
	defer close(d.done)
	for {
		message, err := d.transport.Receive(d.ctx)
		if err != nil {
			if d.ctx.Err() == nil {
				d.failAll(fmt.Errorf("codex transport receive: %w", err))
			}
			return
		}
		if err := d.handle(message); err != nil {
			d.failAll(err)
			return
		}
	}
}

func (d *Driver) handle(raw []byte) error {
	var message rpcMessage
	if err := json.Unmarshal(raw, &message); err != nil {
		return fmt.Errorf("codex protocol decode: %w", err)
	}
	if message.Method != "" {
		if len(message.ID) != 0 {
			return d.handleServerRequest(message)
		}
		return d.handleNotification(message)
	}
	if len(message.ID) == 0 {
		return errors.New("codex protocol message has neither method nor id")
	}
	return d.handleResponse(message)
}

func (d *Driver) handleResponse(message rpcMessage) error {
	var id int64
	if err := json.Unmarshal(message.ID, &id); err != nil {
		return fmt.Errorf("codex response id: %w", err)
	}
	d.mu.Lock()
	pending, ok := d.pending[id]
	if ok {
		delete(d.pending, id)
	}
	d.mu.Unlock()
	if !ok {
		return fmt.Errorf("codex response for unknown request %d", id)
	}
	if message.Error != nil {
		err := fmt.Errorf("codex rpc %d: %s", message.Error.Code, message.Error.Message)
		if pending.kind == requestInitialize {
			return err
		}
		d.fail(pending.invocation, err)
		return nil
	}

	switch pending.kind {
	case requestInitialize:
		return nil
	case requestThreadStart:
		return d.threadStarted(pending.invocation, message.Result)
	case requestTurnStart:
		return d.turnStarted(pending.invocation, message.Result)
	default:
		return errors.New("codex response has unknown request kind")
	}
}

func (d *Driver) threadStarted(invocation protocol.InvocationID, raw json.RawMessage) error {
	var result threadStartResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return fmt.Errorf("codex thread/start response invalid: %w", err)
	}
	if result.Thread.ID == "" {
		return errors.New("codex thread/start response missing thread id")
	}

	d.mu.Lock()
	s := d.sessions[invocation]
	if s == nil || s.phase != phaseStartingThread {
		d.mu.Unlock()
		return errors.New("codex thread response has no starting invocation")
	}
	s.threadID = result.Thread.ID
	s.phase = phaseStartingTurn
	d.threads[s.threadID] = invocation
	id := d.allocateLocked(requestTurnStart, invocation)
	input := s.input
	s.input = ""
	d.mu.Unlock()

	return d.send(d.ctx, rpcRequest{
		Method: "turn/start",
		ID:     id,
		Params: turnStartParams{ThreadID: result.Thread.ID, Input: []userInput{{Type: "text", Text: input}}},
	})
}

func (d *Driver) turnStarted(invocation protocol.InvocationID, raw json.RawMessage) error {
	var result turnStartResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return fmt.Errorf("codex turn/start response invalid: %w", err)
	}
	if result.Turn.ID == "" {
		return errors.New("codex turn/start response missing turn id")
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	s := d.sessions[invocation]
	if s == nil || (s.phase != phaseStartingTurn && s.phase != phaseRunning) {
		return errors.New("codex turn response has no starting invocation")
	}
	if s.turnID != "" && s.turnID != result.Turn.ID {
		return errors.New("codex turn response conflicts with turn notification")
	}
	s.turnID = result.Turn.ID
	s.phase = phaseRunning
	return nil
}

func (d *Driver) handleNotification(message rpcMessage) error {
	switch message.Method {
	case "turn/started":
		var event turnNotification
		if err := json.Unmarshal(message.Params, &event); err != nil {
			return err
		}
		d.mu.Lock()
		if s := d.sessionByThreadLocked(event.ThreadID); s != nil {
			s.turnID = event.Turn.ID
			if s.phase == phaseStartingTurn {
				s.phase = phaseRunning
			}
		}
		d.mu.Unlock()
		return nil
	case "item/completed":
		return d.itemCompleted(message.Params)
	case "turn/completed":
		return d.turnCompleted(message.Params)
	default:
		return nil
	}
}

func (d *Driver) itemCompleted(raw json.RawMessage) error {
	var event itemCompletedParams
	if err := json.Unmarshal(raw, &event); err != nil {
		return err
	}
	var item agentMessageItem
	if err := json.Unmarshal(event.Item, &item); err != nil || item.Type != "agentMessage" {
		return nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if s := d.sessionByThreadLocked(event.ThreadID); s != nil && s.turnID == event.TurnID {
		s.finalText = item.Text
	}
	return nil
}

func (d *Driver) turnCompleted(raw json.RawMessage) error {
	var event turnNotification
	if err := json.Unmarshal(raw, &event); err != nil {
		return err
	}
	d.mu.Lock()
	s := d.sessionByThreadLocked(event.ThreadID)
	if s == nil {
		d.mu.Unlock()
		return nil
	}
	if event.Turn.ID != s.turnID {
		d.mu.Unlock()
		return errors.New("codex completed an unexpected turn")
	}
	if event.Turn.Status != "completed" {
		s.phase = phaseFaulted
		output := takeOutput(s)
		d.mu.Unlock()
		if output != nil {
			message := "codex turn " + event.Turn.Status
			if event.Turn.Error != nil && event.Turn.Error.Message != "" {
				message += ": " + event.Turn.Error.Message
			}
			return output("", errors.New(message))
		}
		return nil
	}
	if s.finalText == "" {
		s.phase = phaseFaulted
		output := takeOutput(s)
		d.mu.Unlock()
		if output != nil {
			return output("", errors.New("codex turn completed without an agent message"))
		}
		return nil
	}
	o := protocol.Observation{}
	o.Content.Bytes = []byte(s.finalText)
	normalized, err := json.Marshal([]normalizedOutput{{Kind: "return", Value: &o}})
	if err != nil {
		d.mu.Unlock()
		return err
	}
	s.phase = phaseCompleted
	output := takeOutput(s)
	d.mu.Unlock()
	if output == nil {
		return errors.New("codex completion has no active handoff")
	}
	return output(string(normalized), nil)
}

func (d *Driver) handleServerRequest(message rpcMessage) error {
	if message.Method != "item/tool/call" {
		_ = d.send(d.ctx, rpcResponse{ID: message.ID, Error: &rpcError{Code: -32601, Message: "unsupported Codex server request"}})
		return fmt.Errorf("unsupported Codex server request %q", message.Method)
	}
	var call dynamicToolCallParams
	if err := json.Unmarshal(message.Params, &call); err != nil {
		return err
	}
	if call.Namespace != nil {
		return errors.New("codex driver does not yet support namespaced dynamic tools")
	}

	d.mu.Lock()
	s := d.sessionByThreadLocked(call.ThreadID)
	if s == nil || s.phase != phaseRunning || s.turnID != call.TurnID || s.pendingTool != nil {
		d.mu.Unlock()
		return errors.New("codex dynamic tool call does not match a running invocation")
	}
	capability, ok := s.tools[call.Tool]
	if !ok {
		d.mu.Unlock()
		return fmt.Errorf("codex selected unknown dynamic tool %q", call.Tool)
	}
	params, err := argumentsToObservations(call.Arguments)
	if err != nil {
		d.mu.Unlock()
		return err
	}
	normalized, err := json.Marshal([]normalizedOutput{{
		Kind:   "invoke",
		Target: capability.Target,
		Action: capability.Action,
		Params: params,
	}})
	if err != nil {
		d.mu.Unlock()
		return err
	}
	s.pendingTool = &pendingTool{requestID: append(json.RawMessage(nil), message.ID...), callID: call.CallID}
	s.phase = phaseWaitingForChild
	output := takeOutput(s)
	d.mu.Unlock()
	if output == nil {
		return errors.New("codex tool call has no active handoff")
	}
	return output(string(normalized), nil)
}

func (d *Driver) fail(invocation protocol.InvocationID, err error) {
	d.mu.Lock()
	s := d.sessions[invocation]
	if s != nil {
		s.phase = phaseFaulted
	}
	var output func(string, error) error
	if s != nil {
		output = takeOutput(s)
	}
	d.mu.Unlock()
	if output != nil {
		_ = output("", err)
	}
}

func (d *Driver) failAll(err error) {
	d.mu.Lock()
	outputs := make([]func(string, error) error, 0, len(d.sessions))
	for _, s := range d.sessions {
		s.phase = phaseFaulted
		if output := takeOutput(s); output != nil {
			outputs = append(outputs, output)
		}
	}
	d.mu.Unlock()
	for _, output := range outputs {
		_ = output("", err)
	}
}

func (d *Driver) drop(invocation protocol.InvocationID, requestID int64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.pending, requestID)
	delete(d.sessions, invocation)
}

func (d *Driver) allocateLocked(kind requestKind, invocation protocol.InvocationID) int64 {
	id := d.nextID
	d.nextID++
	d.pending[id] = pendingRequest{kind: kind, invocation: invocation}
	return id
}

func (d *Driver) sessionByThreadLocked(threadID string) *session {
	return d.sessions[d.threads[threadID]]
}

func (d *Driver) send(ctx context.Context, value any) error {
	message, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return d.transport.Send(ctx, message)
}

func takeOutput(s *session) func(string, error) error {
	output := s.output
	s.output = nil
	return output
}

func compileTools(focus []protocol.Capability) ([]dynamicTool, map[string]protocol.Capability) {
	tools := make([]dynamicTool, 0, len(focus))
	targets := make(map[string]protocol.Capability, len(focus))
	for i, capability := range focus {
		name := fmt.Sprintf("cap_%03d", i)
		targets[name] = cloneCapability(capability)
		description := strings.TrimSpace(capability.Action.Value)
		if description == "" {
			description = "Invoke focused Skyra capability"
		}
		tools = append(tools, dynamicTool{
			Type:        "function",
			Name:        name,
			Description: description,
			InputSchema: json.RawMessage(genericInputSchema),
		})
	}
	return tools, targets
}

func renderInput(invoke protocol.Invoke) string {
	type renderedObservation struct {
		Name  string          `json:"name,omitempty"`
		Value json.RawMessage `json:"value"`
	}
	type application struct {
		Method     string                `json:"method"`
		Parameters []renderedObservation `json:"parameters"`
	}

	parameters := make([]renderedObservation, 0, len(invoke.Params))
	for _, parameter := range invoke.Params {
		value := append(json.RawMessage(nil), parameter.Content.Bytes...)
		if !json.Valid(value) {
			value, _ = json.Marshal(string(parameter.Content.Bytes))
		}
		parameters = append(parameters, renderedObservation{
			Name:  parameter.Segment.By.Value,
			Value: value,
		})
	}
	rendered, err := json.Marshal(application{Method: invoke.Action.Value, Parameters: parameters})
	if err != nil {
		// The shape above contains only strings and valid RawMessages.
		panic(err)
	}
	return string(rendered)
}

func argumentsToObservations(raw json.RawMessage) ([]protocol.Observation, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return nil, fmt.Errorf("codex tool arguments must be an object: %w", err)
	}
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	observations := make([]protocol.Observation, 0, len(keys))
	for _, key := range keys {
		o := protocol.Observation{}
		o.Segment.By.Value = key
		o.Content.Bytes = append([]byte(nil), object[key]...)
		observations = append(observations, o)
	}
	return observations, nil
}

func singleReturnedObservation(register protocol.Register) (protocol.Observation, error) {
	if len(register.Order) != 1 || len(register.Children) != 1 {
		return protocol.Observation{}, errors.New("codex prototype requires exactly one returned child")
	}
	entry, ok := register.Children[register.Order[0]]
	if !ok || entry.Observation == nil {
		return protocol.Observation{}, errors.New("codex child observation is not ready")
	}
	return cloneObservation(entry.Observation.Observation), nil
}

func cloneObservation(o protocol.Observation) protocol.Observation {
	o.Content.Bytes = append([]byte(nil), o.Content.Bytes...)
	return o
}

func cloneObservations(in []protocol.Observation) []protocol.Observation {
	out := make([]protocol.Observation, len(in))
	for i := range in {
		out[i] = cloneObservation(in[i])
	}
	return out
}

func cloneCapability(c protocol.Capability) protocol.Capability {
	c.Sig = append([]byte(nil), c.Sig...)
	return c
}
