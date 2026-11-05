// A deterministic provider for the Mission tests, speaking the normal tree
// instruction protocol in every role a Mission gives it:
//
//	Plan's delegate root  if it may look around: cuts "Find" for a pwd child,
//	                      "List" for an ls child passed that directory, "Read"
//	                      for a read child passed it and the listing; then
//	                      "View" for a viewer child, then "Propose…" for a
//	                      proposer child passed the view; then returns the spec
//	pwd, ls, read         find, list, and read the first file listed
//	the viewer            invokes view_plan_contracts
//	the proposer          proposes the echo tool from the view, citing the first
//	                      input and quoting the request after words of its own,
//	                      then what a human wrote in a passed spec
//	Launch                echoes its input
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
	if back := regexp.MustCompile(`(?m)^  (\d+): .*\(what came back\):?$`).FindStringSubmatch(prompt); back != nil {
		reply(`[{"return":{"members":[` + back[1] + `]}}]`)
		return
	}
	if strings.Contains(prompt, "a delegate node") {
		delegate(prompt)
		return
	}
	switch {
	case held(prompt, "plan.pwd") >= 0:
		reply(`[{"create_input":{"pwd":{}}},{"perform_effect":{"contract":0,"input":{"earlier":0}}}]`)
	case held(prompt, "plan.ls") >= 0:
		reply(fmt.Sprintf(`[{"create_input":{"ls":{"dir":{"name":%d}}}},{"perform_effect":{"contract":0,"input":{"earlier":0}}}]`, value(prompt, directory)))
	case held(prompt, "plan.read") >= 0:
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
		reply(fmt.Sprintf(`[{"create_input":{"read":{"dir":{"name":%d},"path":%q}}},{"perform_effect":{"contract":0,"input":{"earlier":0}}}]`, value(prompt, directory), file))
	case held(prompt, "view-plan-contracts") >= 0:
		reply(`[{"create_input":{"view_plan_contracts":{}}},{"perform_effect":{"contract":0,"input":{"earlier":0}}}]`)
	case held(prompt, "propose-spec") >= 0:
		cite := regexp.MustCompile(`(?m)\|\s+([0-9a-f]{64})$`).FindStringSubmatch(prompt)
		view := named(prompt, `{"contracts":`)
		if cite == nil || view < 0 {
			panic("the proposer was not passed the view and an input to cite")
		}
		// The goal adds a few words of its own, then quotes the request.
		// The request reaches Plan as a passed observation in its goal.
		var request string
		if m := regexp.MustCompile(`Passed observation ` + cite[1] + ` from [^\n]*:\n([^\n]+)`).FindStringSubmatch(prompt); m != nil {
			if json.Unmarshal([]byte(m[1]), &request) != nil {
				request = m[1]
			}
		}
		if request == "" {
			panic("the proposer was not passed the request to quote")
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
		reply(`[{"create_input":` + string(spec) + `},{"perform_effect":{"contract":0,"input":{"earlier":0}}}]`)
	default:
		reply(`[{"create_input":{"text":{"name":0}}},{"perform_effect":{"contract":0,"input":{"earlier":0}}}]`)
	}
}

func delegate(prompt string) {
	if spec := named(prompt, `{"capabilities":`); spec >= 0 {
		reply(`[{"return":{"members":[` + strconv.Itoa(spec) + `]}}]`)
		return
	}
	lines := goalLines(prompt)
	explore := held(prompt, "plan.pwd") >= 0
	if view := named(prompt, `{"contracts":`); view < 0 && explore && value(prompt, read) < 0 {
		var step, contract string
		passed := ""
		switch {
		case value(prompt, listing) >= 0:
			step, contract, passed = "Read a file", "plan.read", fmt.Sprintf(",%d,%d", value(prompt, directory), value(prompt, listing))
		case value(prompt, directory) >= 0:
			step, contract, passed = "List a directory", "plan.ls", fmt.Sprintf(",%d", value(prompt, directory))
		default:
			step, contract = "Find where you are", "plan.pwd"
		}
		at := find(lines, step)
		reply(fmt.Sprintf(`[{"goal":{"of":0,"lines":[%d,%d]}},{"delegate":{"observations":[{"earlier":0}%s],"contracts":[%d]}}]`, at, at, passed, held(prompt, contract)))
		return
	}
	if view := named(prompt, `{"contracts":`); view >= 0 {
		from := find(lines, "Propose the spec")
		reply(fmt.Sprintf(`[{"goal":{"of":0,"lines":[%d,%d]}},{"delegate":{"observations":[{"earlier":0},%d],"contracts":[%d]}}]`, from, len(lines), view, held(prompt, "propose-spec")))
		return
	}
	at := find(lines, "View the capabilities")
	reply(fmt.Sprintf(`[{"goal":{"of":0,"lines":[%d,%d]}},{"delegate":{"observations":[{"earlier":0}],"contracts":[%d]}}]`, at, at, held(prompt, "view-plan-contracts")))
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

// goalLines are name 0's lines, as the prompt numbers them.
func goalLines(prompt string) []string {
	var out []string
	in := false
	for _, l := range strings.Split(prompt, "\n") {
		if strings.HasPrefix(l, "  0: text of lines") {
			in = true
			continue
		}
		if in {
			m := regexp.MustCompile(`^      \d+\| (.*)$`).FindStringSubmatch(l)
			if m == nil {
				break
			}
			out = append(out, m[1])
		}
	}
	return out
}

func find(lines []string, prefix string) int {
	for i, l := range lines {
		if strings.HasPrefix(l, prefix) {
			return i + 1
		}
	}
	panic("no line " + prefix)
}

func reply(instruction string) {
	result, _ := json.Marshal(map[string]any{"to": os.Getenv("SKYRA_FRAME"), "payload": map[string]string{"text": instruction}})
	fmt.Println(string(result))
}
