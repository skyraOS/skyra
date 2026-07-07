package dnav2

import (
	"testing"

	"skyra-v1/src/skyra/world"
)

// The active-grouping reference (active-grouping.md): Create makes a grouping active,
// newer Creates take over, and SetActive can switch back — the "current episode" the
// sensory lane will attach interactions to. The active ID lives on the slice.
func TestActiveTracksCurrentEpisode(t *testing.T) {
	gi := NewGroupingIndex(world.NewSlice("skyra:nucleus", "skyra"))
	if gi.Active() != nil {
		t.Fatal("no active episode before the first grouping")
	}
	g1 := gi.Create("g1", "first", 1)
	if gi.Active() != g1 {
		t.Fatal("Create should make the new grouping active")
	}
	g2 := gi.Create("g2", "second", 2)
	if gi.Active() != g2 {
		t.Fatal("a newer Create becomes the active episode")
	}
	gi.SetActive(g1)
	if gi.Active() != g1 {
		t.Fatal("SetActive switches the active episode (e.g. returning to a prior thread)")
	}
}

// The framing operator parses both forms; edit wins over a stray <new> in its gist.
func TestParseFrameOp(t *testing.T) {
	op, id, gist, ok := ParseFrameOp(`<edit id="g:2">helping with the kernel</edit>`)
	if !ok || op != "edit" || id != "g:2" || gist != "helping with the kernel" {
		t.Fatalf("edit: op=%q id=%q gist=%q ok=%v", op, id, gist, ok)
	}
	op, id, gist, ok = ParseFrameOp(`<new>a fresh topic</new>`)
	if !ok || op != "new" || id != "" || gist != "a fresh topic" {
		t.Fatalf("new: op=%q id=%q gist=%q ok=%v", op, id, gist, ok)
	}
	if _, _, _, ok := ParseFrameOp("no tags here"); ok {
		t.Fatal("garbage should not parse")
	}
}
