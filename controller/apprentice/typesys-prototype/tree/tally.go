package tree

import "strings"

// --- tally (59) ---

// A Vote is one voter's branch selecting the values: the static
// implementation (the voter and its branch) and its decision for these
// values (the canonical instruction, each $n a position).
type Vote struct {
	Voter  int    `json:"voter"`
	Branch int    `json:"branch"`
	Rank   int    `json:"rank"`
	Instr  string `json:"instr"`
}

// A Ballot is what evaluating a base's refinement on an arrival gives: the
// static interface (the base), its dynamic observation (the values, by
// hash), and the votes cast.
type Ballot struct {
	Base        string `json:"base"`
	Observation string `json:"observation"`
	Votes       []Vote `json:"votes"`
}

// ObservationID is the values' identity: equal values, equal id.
func ObservationID(vals map[string]any) string {
	return sum([]any{observationTag, vals})
}

// Cast evaluates each branch on the values: every branch that selects them
// is its voter's vote. A chain's branches have effective conditions, so a
// voter casts at most one.
func Cast(bs []Branch, vals map[string]any) ([]Vote, error) {
	sel, err := selected(bs, vals)
	if err != nil {
		return nil, err
	}
	votes := make([]Vote, len(sel))
	for i, b := range sel {
		votes[i] = Vote{Voter: b.Voter, Branch: b.ID, Rank: b.Rank, Instr: b.Instr}
	}
	return votes, nil
}

// NewBallot casts the refinement's votes on an arrival at its base.
func NewBallot(base string, bs []Branch, vals map[string]any) (Ballot, error) {
	votes, err := Cast(bs, vals)
	return Ballot{Base: base, Observation: ObservationID(vals), Votes: votes}, err
}

// A Tally is how a base's votes decide. Inhibition is ι in thousandths
// (1000 is 1): the margin a winner needs is ⌈(1 − ι) · cast⌉, never below 1,
// so 1000 is a plurality and 0 is every vote cast. Floor is the coarsest
// level counted, 2 when 0: continuing is never merged with finishing, one
// tool with another, or exit 0 with exit 1.
type Tally struct {
	Inhibition int
	Floor      int
}

// Exact is the finest level: the whole instruction.
const Exact = 4

// Required is the margin a winner needs among cast votes.
func Required(inhibition, cast int) int {
	return max(1, ((1000-min(max(inhibition, 0), 1000))*cast+999)/1000)
}

// Levels are the levels a tally tries, the exact one first, down to the
// floor.
func (t Tally) Levels() []int {
	floor := t.Floor
	if floor == 0 {
		floor = 2
	}
	var ls []int
	for k := Exact; k >= floor; k-- {
		ls = append(ls, k)
	}
	return ls
}

// InstrParts are a stored instruction's parts: its form, then its head (the
// tool run, a return's exit status, the delegated goal), then the rest in
// order.
func InstrParts(instr string) []string {
	if inner, ok := strings.CutPrefix(instr, "delegate("); ok {
		args := strings.Split(strings.TrimSuffix(inner, ")"), ", ")
		goal := strings.TrimSuffix(strings.TrimPrefix(args[0], "goal("), ")")
		return append([]string{"delegate", goal}, args[1:]...)
	}
	fields := strings.Split(instr, " ")
	if fields[0] == "return" {
		return fields
	}
	return append([]string{"run"}, fields...)
}

// GroupKey is an instruction's group at level k: its first k parts, or the
// whole instruction at the exact level.
func GroupKey(k int, instr string) string {
	if k >= Exact {
		return instr
	}
	ps := InstrParts(instr)
	return strings.Join(ps[:min(k, len(ps))], "\x00")
}

// Decide tallies the votes, finest level first. A group wins a level when
// it has the margin and leads every other group by it; at most one can.
// The tally runs the exact instruction in the winning group with the most
// votes, ties to the oldest. With no votes it is null; with no winner at any
// level, a conflict. The outcome does not depend on the votes' order.
func (t Tally) Decide(votes []Vote) Selection {
	if len(votes) == 0 {
		return Selection{Kind: SelectNull}
	}
	req := Required(t.Inhibition, len(votes))
	exact := map[string]int{}
	for _, v := range votes {
		exact[v.Instr]++
	}
	for _, k := range t.Levels() {
		count := map[string]int{}
		for _, v := range votes {
			count[GroupKey(k, v.Instr)]++
		}
		var best *Vote
		for i, x := range votes {
			g := GroupKey(k, x.Instr)
			n, wins := count[g], req <= count[g]
			for h, c := range count {
				if h != g && c+req > n {
					wins = false
				}
			}
			if wins && (best == nil || exact[x.Instr] > exact[best.Instr] ||
				exact[x.Instr] == exact[best.Instr] && x.Rank < best.Rank) {
				best = &votes[i]
			}
		}
		if best != nil {
			return Selection{Kind: SelectOne, Instr: best.Instr, Level: k}
		}
	}
	return Selection{Kind: SelectConflict}
}
