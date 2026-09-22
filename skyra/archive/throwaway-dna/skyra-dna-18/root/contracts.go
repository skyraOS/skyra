package main

import (
	"context"
	"encoding/json"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"

	p "skyra-v1/src/skyra/skyra-dna-18/protocol"
)

type Bootstrap struct {
	EchoPath     string   `json:"echo"`
	Observations []string `json:"observations"`
}
type RootProgram struct {
	arguments  []string
	boot       Bootstrap
	kernel     *Kernel
	execution  *ExecutionContext
	ingress    []*Ingress
	egress     []*Egress
	client     *endpoint
	results    *endpoint
	helper     *helper
	parentDone <-chan error
	trace      *Trace
}
type Kernel struct {
	runtimeState *RuntimeState
	scheduler    *Scheduler
	trace        *Trace
}
type RuntimeState struct {
	memory *Memory
	active *Active
}
type Memory struct {
	mu          sync.Mutex
	executables map[p.ExecutableID]Executable
	contexts    map[p.ContextID]*ExecutionContext
}
type Active struct {
	mu           sync.RWMutex
	contexts     map[p.ContextID]*ExecutionContext
	reservations map[uint64]p.ContextID
	trace        *Trace
}
type Executable struct {
	ID        p.ExecutableID
	Contracts []ArgContract
}
type ArgContract struct{ Operation string }

// queuedMove keeps authenticated ingress out of executable-controlled JSON.
type queuedMove struct {
	move   p.Move
	source *binding
}
type DispatchMove func(context.Context, queuedMove) error
type Scheduler struct {
	mu        sync.Mutex
	queue     []queuedMove
	realizing map[p.ContextID]bool
	changed   chan struct{}
	trace     *Trace
}
type ExecutionContext struct {
	mu         sync.Mutex
	identity   p.ContextID
	executable Executable
	horizon    map[uint64]*mapping
	bookkeeper *Bookkeeper
	live       bool
	trace      *Trace
}
type mapping struct {
	actual     uint64
	executable Executable
	endpoint   *endpoint
}
type Bookkeeper struct {
	frames   map[p.FrameID]*frame
	next     p.FrameID
	requests map[uint64]bool
	rootSink *Egress
}
type frame struct {
	id      p.FrameID
	request uint64
	parent  p.FrameID
	waiting p.FrameID
	phase   string
	target  p.ExecutableID
	binding *binding
	state   []byte
	result  p.Observation
}

// Each replacement gets a new, never-reused binding object and generation.
// Its pointer is trusted host provenance, never decoded from the wire.
type binding struct {
	owner      p.ContextID
	actual     uint64
	target     p.ExecutableID
	generation uint64
	external   bool
	live       atomic.Bool
}
type endpoint struct {
	ingress    *Ingress
	egress     *Egress
	connection io.ReadWriteCloser
	binding    *binding
	once       sync.Once
}
type Ingress struct {
	endpoint *endpoint
	mu       sync.Mutex
	decoder  *json.Decoder
}
type Egress struct {
	endpoint *endpoint
	mu       sync.Mutex
	encoder  *json.Encoder
}
type helper struct {
	command  *exec.Cmd
	endpoint *endpoint
	waited   bool
}
type duplex struct {
	io.ReadCloser
	writer io.WriteCloser
}
type Trace struct {
	mu     sync.Mutex
	events []Event
	output io.Writer
}
type Event struct {
	Number     uint64      `json:"event"`
	Owner      string      `json:"owner"`
	Context    p.ContextID `json:"context,omitempty"`
	Move       uint64      `json:"move,omitempty"`
	Frame      p.FrameID   `json:"frame,omitempty"`
	Transition string      `json:"transition"`
}
