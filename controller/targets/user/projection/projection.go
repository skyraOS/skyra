// Package projection formats Boundary-supplied conversation work for the
// existing Skyra.app snapshot protocol. It performs no I/O or provider work.
package projection

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

type Record struct {
	Predecessor, State, To, Frame, Round string
	Input                                string
	Payload                              json.RawMessage
	Fault                                *Fault
	Deps                                 []string
	Supervisor                           bool
	Reach                                map[string]json.RawMessage
	Authority                            json.RawMessage
	Author, Nonce                        string
}
type Fault struct{ Kind, Cause string }
type Boundary struct {
	Stamp     string                     `json:"stamp"`
	Surface   json.RawMessage            `json:"surface"`
	Frame     string                     `json:"frame"`
	Name      string                     `json:"name"`
	Origin    string                     `json:"origin"`
	Reach     map[string]json.RawMessage `json:"reach"`
	Authority json.RawMessage            `json:"authority"`
	Refs      map[string]struct {
		Actor     string
		Reach     []string
		Authority json.RawMessage
	} `json:"refs"`
	Forked bool      `json:"forked"`
	Resume *Boundary `json:"resume"`
}
type Tree struct {
	Version    int                 `json:"version"`
	Root       string              `json:"root"`
	Frontier   []string            `json:"frontier"`
	Records    map[string]Record   `json:"records"`
	Order      []string            `json:"order"`
	Boundaries map[string]Boundary `json:"boundaries"`
	Opaque     map[string]string   `json:"opaque"`
	Covered    map[string]string   `json:"covered"`
	Inputs     map[string][]byte   `json:"inputs"`
}
type Availability struct {
	InputLive        bool `json:"input_live"`
	TurnLive         bool `json:"turn_live"`
	RecoveryRequired bool `json:"recovery_required"`
}
type Call struct {
	To     string `json:"to"`
	Arg    string `json:"arg"`
	State  string `json:"state"`
	Result string `json:"result,omitempty"`
	Fault  string `json:"fault,omitempty"`
	Cause  string `json:"cause,omitempty"`
	Stamp  string `json:"stamp"`
}
type Entry struct {
	Role      string  `json:"role"`
	Text      string  `json:"text"`
	TS        int     `json:"ts"`
	RequestID *string `json:"requestId"`
	Call      *Call   `json:"call,omitempty"`
}
type SurfaceView struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}
type BoundaryView struct {
	ID     string        `json:"id"`
	Root   bool          `json:"root"`
	Actors []string      `json:"actors"`
	Reach  []SurfaceView `json:"reach"`
	Grants int           `json:"grants"`
}
type Conversation struct {
	Conversation     string         `json:"conversation"`
	Phase            string         `json:"phase"`
	Entries          []Entry        `json:"entries"`
	Boundaries       []BoundaryView `json:"boundaries"`
	PendingRequestID *string        `json:"pendingRequestId"`
}
type Request struct {
	RequestID string `json:"requestId"`
	Status    string `json:"status"`
	TS        int    `json:"ts"`
	Code      string `json:"code,omitempty"`
	Message   string `json:"message,omitempty"`
}
type View struct {
	Conversation Conversation `json:"conversation"`
	Requests     []Request    `json:"requests"`
}
type Snapshot struct {
	Type          string         `json:"type"`
	UIIncarnation string         `json:"uiIncarnation"`
	Revision      int            `json:"revision"`
	Conversations []Conversation `json:"conversations"`
	Requests      []Request      `json:"requests"`
}

// ConversationState keeps child completion distinct from parent consumption.
// In particular only a joined outcome completes a request. No descriptor
// liveness is inferred from historical records.
func ConversationState(id string, root string, tr Tree, availability Availability) (View, error) {
	out := View{Conversation: Conversation{Conversation: id, Phase: "closed", Entries: []Entry{}, Boundaries: []BoundaryView{}}, Requests: []Request{}}
	if id == "" || tr.Version != 1 || tr.Root != root || len(tr.Frontier) == 0 {
		return out, fmt.Errorf("invalid retained conversation tree")
	}
	if _, ok := tr.Records[root]; !ok {
		return out, fmt.Errorf("tree omits conversation root")
	}
	if availability.InputLive && availability.TurnLive || availability.RecoveryRequired && (availability.InputLive || availability.TurnLive) {
		return out, fmt.Errorf("incompatible continuation availability")
	}
	if availability.RecoveryRequired {
		out.Conversation.Phase = "failed"
	} else if availability.TurnLive {
		out.Conversation.Phase = "pending"
	} else if availability.InputLive {
		out.Conversation.Phase = "ready"
	}
	consumed := map[string]bool{}
	for _, r := range tr.Records {
		if r.State == "resolve" {
			for _, d := range r.Deps {
				consumed[d] = true
			}
		}
	}
	toolIndex := map[string]int{}
	requestIndex := map[string]int{}
	pending := ""
	for _, s := range tr.Order {
		r, ok := tr.Records[s]
		if !ok {
			return out, fmt.Errorf("tree order names missing work")
		}
		h := tr.Boundaries[s]
		frame := h.Frame
		if call, ok := tr.Opaque[s]; ok {
			frame = call
		}
		switch r.State {
		case "invoke":
			if r.To == "browser" {
				continue
			}
			if r.To == "codex" || r.To == "codex.turn" || has(r.Payload, "conversation.turn") {
				continue
			}
			if !modelTool(tr, s, r.To) {
				continue
			}
			c := Call{To: r.To, Arg: argument(r.Payload), State: "running", Stamp: s}
			toolIndex[s] = len(out.Conversation.Entries)
			out.Conversation.Entries = append(out.Conversation.Entries, Entry{Role: "call", Text: strings.TrimSpace(c.To + " " + c.Arg), Call: &c})
		case "return", "fault":
			inv := tr.Records[frame]
			if r.Fault != nil && !r.Supervisor {
				var attempted struct {
					To      string          `json:"to"`
					Payload json.RawMessage `json:"payload"`
				}
				if json.Unmarshal(tr.Inputs[s], &attempted) == nil && modelTool(tr, s, attempted.To) && !strings.HasPrefix(r.Fault.Cause, "return: ") {
					c := Call{To: attempted.To, Arg: argument(attempted.Payload), State: "fault", Fault: r.Fault.Kind, Cause: r.Fault.Cause, Stamp: s}
					out.Conversation.Entries = append(out.Conversation.Entries, Entry{Role: "call", Text: strings.TrimSpace(c.To + " " + c.Arg), Call: &c})
					if pending != "" && consumed[s] && (inv.To == "codex" || inv.To == "codex.turn" || has(inv.Payload, "conversation.turn")) {
						out.Requests[requestIndex[pending]].Status = "failed"
						pending = ""
					}
					continue
				}
			}
			if frame == root {
				out.Conversation.Phase = "closed"
				continue
			}
			if inv.To == "browser" && r.State == "return" {
				// Legacy text and structured accepted actions use the same projection.
				var text string
				var action struct {
					Text, RequestID string
					Operation       string
				}
				if json.Unmarshal(r.Payload, &text) != nil {
					if json.Unmarshal(r.Payload, &action) != nil {
						return out, fmt.Errorf("invalid accepted human action")
					}
					text = action.Text
				}
				if text == "/close" || action.Operation == "close" {
					continue
				}
				e := Entry{Role: "user", Text: text}
				if action.RequestID != "" {
					rid := action.RequestID
					e.RequestID = &rid
					pending = rid
					requestIndex[rid] = len(out.Requests)
					out.Requests = append(out.Requests, Request{RequestID: rid, Status: "accepted", TS: len(out.Conversation.Entries)})
				}
				out.Conversation.Entries = append(out.Conversation.Entries, e)
				continue
			}
			if i, ok := toolIndex[frame]; ok {
				c := out.Conversation.Entries[i].Call
				if r.Fault != nil {
					c.State = "fault"
					c.Fault = r.Fault.Kind
					c.Cause = r.Fault.Cause
				} else {
					c.State = "returned"
					c.Result = summarize(r.Payload)
				}
				continue
			}
			if inv.To == "codex" || inv.To == "codex.turn" || has(inv.Payload, "conversation.turn") {
				text := ""
				status := "returned"
				if r.Fault != nil {
					status = "failed"
					text = fmt.Sprintf("⚠ the agent faulted (%s): %s", r.Fault.Kind, r.Fault.Cause)
				} else {
					var answer struct{ Text, Refused string }
					if json.Unmarshal(r.Payload, &answer) == nil {
						text = answer.Text
						if answer.Refused != "" {
							text = "⚠ the agent attempted " + answer.Refused + "; the turn was interrupted"
						}
					} else {
						_ = json.Unmarshal(r.Payload, &text)
					}
				}
				e := Entry{Role: "agent", Text: text}
				if pending != "" {
					rid := pending
					e.RequestID = &rid
					if consumed[s] {
						i := requestIndex[rid]
						out.Requests[i].Status = status
						pending = ""
					}
				}
				out.Conversation.Entries = append(out.Conversation.Entries, e)
			}
		}
	}
	if pending != "" {
		out.Conversation.PendingRequestID = &pending
	}
	for i := range out.Conversation.Entries {
		out.Conversation.Entries[i].TS = i
	}
	out.Conversation.Boundaries = views(tr)
	return out, nil
}

func modelTool(tr Tree, stamp, to string) bool {
	if to == "" || to == "user" || to == "browser" || to == "interpreter" || to == "system" || strings.HasPrefix(to, "codex.") || strings.HasPrefix(to, "system.") {
		return false
	}
	for s := tr.Records[stamp].Predecessor; s != ""; {
		r, ok := tr.Records[s]
		if !ok {
			return false
		}
		if r.State == "invoke" && (r.To == "codex" || r.To == "codex.turn" || has(r.Payload, "conversation.turn")) {
			return true
		}
		s = r.Predecessor
	}
	return false
}
func has(raw json.RawMessage, key string) bool {
	var m map[string]json.RawMessage
	if json.Unmarshal(raw, &m) != nil {
		return false
	}
	_, ok := m[key]
	return ok
}
func argument(raw json.RawMessage) string {
	var m map[string]json.RawMessage
	if json.Unmarshal(raw, &m) == nil {
		for _, k := range []string{"path", "text", "command"} {
			var s string
			if json.Unmarshal(m[k], &s) == nil && s != "" {
				return clip(s)
			}
		}
		if len(m) == 0 {
			return ""
		}
	}
	return clip(string(raw))
}
func clip(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	rs := []rune(s)
	if len(rs) > 60 {
		return string(rs[:59]) + "…"
	}
	return s
}
func summarize(raw json.RawMessage) string {
	var r struct{ Error string }
	if json.Unmarshal(raw, &r) == nil && r.Error != "" {
		return "error: " + r.Error
	}
	if len(raw) < 1000 {
		return fmt.Sprintf("returned %d B", len(raw))
	}
	return fmt.Sprintf("returned %.1f kB", float64(len(raw))/1000)
}
func hash(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func compact(raw json.RawMessage) []byte {
	var b bytes.Buffer
	if json.Compact(&b, raw) != nil {
		return nil
	}
	return b.Bytes()
}
func views(tr Tree) []BoundaryView {
	all := map[string]*BoundaryView{}
	rootID := ""
	add := func(h Boundary, actor string, isRoot bool) {
		ids := map[string]string{}
		for name, s := range h.Reach {
			ids[name] = hash(compact(s))
		}
		b, _ := json.Marshal(map[string]any{"reach": ids, "authority": h.Authority})
		id := hash(b)
		v := all[id]
		if v == nil {
			v = &BoundaryView{ID: id, Actors: []string{}, Reach: []SurfaceView{}}
			for name, identity := range ids {
				v.Reach = append(v.Reach, SurfaceView{name, identity})
			}
			slices.SortFunc(v.Reach, func(a, b SurfaceView) int { return strings.Compare(a.Name, b.Name) })
			var grants struct {
				Read, Write, Mach, Connect []json.RawMessage
				Scratch                    string
			}
			_ = json.Unmarshal(h.Authority, &grants)
			v.Grants = len(grants.Read) + len(grants.Write) + len(grants.Mach) + len(grants.Connect)
			if grants.Scratch != "" {
				v.Grants++
			}
			all[id] = v
		}
		if isRoot {
			v.Root = true
			rootID = id
		}
		if actor != "" && !slices.Contains(v.Actors, actor) {
			v.Actors = append(v.Actors, actor)
			slices.Sort(v.Actors)
		}
	}
	for _, stamp := range tr.Order {
		h, ok := tr.Boundaries[stamp]
		if !ok {
			continue
		}
		add(h, h.Name, stamp == tr.Root)
		for _, g := range h.Refs {
			reach := map[string]json.RawMessage{}
			for _, name := range g.Reach {
				if s, ok := h.Reach[name]; ok {
					reach[name] = s
				}
			}
			add(Boundary{Reach: reach, Authority: g.Authority}, "", false)
		}
	}
	ids := []string{}
	for id := range all {
		if id != rootID {
			ids = append(ids, id)
		}
	}
	slices.Sort(ids)
	out := []BoundaryView{}
	if v := all[rootID]; v != nil {
		out = append(out, *v)
	}
	for _, id := range ids {
		out = append(out, *all[id])
	}
	return out
}

func AppState(incarnation string, revision int, views []View) (Snapshot, error) {
	out := Snapshot{Type: "snapshot", UIIncarnation: incarnation, Revision: revision, Conversations: []Conversation{}, Requests: []Request{}}
	if incarnation == "" || revision < 0 {
		return out, fmt.Errorf("invalid publication identity")
	}
	ids, requests := map[string]bool{}, map[string]bool{}
	for _, v := range views {
		c := v.Conversation
		if c.Conversation == "" || ids[c.Conversation] {
			return out, fmt.Errorf("invalid or duplicate conversation")
		}
		ids[c.Conversation] = true
		switch c.Phase {
		case "ready", "pending", "failed", "closed":
		default:
			return out, fmt.Errorf("invalid phase")
		}
		if c.Entries == nil || c.Boundaries == nil || v.Requests == nil {
			return out, fmt.Errorf("incomplete conversation projection")
		}
		out.Conversations = append(out.Conversations, c)
		for _, r := range v.Requests {
			if r.RequestID == "" || requests[r.RequestID] {
				return out, fmt.Errorf("duplicate request identity")
			}
			requests[r.RequestID] = true
			out.Requests = append(out.Requests, r)
		}
	}
	return out, nil
}
