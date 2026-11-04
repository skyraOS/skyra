// Command treemodel is a scripted model for the apprentice's trees (56,
// step 5): a leaf that reads a turn's prompt as a model would and answers
// with instructions, so a tree runs through the kernel with no provider. An
// act node invokes contract 0 with {"text":"hello"}; a delegate node
// delegates its goal with contract 0; either returns what came back.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var line = regexp.MustCompile(`(?m)^  (\d+): .*\((.*)\)$`)

func main() {
	var in struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fmt.Fprintln(os.Stderr, "treemodel:", err)
		os.Exit(1)
	}
	out := `[{"goal":"say hello through echo"},{"delegate":{"observations":[{"earlier":0}],"contracts":[0]}}]`
	if strings.Contains(in.Text, "an act node") {
		out = `[{"create_input":"{\"text\":\"hello\"}"},{"perform_effect":{"contract":0,"input":{"earlier":0}}}]`
	}
	for _, m := range line.FindAllStringSubmatch(in.Text, -1) {
		if m[2] == "what came back" || strings.HasSuffix(m[2], " returned") {
			out = `[{"return":{"members":[` + m[1] + `]}}]`
		}
	}
	b, _ := json.Marshal(map[string]any{"to": os.Getenv("SKYRA_FRAME"), "payload": map[string]string{"text": out}})
	fmt.Println(string(b))
}
