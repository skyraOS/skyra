// propose-spec is an ordinary leaf contract. Its output is a proposal, never
// a capability grant. The trusted Mission bridge validates it independently.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"skyra/controller/core/contract"
	"skyra/controller/core/mission"
)

func propose(raw []byte) (json.RawMessage, error) {
	value, err := contract.Decode(raw)
	if err != nil {
		return nil, err
	}
	if err = contract.Fits(mission.ProposeSchema(), value); err != nil {
		return nil, err
	}
	var input struct {
		Spec json.RawMessage `json:"propose_spec"`
	}
	if err = json.Unmarshal(raw, &input); err != nil {
		return nil, err
	}
	spec, err := mission.DecodeSpec(input.Spec)
	if err != nil {
		return nil, err
	}
	return json.Marshal(spec)
}
func main() {
	raw, err := io.ReadAll(io.LimitReader(os.Stdin, 1024*1024+1))
	if err == nil && len(raw) > 1024*1024 {
		err = fmt.Errorf("proposal exceeds 1 MiB")
	}
	var value json.RawMessage
	if err == nil {
		value, err = propose(raw)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "propose-spec:", err)
		os.Exit(1)
	}
	output, _ := json.Marshal(map[string]any{"to": os.Getenv("SKYRA_FRAME"), "payload": value})
	fmt.Println(string(output))
}
