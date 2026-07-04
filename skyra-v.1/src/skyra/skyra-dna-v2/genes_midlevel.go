package dnav2

import (
	"fmt"
	"sort"
	"strings"

	"skyra-v1/src/skyra"
	dna "skyra-v1/src/skyra/skyra-dna"
)

// midlevel:frame — assemble the exchange from the entity index, infer the gist
// (real inference), create the grouping. Ported from v1 midlevel.observeExternal
// (cold-start create path). Sets root.Grouping so the membrane threads it to the
// fan-out's predict workers. Reports "framed" / "empty".
func midlevelFrame(root *EphemeralRoot, origin skyra.Space, impulse any) {
	present := renderExchange(root.Entities)
	if present == "" {
		root.Condition = "empty"
		return
	}
	infer := root.DNA().Infer()
	if infer == nil {
		root.Condition = "empty"
		return
	}
	resp, err := infer(coldStartSystem, present)
	if err != nil {
		root.Condition = "empty" // a failed call frames nothing this cycle
		return
	}
	label, ok := dna.ParseLabel(resp)
	if !ok {
		root.Condition = "empty"
		return
	}
	ts := root.DNA().Tick()
	groupID := fmt.Sprintf("%s:g:%d", root.DNA().BeingID(), ts)
	root.Grouping = root.DNA().Groupings().Create(groupID, label, ts)
	root.Condition = "framed"
}

// renderExchange turns the entity index's interaction records into a timestamp-
// ordered transcript (who said what) — the present the framing pass reads.
func renderExchange(ei *EntityIndex) string {
	if ei == nil {
		return ""
	}
	type line struct {
		ts int
		s  string
	}
	var lines []line
	for _, e := range ei.Snapshot() {
		for _, c := range e.Children() {
			mc, ok := c.(*dna.MemoryCell)
			if !ok {
				continue
			}
			if rec, ok := mc.Content().(dna.InteractionRecord); ok {
				lines = append(lines, line{rec.Timestamp, fmt.Sprintf("[t%d] %s: %s", rec.Timestamp, rec.From, rec.Utterance)})
			}
		}
	}
	sort.SliceStable(lines, func(i, j int) bool { return lines[i].ts < lines[j].ts })
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(l.s)
		b.WriteByte('\n')
	}
	return b.String()
}

// coldStartSystem — the framing prompt (ported verbatim from v1).
const coldStartSystem = "You are an evolving being reading a live exchange. In one short " +
	"phrase, predict what this exchange is about and what you should be trying to do. " +
	"Respond with only the tag: <label>your framing</label>"
