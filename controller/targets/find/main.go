// Command find delegates bounded searching to a filesystem-only engine.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"skyra/controller/targets/find/search"
	"time"
)

type input = search.Input
type result = search.Result

func find(root, scratch string, in input, now time.Time) result {
	return search.Find(root, scratch, in, now)
}
func main() {
	var in input
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fmt.Fprintln(os.Stderr, "find:", err)
		os.Exit(1)
	}
	root := os.Getenv("SKYRA_WORKSPACE")
	out, _ := json.Marshal(map[string]any{"to": os.Getenv("SKYRA_FRAME"), "payload": find(root, os.Getenv("SKYRA_SCRATCH"), in, time.Now())})
	fmt.Println(string(out))
}
