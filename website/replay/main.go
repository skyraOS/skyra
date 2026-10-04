//go:build js && wasm

// Command replay is the controller's replay, compiled to WebAssembly for the
// website. It exposes one function to the page:
//
//	skyraVerify(bundle) → result
//
// where bundle is a run as website/export writes it (possibly edited by the
// visitor) and result is JSON: the chain from genesis to the head, each record
// replayed from genesis by the same code that checks runs on the Mac.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"syscall/js"

	"skyra/controller/core/record"
	"skyra/controller/runtime/replay"
)

type bundle struct {
	Head    string            `json:"head"`
	Objects map[string]string `json:"objects"`
}

type objects map[string]string

func (o objects) Get(kind, h string) ([]byte, error) {
	s, ok := o[kind+"/"+h]
	if !ok {
		return nil, fmt.Errorf("%s/%s: %w", kind, record.Short(h), os.ErrNotExist)
	}
	return []byte(s), nil
}

// Step is one record of the chain, and whether replay reaches it.
type Step struct {
	Stamp    string `json:"stamp"`
	Sum      string `json:"sum"` // what its stored bytes hash to now
	State    string `json:"state"`
	To       string `json:"to,omitempty"`
	Input    string `json:"input,omitempty"`
	Describe string `json:"describe"`
	OK       bool   `json:"ok"`
	Error    string `json:"error,omitempty"`
}

type result struct {
	Head    string `json:"head"`
	Chain   []Step `json:"chain"`
	Records int    `json:"records"` // includes the siblings checked at resolves
	OK      bool   `json:"ok"`      // the head replays
	Error   string `json:"error"`   // why it does not, or why the chain could not be read
}

func verify(raw string) result {
	var b bundle
	if err := json.Unmarshal([]byte(raw), &b); err != nil {
		return result{Error: "the bundle is not JSON: " + err.Error()}
	}
	o := objects(b.Objects)
	res := result{Head: b.Head}

	// The chain as the stored records state it, head back to genesis. It is
	// only read here; whether any of it holds is replay's to say.
	var chain []Step
	for s := b.Head; s != ""; {
		raw, err := o.Get("records", s)
		if err != nil {
			res.Error = err.Error()
			break
		}
		var r record.Record
		if err := json.Unmarshal(raw, &r); err != nil {
			chain = append([]Step{{Stamp: s, Sum: record.Sum(raw), Describe: "unreadable record"}}, chain...)
			res.Error = fmt.Sprintf("record %s is not a record: %v", record.Short(s), err)
			break
		}
		chain = append([]Step{{Stamp: s, Sum: record.Sum(raw), State: string(r.State), To: r.To, Input: r.Input, Describe: record.Describe(r)}}, chain...)
		if len(chain) > 10000 {
			res.Error = "the chain does not end"
			break
		}
		s = r.Predecessor
	}

	// Each record replayed from genesis: every one before a change still
	// replays, and none after it does.
	for i := range chain {
		_, err := replay.Lineage(o, chain[i].Stamp)
		chain[i].OK = err == nil
		if err != nil {
			chain[i].Error = err.Error()
		}
	}
	res.Chain = chain
	if res.Error == "" {
		n, err := replay.Lineage(o, b.Head)
		res.Records = n
		res.OK = err == nil
		if err != nil {
			res.Error = err.Error()
		}
	}
	return res
}

func main() {
	js.Global().Set("skyraVerify", js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) != 1 || args[0].Type() != js.TypeString {
			return js.ValueOf(`{"error":"skyraVerify takes the bundle as a string"}`)
		}
		out, err := json.Marshal(verify(args[0].String()))
		if err != nil {
			out, _ = json.Marshal(result{Error: err.Error()})
		}
		return js.ValueOf(string(out))
	}))
	select {}
}
