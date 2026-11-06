// A deterministic provider for the Mission tests, speaking the normal tree
// instruction protocol in every role a Mission gives it:
//
//	Plan       one act node holding all its contracts, which it invokes by
//	           number in turn: if it may look around, pwd, then ls of that
//	           directory, then read of the first file listed; then
//	           view_plan_contracts; then propose_spec, proposing the echo tool
//	           from the view, citing the first input and quoting the request
//	           after words of its own, then what a human wrote in a passed
//	           spec; then it returns the spec
//	Launch     echoes its input, and returns what came back
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	var input struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		panic(err)
	}
	prompt := input.Text
	if held(prompt, "propose-spec") >= 0 {
		plan(prompt)
		return
	}
	if back := regexp.MustCompile(`(?m)^  (\d+): .*\(what came back\):?$`).FindStringSubmatch(prompt); back != nil {
		reply(`[{"return":{"members":[` + back[1] + `]}}]`)
		return
	}
	reply(`[{"create_input":{"text":{"name":0}}},{"perform_effect":{"contract":0,"input":{"earlier":0}}}]`)
}

// plan is Plan's act node: each turn, the next of its contracts.
func plan(prompt string) {
	invoke := func(target, input string) {
		reply(fmt.Sprintf(`[{"create_input":%s},{"perform_effect":{"contract":%d,"input":{"earlier":0}}}]`, input, held(prompt, target)))
	}
	if strings.Contains(prompt, "Hello conversational test") {
		if back := regexp.MustCompile(`(?m)^  (\d+): .*\(what came back\):?$`).FindStringSubmatch(prompt); back != nil {
			reply(`[{"return":{"members":[` + back[1] + `]}}]`)
		} else {
			invoke("plan.reply", `{"text":"Hello!"}`)
		}
		return
	}
	if spec := named(prompt, `{"capabilities":`); spec >= 0 {
		reply(`[{"return":{"members":[` + strconv.Itoa(spec) + `]}}]`)
		return
	}
	view := named(prompt, `{"contracts":`)
	switch {
	case view >= 0:
		invoke("propose-spec", propose(prompt, view))
	case held(prompt, "plan.pwd") < 0 || value(prompt, read) >= 0:
		invoke("view-plan-contracts", `{"view_plan_contracts":{}}`)
	case value(prompt, listing) >= 0:
		var list struct {
			Entries []struct{ Name, Kind string } `json:"entries"`
		}
		_ = json.Unmarshal([]byte(line(prompt, listing)), &list)
		file := ""
		for _, e := range list.Entries {
			if e.Kind == "file" && file == "" {
				file = e.Name
			}
		}
		invoke("plan.read", fmt.Sprintf(`{"read":{"dir":{"name":%d},"path":%q}}`, value(prompt, directory), file))
	case value(prompt, directory) >= 0:
		invoke("plan.ls", fmt.Sprintf(`{"ls":{"dir":{"name":%d}}}`, value(prompt, directory)))
	default:
		invoke("plan.pwd", `{"pwd":{}}`)
	}
}

// propose is the proposal: the echo tool, a few words of Plan's own, the
// request quoted, and what a human wrote in a passed spec, under its profile.
func propose(prompt string, view int) string {
	cite := regexp.MustCompile(`(?m)\|\s+([0-9a-f]{64})$`).FindStringSubmatch(prompt)
	if cite == nil {
		panic("Plan was not passed an input to cite")
	}
	// The request reaches Plan as a passed observation in its goal.
	var request string
	if m := regexp.MustCompile(`Passed observation ` + cite[1] + ` from [^\n]*:\n([^\n]+)`).FindStringSubmatch(prompt); m != nil {
		if json.Unmarshal([]byte(m[1]), &request) != nil {
			request = m[1]
		}
	}
	if request == "" {
		panic("Plan was not passed the request to quote")
	}
	goal := []map[string]string{{"text": "Echo the cited request:"}, {"quote": request, "of": cite[1]}}
	// A passed spec: keep what the human wrote in it, by quoting it.
	profile := map[string]int{"act": 1, "turns": 8, "depth": 3, "fanout": 32, "budget": 64}
	numbered := regexp.MustCompile(`(?m)^[ \t]*\d+\| ?`).ReplaceAllString(prompt, "")
	for _, m := range regexp.MustCompile(`Passed observation ([0-9a-f]{64}) from [^\n]*:\n((?:[^\n]+\n?)+)`).FindAllStringSubmatch(numbered, -1) {
		var prior struct {
			Segments []struct{ Text, Source string } `json:"segments"`
			Profile  map[string]int                  `json:"profile"`
		}
		if json.NewDecoder(strings.NewReader(m[2])).Decode(&prior) != nil {
			continue
		}
		if prior.Profile != nil {
			profile = prior.Profile
		}
		for _, seg := range prior.Segments {
			if seg.Source == "human" && len(strings.Fields(seg.Text)) >= 3 {
				goal = append(goal, map[string]string{"quote": seg.Text, "of": m[1]})
			}
		}
	}
	spec, _ := json.Marshal(map[string]any{"propose_spec": map[string]any{"goal": goal, "inputs": []string{cite[1]}, "catalogue": map[string]int{"name": view}, "capabilities": []string{"echo"}, "profile": profile, "note": "I wrote the spec from your message."}})
	return string(spec)
}

// The observations workdir returns, by what their values hold.
var (
	directory = regexp.MustCompile(`(?m)^  (\d+): (\{"dir":"[^"]*"\}) \(`)
	listing   = regexp.MustCompile(`(?m)^  (\d+): (\{"dir":.*"entries":.*\}) \(`)
	read      = regexp.MustCompile(`(?m)^  (\d+): (\{"dir":.*"text":.*\}) \(`)
)

// value is the number of the name whose value pattern matches, or -1.
func value(prompt string, pattern *regexp.Regexp) int {
	if m := pattern.FindStringSubmatch(prompt); m != nil {
		n, _ := strconv.Atoi(m[1])
		return n
	}
	return -1
}

func line(prompt string, pattern *regexp.Regexp) string {
	if m := pattern.FindStringSubmatch(prompt); m != nil {
		return m[2]
	}
	return ""
}

// held is the number of the contract the prompt offers for target, or -1.
func held(prompt, target string) int {
	if m := regexp.MustCompile(`(?m)^  (\d+): ` + regexp.QuoteMeta(target) + `, input`).FindStringSubmatch(prompt); m != nil {
		n, _ := strconv.Atoi(m[1])
		return n
	}
	return -1
}

// named is the number of a name whose value starts with prefix, or -1.
func named(prompt, prefix string) int {
	if m := regexp.MustCompile(`(?m)^  (\d+): ` + regexp.QuoteMeta(prefix)).FindStringSubmatch(prompt); m != nil {
		n, _ := strconv.Atoi(m[1])
		return n
	}
	return -1
}

func reply(instruction string) {
	result, _ := json.Marshal(map[string]any{"to": os.Getenv("SKYRA_FRAME"), "payload": map[string]string{"text": instruction}})
	fmt.Println(string(result))
}
