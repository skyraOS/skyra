// Package creation provides the one-purpose capability peer and minted echo peer.
package creation

import (
	"encoding/json"
	"fmt"
	"io"
	p "skyra-v1/src/skyra/skyra-dna-18.5/protocol"
	"strconv"
)

const Capability p.ExecutableID = 200

// Mint is trusted mediation scoped at wiring time; no caller-supplied owner or
// virtual/actual port can reach it. The request names only the grant template.
type Mint func(p.ExecutableID) (p.ExecutableID, error)

func Serve(connection io.ReadWriteCloser, mint Mint) error {
	defer connection.Close()
	decoder, encoder := json.NewDecoder(connection), json.NewEncoder(connection)
	for {
		var m p.Move
		if err := decoder.Decode(&m); err != nil {
			return err
		}
		if err := m.Validate(); err != nil {
			return err
		}
		if m.Target != Capability || m.Invoke == nil {
			return fmt.Errorf("expected creation invocation")
		}
		if err := p.CheckArguments(m.Invoke.Args, "create"); err != nil {
			return err
		}
		template, err := strconv.ParseUint(string(m.Invoke.Args[1].Bytes), 10, 64)
		var id p.ExecutableID
		if err == nil {
			id, err = mint(p.ExecutableID(template))
		}
		out := p.Move{Identity: m.Identity, Context: m.Context, Source: m.Invoke.Frame}
		if err != nil {
			out.Fault = &p.Fault{Frame: m.Invoke.Frame, Message: err.Error()}
		} else {
			out.Return = &p.Return{Answering: m.Invoke.Frame, Observation: p.Observation{Key: "executable", Bytes: []byte(strconv.FormatUint(uint64(id), 10))}}
		}
		if err := encoder.Encode(out); err != nil {
			return err
		}
	}
}

func ServeEcho(connection io.ReadWriteCloser, id p.ExecutableID) error {
	defer connection.Close()
	decoder, encoder := json.NewDecoder(connection), json.NewEncoder(connection)
	for {
		var m p.Move
		if err := decoder.Decode(&m); err != nil {
			return err
		}
		if err := m.Validate(); err != nil {
			return err
		}
		if m.Target != id || m.Invoke == nil {
			return fmt.Errorf("expected minted echo invocation")
		}
		if err := p.CheckArguments(m.Invoke.Args, "echo"); err != nil {
			return err
		}
		out := p.Move{Identity: m.Identity, Context: m.Context, Source: m.Invoke.Frame, Return: &p.Return{Answering: m.Invoke.Frame, Observation: m.Invoke.Args[1]}}
		if err := encoder.Encode(out); err != nil {
			return err
		}
	}
}
