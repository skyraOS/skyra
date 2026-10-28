// Command service-io is the one adapter in front of the service host
// (proposal 49): it performs one operation on a session and exits. It reaches
// exactly one thing, the service host's socket, which it finds among the
// sockets its Horizon grants (SKYRA_CONNECTIONS) as service.sock in a
// skyra-service-* directory. Its observation is the operation's record; it
// owns no process, no state and no side store.
//
// open starts a session of a registered program and returns its handle. The
// procedure keeps that handle and passes it on every exchange, inspect and
// stop.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"skyra/controller/service"
)

func main() {
	var q service.Request
	if err := json.NewDecoder(os.Stdin).Decode(&q); err != nil {
		fail(err)
	}
	if q.Program != "" && q.Op == "" {
		q.Op = "open"
	}
	socket, err := socket(os.Getenv("SKYRA_CONNECTIONS"))
	if err != nil {
		fail(err)
	}
	o, err := perform(q, socket)
	if err != nil {
		fail(err)
	}
	b, _ := json.Marshal(map[string]any{"to": os.Getenv("SKYRA_FRAME"), "payload": o})
	fmt.Println(string(b))
}

// perform checks the request, then asks the service host. A host that cannot
// be reached is a session that is gone: nothing was delivered.
func perform(q service.Request, socket string) (service.Observation, error) {
	switch {
	case q.Op == "open":
		if q.Program == "" || q.Incarnation != "" || len(q.Deliver) != 0 {
			return service.Observation{}, errors.New("invalid open")
		}
		q.WaitMS = 1
	case q.Incarnation == "" || q.Program != "" || (q.Op != "exchange" && q.Op != "stop" && q.Op != "inspect") || q.WaitMS < 1 || q.WaitMS > 30000:
		return service.Observation{}, errors.New("invalid operation")
	}
	if socket == "" {
		return service.Observation{}, errors.New("no service host granted")
	}
	q.Frame = os.Getenv("SKYRA_FRAME")
	return service.Ask(socket, q), nil
}

// socket is the one service host socket among the granted sockets: none, or
// more than one, is no grant.
func socket(raw string) (string, error) {
	var sockets []string
	if raw != "" && json.Unmarshal([]byte(raw), &sockets) != nil {
		return "", errors.New("malformed connection grant")
	}
	selected := ""
	for _, s := range sockets {
		if filepath.IsAbs(s) && filepath.Base(s) == "service.sock" && strings.HasPrefix(filepath.Base(filepath.Dir(s)), "skyra-service-") {
			if selected != "" {
				return "", errors.New("more than one service host granted")
			}
			selected = s
		}
	}
	return selected, nil
}

func fail(err error) { fmt.Fprintln(os.Stderr, "service-io:", err); os.Exit(1) }
