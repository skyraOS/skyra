// propose-spec is an ordinary leaf contract. Its input names the tools a Spec
// proposes from the contracts view_plan_contracts returned; its output is that Spec,
// with their references. A proposal is never a capability grant: the trusted
// Mission bridge validates it independently.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"skyra/controller/core/mission"
)

func propose(raw []byte) (json.RawMessage, error) {
	spec, err := mission.Propose(raw)
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
