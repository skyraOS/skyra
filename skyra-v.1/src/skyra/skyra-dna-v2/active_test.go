package dnav2

import (
	"testing"

	"skyra-v1/src/skyra/world"
)

// The active-episode reference (active-grouping.md): Create makes an episode active,
// newer Creates take over, and SetActive can switch back — the "current episode" the
// sensory lane will attach interactions to. The active ID lives on the slice.
func TestActiveTracksCurrentEpisode(t *testing.T) {
	episodes := NewEpisodeStore(world.NewSlice("skyra:nucleus", "skyra"))
	if episodes.Active() != nil {
		t.Fatal("no active episode before the first episode")
	}
	g1 := episodes.Create("g1", "first", 1)
	if episodes.Active() != g1 {
		t.Fatal("Create should make the new episode active")
	}
	g2 := episodes.Create("g2", "second", 2)
	if episodes.Active() != g2 {
		t.Fatal("a newer Create becomes the active episode")
	}
	episodes.SetActive(g1)
	if episodes.Active() != g1 {
		t.Fatal("SetActive switches the active episode (e.g. returning to a prior thread)")
	}
}

// The framing operator parses both forms; edit wins over a stray <new> in its frame.
func TestParseFrameOp(t *testing.T) {
	op, id, frame, ok := ParseFrameOp(`<edit id="g:2">helping with the kernel</edit>`)
	if !ok || op != "edit" || id != "g:2" || frame != "helping with the kernel" {
		t.Fatalf("edit: op=%q id=%q frame=%q ok=%v", op, id, frame, ok)
	}
	op, id, frame, ok = ParseFrameOp(`<new>a fresh topic</new>`)
	if !ok || op != "new" || id != "" || frame != "a fresh topic" {
		t.Fatalf("new: op=%q id=%q frame=%q ok=%v", op, id, frame, ok)
	}
	if _, _, _, ok := ParseFrameOp("no tags here"); ok {
		t.Fatal("garbage should not parse")
	}
}
