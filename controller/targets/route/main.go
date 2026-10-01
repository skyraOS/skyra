// Command route is the route runner (next/30): a runtime-owned surface that
// runs one route, a Starlark program, once over one return, and decides. It
// runs under an empty Horizon: no reach, no grants, the floor and its own exec
// only. It is its own module, so the controller stays standard-library Go.
//
// It reads, on stdin:
//
//	{"source": "<starlark>", "return": <json>, "previous": <json or null>, "targets": ["name", ...]}
//
// and returns to its slot (SKYRA_FRAME) with exactly one decision:
//
//	{"interrupt": true} | {"send": {"to": "<name>", "payload": <json>}} | {"fault": "<cause>"}
//
// A route is deterministic: no files, network, clock or randomness, and the
// previous return is supplied by the runtime. The same input gives the same
// bytes out, so replay can run it again.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func main() {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "route:", err)
		os.Exit(1)
	}
	var in Input
	d := Decision{Fault: "route: the runner's input is not readable"}
	if json.Unmarshal(b, &in) == nil {
		d = Run(in)
	}
	out, _ := json.Marshal(map[string]any{"to": os.Getenv("SKYRA_FRAME"), "payload": d})
	fmt.Println(string(out))
}
