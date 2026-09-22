package main

import (
	"encoding/json"
	"os/exec"
	p "skyra-v1/src/skyra/skyra-dna-19/protocol"
	"time"
)

// All provider vocabulary and mutable session state are private to this process.
type rpc struct {
	ID     json.RawMessage `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *rpcError       `json:"error,omitempty"`
}
type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
type session struct {
	forkSource   *session
	locked       bool
	seenCalls    map[string]bool
	ref          p.ParentRef
	thread, turn string
	tools        []string
	history      []historyEntry
	work         *p.Move
	pending      *dependency
	retired      bool
	final        string
	deadline     time.Time
}
type historyEntry struct {
	Kind string `json:"kind"`
	Text string `json:"text"`
}
type dependency struct {
	id          json.RawMessage
	call, nonce string
}
type request struct {
	kind     string
	move     p.Move
	session  *session
	tools    []string
	deadline time.Time
}
type provider struct {
	command *exec.Cmd
	send    chan any
	events  chan rpc
	errors  chan error
	close   func()
}
type actor struct {
	seenCommands    map[string]bool
	usedThreads     map[string]bool
	descriptor      p.Descriptor
	digest          string
	key             []byte
	namespace       string
	providerPath    string
	providerArgs    []string
	provider        *provider
	sessions        map[string]*session
	threads         map[string]*session
	requests        map[string]request
	commands        map[string]*session
	sequence        uint64
	connection      string
	connectionOwner p.ContextID
	ready           bool
	failed          bool
	emit            func(p.Move) error
}
