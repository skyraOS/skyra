package dnav2

import (
	"fmt"
	"sort"
	"strings"
)

// PraxisSnapshot is a read-only projection of the causal substrate. Inspection
// deliberately ignores edge strength: a zero-strength edge remains historical
// fact even though ordinary cognitive traversal gives it no excitation.
type PraxisSnapshot struct {
	ActiveSetID string              `json:"activeSetId,omitempty"`
	ContextID   string              `json:"contextId,omitempty"`
	Sets        []PraxisSetView     `json:"sets"`
	Nodes       []PraxisNodeView    `json:"nodes"`
	Edges       []PraxisEdgeView    `json:"edges"`
	Budgets     []PraxisBudgetView  `json:"budgets"`
	Closures    []FinalClosureMatch `json:"finalClosureCandidates"`
}

// PraxisBudgetView is one lineage's AIMD controller state. Budgets belong to
// the emitting voices (Efficient/Final), not to contextual edges.
type PraxisBudgetView struct {
	Lineage string             `json:"lineage"`
	Current float64            `json:"current"`
	History []QueryBudgetEvent `json:"history,omitempty"`
}

type PraxisSetView struct {
	AitiaSet
	Amplitude map[string]float64 `json:"amplitude"`
}

type PraxisNodeView struct {
	ID          string       `json:"id"`
	Sentence    string       `json:"sentence"`
	Timestamp   int          `json:"timestamp"`
	Provenance  Provenance   `json:"provenance"`
	MaterialIDs []string     `json:"materialIds"`
	FormalEdges []EdgeRecord `json:"formalEdges"`
}

type PraxisEdgeView struct {
	EdgeRecord
	HasGrade    bool     `json:"hasGrade"`
	Identity    float64  `json:"identity"`
	Familiarity float64  `json:"familiarity"`
	Strength    float64  `json:"strength"`
	GradeCount  int      `json:"gradeCount"`
	EvidenceIDs []string `json:"evidenceIds"`
}

// HistoryHop is one deterministic typed-port adjacency. It is an inspection
// walk, not a weighted recall traversal.
type HistoryHop struct {
	Depth  int    `json:"depth"`
	FromID string `json:"fromId"`
	Port   string `json:"port"`
	Key    string `json:"key"`
	ToID   string `json:"toId"`
}

// FinalClosureMatch reports the candidate rule without applying a grade: a
// later Efficient source cites an earlier Final target through its immutable
// material/inherited lineage.
type FinalClosureMatch struct {
	FinalSetID        string `json:"finalSetId"`
	FinalEdgeID       string `json:"finalEdgeId"`
	TargetID          string `json:"targetId"`
	LaterSetID        string `json:"laterSetId"`
	EfficientEdgeID   string `json:"efficientEdgeId"`
	EfficientSourceID string `json:"efficientSourceId"`
}

// InspectPraxis returns all contexts when contextID is empty, or one Aitia
// context otherwise. The active pointer is reported independently.
func (d *DNA) InspectPraxis(contextID string) PraxisSnapshot {
	d.causal.mu.Lock()
	defer d.causal.mu.Unlock()

	snapshot := PraxisSnapshot{ContextID: contextID}
	snapshot.ActiveSetID = activeSetGet(d.nucleus, d.beingID)
	signed := d.signed.All()
	budgets := d.budgets.All()
	tick := d.Clock()

	setIDs := sortedSetIDs(d.causal.sets)
	nodeIDs := map[string]bool{}
	for _, id := range setIDs {
		set := d.causal.sets[id]
		if contextID != "" && set.ID != contextID {
			continue
		}
		strengths := map[string]float64{}
		for _, edgeID := range []string{set.EfficientEdgeID, set.FinalEdgeID} {
			if state, ok := signed[edgeID]; ok {
				strengths[edgeID] = state.strength(tick)
			}
		}
		snapshot.Sets = append(snapshot.Sets, PraxisSetView{AitiaSet: set, Amplitude: AllocateAmplitude(1, strengths)})
		nodeIDs[set.PresentID] = true
		for _, edgeID := range []string{set.EfficientEdgeID, set.FinalEdgeID} {
			if edge := d.causal.edgeByID[edgeID]; edge != nil {
				nodeIDs[edge.Record.FromID], nodeIDs[edge.Record.ToID] = true, true
			}
		}
	}
	if contextID == "" {
		for id := range d.causal.byID {
			nodeIDs[id] = true
		}
	}

	for _, id := range sortedBoolKeys(nodeIDs) {
		cell := d.causal.byID[id]
		if cell == nil {
			continue
		}
		content, _ := cell.Content().(CausalNodeContent)
		view := PraxisNodeView{ID: id, Sentence: content.Sentence, Timestamp: content.Timestamp,
			Provenance: content.Provenance, MaterialIDs: append([]string(nil), d.causal.materials[id]...)}
		for _, record := range d.causal.edges[id] {
			if record.Kind == EdgeStructural {
				view.FormalEdges = append(view.FormalEdges, record)
			}
		}
		snapshot.Nodes = append(snapshot.Nodes, view)
	}

	edgeIDs := make([]string, 0, len(d.causal.edgeByID))
	for id, edge := range d.causal.edgeByID {
		if contextID == "" || edge.Record.ContextID == contextID || nodeIDs[edge.Record.ContextID] {
			edgeIDs = append(edgeIDs, id)
		}
	}
	sort.Strings(edgeIDs)
	for _, id := range edgeIDs {
		edge := d.causal.edgeByID[id]
		view := PraxisEdgeView{EdgeRecord: edge.Record}
		if state, ok := signed[id]; ok {
			view.HasGrade, view.Identity, view.Familiarity = true, state.Identity, state.readF(tick)
			view.Strength, view.GradeCount = state.strength(tick), state.GradeCount
		}
		for evidenceID := range edge.ChildrenIn(PortEvidence) {
			view.EvidenceIDs = append(view.EvidenceIDs, evidenceID)
		}
		sort.Strings(view.EvidenceIDs)
		snapshot.Edges = append(snapshot.Edges, view)
	}
	for _, lineage := range sortedBoolKeys(boolKeys(budgets)) {
		snapshot.Budgets = append(snapshot.Budgets, PraxisBudgetView{
			Lineage: lineage, Current: budgets[lineage], History: d.budgets.History(lineage),
		})
	}
	snapshot.Closures = d.finalClosureCandidatesLocked()
	return snapshot
}

func boolKeys(values map[string]float64) map[string]bool {
	out := make(map[string]bool, len(values))
	for key := range values {
		out[key] = true
	}
	return out
}

// WalkHistory follows explicit typed ports regardless of signed magnitude.
func (d *DNA) WalkHistory(rootID string, maxDepth int, ports ...string) []HistoryHop {
	if maxDepth <= 0 {
		return nil
	}
	allowed := map[string]bool{}
	for _, port := range ports {
		allowed[port] = true
	}
	d.causal.mu.Lock()
	defer d.causal.mu.Unlock()
	root := inspectionCellLocked(d.causal, rootID)
	if root == nil {
		return nil
	}
	type item struct {
		cell  *MemoryCell
		depth int
	}
	queue := []item{{root, 0}}
	seen := map[string]bool{rootID: true}
	var out []HistoryHop
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current.depth >= maxDepth {
			continue
		}
		for _, port := range current.cell.ChildPorts() {
			if len(allowed) > 0 && !allowed[port] {
				continue
			}
			children := current.cell.ChildrenIn(port)
			keys := make([]string, 0, len(children))
			for key := range children {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				child, ok := children[key].(interface{ ID() string })
				if !ok {
					continue
				}
				out = append(out, HistoryHop{Depth: current.depth + 1, FromID: current.cell.ID(), Port: port, Key: key, ToID: child.ID()})
				if next := inspectionCellLocked(d.causal, child.ID()); next != nil && !seen[next.ID()] {
					seen[next.ID()] = true
					queue = append(queue, item{next, current.depth + 1})
				}
			}
		}
	}
	return out
}

// FinalClosureCandidates evaluates the candidate rule read-only.
func (d *DNA) FinalClosureCandidates() []FinalClosureMatch {
	d.causal.mu.Lock()
	defer d.causal.mu.Unlock()
	return d.finalClosureCandidatesLocked()
}

func (d *DNA) finalClosureCandidatesLocked() []FinalClosureMatch {
	sets := d.causal.sets
	ids := sortedSetIDs(sets)
	var out []FinalClosureMatch
	for _, finalSetID := range ids {
		prior := sets[finalSetID]
		finalEdge := d.causal.edgeByID[prior.FinalEdgeID]
		if finalEdge == nil {
			continue
		}
		target := finalEdge.Record.ToID
		for _, laterSetID := range ids {
			later := sets[laterSetID]
			if later.Timestamp <= prior.Timestamp {
				continue
			}
			efficientEdge := d.causal.edgeByID[later.EfficientEdgeID]
			if efficientEdge == nil {
				continue
			}
			source := efficientEdge.Record.FromID
			if d.causal.lineageContainsLocked(source, target, map[string]bool{}) {
				out = append(out, FinalClosureMatch{FinalSetID: prior.ID, FinalEdgeID: prior.FinalEdgeID,
					TargetID: target, LaterSetID: later.ID, EfficientEdgeID: later.EfficientEdgeID, EfficientSourceID: source})
			}
		}
	}
	return out
}

func (ci *CausalNodes) lineageContainsLocked(nodeID, targetID string, seen map[string]bool) bool {
	if nodeID == targetID {
		return true
	}
	if seen[nodeID] {
		return false
	}
	seen[nodeID] = true
	for _, materialID := range ci.materials[nodeID] {
		if materialID == targetID || ci.lineageContainsLocked(materialID, targetID, seen) {
			return true
		}
	}
	for _, record := range ci.edges[nodeID] {
		if record.Kind == EdgeInherited && (record.FromID == targetID || ci.lineageContainsLocked(record.FromID, targetID, seen)) {
			return true
		}
	}
	return false
}

func inspectionCellLocked(ci *CausalNodes, id string) *MemoryCell {
	if cell := ci.byID[id]; cell != nil {
		return cell
	}
	if edge := ci.edgeByID[id]; edge != nil {
		return edge.MemoryCell
	}
	return nil
}

func sortedSetIDs(sets map[string]AitiaSet) []string {
	ids := make([]string, 0, len(sets))
	for id := range sets {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		a, b := sets[ids[i]], sets[ids[j]]
		if a.Timestamp == b.Timestamp {
			return ids[i] < ids[j]
		}
		return a.Timestamp < b.Timestamp
	})
	return ids
}

func sortedBoolKeys(values map[string]bool) []string {
	ids := make([]string, 0, len(values))
	for id := range values {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// RenderPraxis renders a compact read-only screen suitable for the TUI and
// command-line validators.
func RenderPraxis(snapshot PraxisSnapshot) string {
	var b strings.Builder
	fmt.Fprintf(&b, "PRAXIS  active=%s  context=%s\n", valueOr(snapshot.ActiveSetID, "none"), valueOr(snapshot.ContextID, "all"))
	for _, set := range snapshot.Sets {
		fmt.Fprintf(&b, "\nSET %s  present=%s\n", set.ID, set.PresentID)
		fmt.Fprintf(&b, "  efficient=%s  amplitude=%.3f\n", set.EfficientEdgeID, set.Amplitude[set.EfficientEdgeID])
		fmt.Fprintf(&b, "  final=%s      amplitude=%.3f\n", set.FinalEdgeID, set.Amplitude[set.FinalEdgeID])
	}
	for _, budget := range snapshot.Budgets {
		fmt.Fprintf(&b, "\nBUDGET %s  current=%.1f  events=%d\n", budget.Lineage, budget.Current, len(budget.History))
	}
	for _, node := range snapshot.Nodes {
		fmt.Fprintf(&b, "\nNODE %s [%s] %s\n", node.ID, node.Provenance, node.Sentence)
		for _, edge := range node.FormalEdges {
			fmt.Fprintf(&b, "  %s --%s--> %s\n", edge.FromID, edge.Label, edge.ToID)
		}
	}
	for _, edge := range snapshot.Edges {
		if edge.Kind == EdgeStructural {
			continue
		}
		grade := "ungraded"
		if edge.HasGrade {
			grade = fmt.Sprintf("strength=%+.3f I=%+.3f F=%+.3f n=%d", edge.Strength, edge.Identity, edge.Familiarity, edge.GradeCount)
		}
		fmt.Fprintf(&b, "\nEDGE %s [%s] %s -> %s  %s  evidence=%v\n", edge.ID, edge.Kind, edge.FromID, edge.ToID, grade, edge.EvidenceIDs)
	}
	if len(snapshot.Closures) > 0 {
		fmt.Fprintf(&b, "\nFINAL CLOSURE CANDIDATES: %d\n", len(snapshot.Closures))
	}
	return strings.TrimRight(b.String(), "\n")
}

func valueOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
