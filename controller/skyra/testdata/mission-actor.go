// A deterministic provider fixture speaking the normal tree instruction protocol.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

func main() {
	var input struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		panic(err)
	}
	instruction := `[{"create_input":{"text":{"name":0}}},{"perform_effect":{"contract":0,"input":{"earlier":0}}}]`
	for _, match := range regexp.MustCompile(`(?m)^  (\d+): .*\(what came back\):?$`).FindAllStringSubmatch(input.Text, -1) {
		instruction = `[{"return":{"members":[` + match[1] + `]}}]`
	}
	result, _ := json.Marshal(map[string]any{"to": os.Getenv("SKYRA_FRAME"), "payload": map[string]string{"text": instruction}})
	fmt.Println(string(result))
}
