package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	p "skyra-v1/src/skyra/protocol"
	"time"
)

// serve runs the user executable: framed Moves over in/out to the host, the
// browser over the WebSocket listener. All interaction state is mutated only in
// the single select loop; the browser transport and stdin decoder run in their
// own goroutines and communicate over channels.
func serve(in *os.File, out *os.File, listenAddr string) error {
	b := newBrowser()
	actual, err := b.start(listenAddr)
	if err != nil {
		return err
	}
	defer b.close()
	fmt.Fprintf(os.Stderr, "[user] browser listening on ws://%s/ws\n", actual)

	eng := newEngine(fmt.Sprintf("ui-%d-%d", os.Getpid(), time.Now().UnixMilli()))
	enc := json.NewEncoder(out)

	var uframe p.FrameID
	var ctx p.ContextID
	var cfg = agentConfig{Agent: "echo"}
	var started bool // whether the Codex thread has been started for U
	var seq uint64 = 1000
	emit := func(m p.Move) error { seq++; m.Identity = seq; return enc.Encode(m) }
	push := func(msgs []any) {
		for _, m := range msgs {
			b.send(m)
		}
	}

	moves := make(chan p.Move)
	decodeErr := make(chan error, 1)
	go func() {
		dec := json.NewDecoder(in)
		dec.DisallowUnknownFields()
		for {
			var m p.Move
			if err := dec.Decode(&m); err != nil {
				decodeErr <- err
				return
			}
			moves <- m
		}
	}()

	for {
		select {
		case m := <-moves:
			if err := m.Validate(); err != nil {
				continue
			}
			switch {
			case m.Invoke != nil:
				// The `interact` activation establishes U. It stays alive across
				// agent commands; its only terminal Return is the explicit close.
				if m.Target != p.User || p.CheckArguments(m.Invoke.Args, "interact") != nil {
					continue
				}
				uframe, ctx = m.Invoke.Frame, m.Context
				cfg = decodeAgentConfig(m.Invoke.Args[1].Bytes)
				eng.activate()
				b.send(eng.snapshot())
			case m.Return != nil:
				c, err := decodeContinuation(m.State)
				if err != nil {
					continue
				}
				push(eng.childReturn(c.RequestID, string(m.Return.Observation.Bytes)))
			case m.Fault != nil:
				c, err := decodeContinuation(m.State)
				if err != nil {
					continue
				}
				push(eng.childFault(c.RequestID, m.Fault.Message))
			}

		case raw := <-b.inbound:
			var msg inboundMessage
			if json.Unmarshal(raw, &msg) != nil {
				continue
			}
			switch msg.Type {
			case "submit":
				out, child := eng.submit(msg)
				push(out)
				if child != nil && uframe != 0 {
					if err := emit(agentInvoke(ctx, uframe, eng.conversation, cfg, &started, child)); err != nil {
						return err
					}
				}
			case "close-interaction":
				out, terminal := eng.close(msg)
				push(out)
				if terminal && uframe != 0 {
					if err := emit(terminalReturn(ctx, uframe)); err != nil {
						return err
					}
					uframe = 0 // U settled; no further child work is authorized
				}
			}

		case <-b.connects:
			b.send(eng.snapshot())

		case err := <-decodeErr:
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
	}
}

// agentInvoke builds the child Invoke that carries a human message to the agent.
// For echo it is a one-shot echo call; for Codex it is the persistent
// conversation — the first message starts the thread, later ones submit input
// on the same host-minted parent. The isolation, contamination fence and
// tool-child routing are the runtime's and Codex's, unchanged.
func agentInvoke(ctx p.ContextID, uframe p.FrameID, conversation string, cfg agentConfig, started *bool, child *childRequest) p.Move {
	target, args := p.Echo, p.Arguments("echo", child.text)
	if cfg.Agent == "codex" {
		target = p.Codex
		tools := cfg.Tools
		if tools == "" {
			tools = "[]"
		}
		if *started {
			args = p.Args("submit-input", "parent", cfg.Parent, "text", child.text)
		} else {
			args = p.Args("start-thread", "connection", cfg.Connection, "parent", cfg.Parent, "text", child.text, "tools", tools)
			*started = true
		}
	}
	return p.Move{
		Context: ctx,
		Source:  uframe,
		Target:  target,
		Invoke:  &p.Invoke{Args: args},
		State:   encodeContinuation(continuation{Conversation: conversation, RequestID: child.requestID}),
	}
}

func terminalReturn(ctx p.ContextID, uframe p.FrameID) p.Move {
	return p.Move{
		Context: ctx,
		Source:  uframe,
		Return:  &p.Return{Answering: uframe, Observation: p.Observation{Key: "text", Bytes: []byte("interaction-closed")}},
	}
}

func main() {
	addr := os.Getenv("SKYRA_USER_ADDR")
	if addr == "" {
		addr = "127.0.0.1:0"
	}
	if err := serve(os.Stdin, os.Stdout, addr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
