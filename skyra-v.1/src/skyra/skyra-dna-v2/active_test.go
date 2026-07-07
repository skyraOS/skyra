package dnav2

import "testing"

// The active-grouping reference (active-grouping.md): Create makes a grouping active,
// newer Creates take over, and SetActive can switch back — the "current episode" the
// sensory lane will attach interactions to.
func TestActiveTracksCurrentEpisode(t *testing.T) {
	gi := NewGroupingIndex(nil)
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
