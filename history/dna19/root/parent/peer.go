package parent

import (
	"encoding/json"
	"fmt"
	"io"
	p "skyra/history/dna19/protocol"
)

// Serve yields without retaining a host-language continuation. All resume state
// is serialized back to the owning Bookkeeper in the child Invoke.
func Serve(connection io.ReadWriteCloser) error {
	defer connection.Close()
	decoder, encoder := json.NewDecoder(connection), json.NewEncoder(connection)
	for {
		var m p.Move
		if err := decoder.Decode(&m); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if err := m.Validate(); err != nil {
			return err
		}
		out := p.Move{Identity: m.Identity, Context: m.Context}
		switch {
		case m.Invoke != nil:
			if m.Target != p.Parent {
				return fmt.Errorf("wrong parent target")
			}
			if err := p.CheckArguments(m.Invoke.Args, "parent"); err != nil {
				return err
			}
			out.Target, out.Source = p.Echo, m.Invoke.Frame
			out.Invoke = &p.Invoke{Args: p.Arguments("echo", string(m.Invoke.Args[1].Bytes))}
			out.State, _ = json.Marshal(Continuation{Prefix: "parent:"})
		case m.Return != nil:
			var state Continuation
			if err := json.Unmarshal(m.State, &state); err != nil {
				return err
			}
			out.Source = m.Source
			out.Return = &p.Return{Answering: m.Source, Observation: p.Observation{Key: "text", Bytes: []byte(state.Prefix + string(m.Return.Observation.Bytes))}}
		default:
			return fmt.Errorf("parent expects invocation or child return")
		}
		if err := encoder.Encode(out); err != nil {
			return err
		}
	}
}
