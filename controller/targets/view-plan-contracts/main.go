// view-plan-contracts is an ordinary leaf contract: what it returns is the
// contracts a Mission's human holds that a Launch may be granted. The host
// writes them, as the bridge's View, into this target's own state before the
// tree boots; a model only asks. Asked twice, it says the same.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"skyra/controller/core/contract"
	"skyra/controller/core/mission"
)

// File is where, in its state, the host leaves the view.
const File = "contracts.json"

func view(state string, raw []byte) (json.RawMessage, error) {
	v, err := contract.Decode(raw)
	if err != nil {
		return nil, err
	}
	if err = contract.Fits(mission.ViewSchema(), v); err != nil {
		return nil, err
	}
	if state == "" {
		return nil, errors.New("no state of its own (SKYRA_STATE is empty)")
	}
	b, err := os.ReadFile(filepath.Join(state, File))
	if err != nil {
		return nil, fmt.Errorf("the host left no contracts: %w", err)
	}
	var held mission.View
	if err = json.Unmarshal(b, &held); err != nil {
		return nil, fmt.Errorf("the host's contracts are not a view: %w", err)
	}
	return json.Marshal(held)
}

func main() {
	raw, err := io.ReadAll(io.LimitReader(os.Stdin, 64*1024))
	var value json.RawMessage
	if err == nil {
		value, err = view(os.Getenv("SKYRA_STATE"), raw)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "view-plan-contracts:", err)
		os.Exit(1)
	}
	output, _ := json.Marshal(map[string]any{"to": os.Getenv("SKYRA_FRAME"), "payload": value})
	fmt.Println(string(output))
}
