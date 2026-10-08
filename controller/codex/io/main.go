// Command codexio performs one recorded operation on a resident stream. The
// runtime grants it exactly one path, its edge's socket (SKYRA_RESIDENT), and
// the holder there performs the operation (runtime/hold, proposal 38). Its
// observation is the operation's record; the worker owns no process
// launcher, host socket or side store.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"skyra/controller/runtime/hold"
)

func main() {
	var q hold.Request
	if err := json.NewDecoder(os.Stdin).Decode(&q); err != nil {
		fail(err)
	}
	o, err := perform(q, os.Getenv("SKYRA_RESIDENT"))
	if err != nil {
		fail(err)
	}
	b, _ := json.Marshal(map[string]any{"to": os.Getenv("SKYRA_FRAME"), "payload": o})
	fmt.Println(string(b))
}

// perform checks the request, then asks the holder. A holder that cannot be
// reached is a resident that is gone: nothing was delivered.
func perform(q hold.Request, socket string) (hold.Observation, error) {
	if q.Incarnation == "" || (q.Op != "exchange" && q.Op != "stop" && q.Op != "inspect") || q.WaitMS < 1 || q.WaitMS > 30000 {
		return hold.Observation{}, errors.New("invalid operation")
	}
	if socket == "" {
		return hold.Observation{}, errors.New("no resident granted")
	}
	q.Frame = os.Getenv("SKYRA_OWNER_FRAME")
	return hold.Ask(socket, q), nil
}

func fail(err error) { fmt.Fprintln(os.Stderr, "codexio:", err); os.Exit(1) }
