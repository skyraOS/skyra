package dnav3

import (
	"testing"
	"time"
)

func obsAt(wall int64, s, p, o string) Observation {
	return Observation{
		Triple:     Triple{Mode: ModeExact, Subject: s, Predicate: p, Object: o},
		Provenance: ProvenanceRaw,
		Wall:       wall,
	}
}

// paperToy is paper test #4's fixture (specs/2026-07-20/observe-paper-test-4.md): the
// five-occurrence household history, unsegmented. The rules must recover what was
// hand-assigned in tests #1–#3: five spans, three classes, S5 provisional.
func paperToy() []Observation {
	day1 := time.Date(2026, 7, 18, 8, 0, 0, 0, time.UTC).UnixMilli()
	min := int64(60_000)
	at := func(base int64, m int64) int64 { return base + m*min }
	evening := at(day1, 11*60+30)  // 19:30
	day2 := at(day1, 25*60)        // next day 09:00
	afternoon := at(day2, 6*60+20) // 15:20
	garage := at(afternoon, 20)    // 15:40

	return []Observation{
		// S1 — the water routine (kitchen fridge).
		obsAt(at(day1, 0), "being:michael", "state", "thirsty"),
		obsAt(at(day1, 2), "thing:bottle", "taken-from", "fridge:kitchen"),
		obsAt(at(day1, 4), "being:michael", "drank", "thing:bottle"),
		obsAt(at(day1, 6), "thing:bottle", "put-back", "fridge:kitchen"),
		// S2 — the fridge confusion (kitchen corrected to garage).
		obsAt(at(evening, 0), "being:michael", "want", "thing:drink"),
		obsAt(at(evening, 2), "being:michael", "rejected", "fridge:kitchen"),
		obsAt(at(evening, 4), "being:michael", "corrected-to", "fridge:garage"),
		obsAt(at(evening, 6), "thing:soda", "fetched-from", "fridge:garage"),
		// S3 — plant care (the "water" homonym lives here as its own identity).
		obsAt(at(day2, 0), "being:michael", "noticed", "thing:plants"),
		obsAt(at(day2, 2), "being:skyra", "watered", "thing:plants"),
		obsAt(at(day2, 4), "thing:plants", "state", "perked"),
		// S4 — the second garage fetch (gives the class its statistics).
		obsAt(at(afternoon, 0), "being:michael", "want", "thing:soda"),
		obsAt(at(afternoon, 2), "thing:soda", "fetched-from", "fridge:garage"),
		obsAt(at(afternoon, 4), "being:michael", "has", "thing:soda"),
		// S5 — the garage moment (open: paper test #3's uncommitted impulse).
		obsAt(at(garage, 0), "being:michael", "holds", "thing:bottle"),
		obsAt(at(garage, 1), "being:michael", "put-back", "thing:bottle"),
	}
}

func TestSpansRecoverPaperToy(t *testing.T) {
	spans := Spans(paperToy(), 0)
	if len(spans) != 5 {
		t.Fatalf("want 5 spans (4/4 boundaries), got %d: %+v", len(spans), spans)
	}
	wantFirst := []int{0, 4, 8, 11, 14}
	for i, s := range spans {
		if s.First != wantFirst[i] {
			t.Fatalf("span %d starts at log[%d], want log[%d]", i, s.First, wantFirst[i])
		}
		if open := i == len(spans)-1; s.Open != open {
			t.Fatalf("span %d Open=%v, want %v — only the tail is open", i, s.Open, open)
		}
	}
}

func TestClassesRecoverPaperToy(t *testing.T) {
	log := paperToy()
	spans := Spans(log, 0)
	classes := Classes(log, spans, 0)
	if len(classes) != 3 {
		t.Fatalf("want 3 classes, got %d: %+v", len(classes), classes)
	}
	members := func(c Class) []int { return c.Members }
	// {S1,S5} water-routine (provisional), {S2,S4} fridge-fetch, {S3} plant-care.
	if got := members(classes[0]); len(got) != 2 || got[0] != 0 || got[1] != 4 {
		t.Fatalf("class 0 members = %v, want [0 4] (bottle + put-back join the garage moment)", got)
	}
	if !classes[0].Open {
		t.Fatal("class 0 holds the open tail span — membership must be provisional (Open)")
	}
	if got := members(classes[1]); len(got) != 2 || got[0] != 1 || got[1] != 3 {
		t.Fatalf("class 1 members = %v, want [1 3] (soda + garage + fetched-from)", got)
	}
	if got := members(classes[2]); len(got) != 1 || got[0] != 2 {
		t.Fatalf("class 2 members = %v, want [2] (plant care stands alone)", got)
	}
	if classes[1].Open || classes[2].Open {
		t.Fatal("closed classes must not be marked Open")
	}
	// One shared entity must never found a class: S1·S2 share only fridge:kitchen.
	if w := classes[0].Signature["fridge:kitchen"]; w <= 0 {
		t.Fatal("fridge:kitchen should be informative (df 2/5), just insufficient alone")
	}
	// Ubiquity carries no information: michael is in every span.
	if w := classes[0].Signature["being:michael"]; w != 0 {
		t.Fatalf("being:michael weight = %v, want 0 (present in all spans)", w)
	}
}

func TestLensOmitsUbiquitousTerms(t *testing.T) {
	log := paperToy()
	classes := Classes(log, Spans(log, 0), 0)
	lens := Lens(classes)
	if _, ok := lens["being:michael"]; ok {
		t.Fatal("lens must omit weight-0 terms — michael would hand every impulse every class")
	}
	if got := lens["thing:bottle"]; len(got) != 1 || got[0] != 0 {
		t.Fatalf("lens[thing:bottle] = %v, want [0]", got)
	}
	if got := lens["fridge:garage"]; len(got) != 1 || got[0] != 1 {
		t.Fatalf("lens[fridge:garage] = %v, want [1]", got)
	}
}

// TestOverSegmentThenMerge encodes the live run's composition finding
// (specs/2026-07-20/live-segmentation-run.md): inside one sitting the gap spectrum has no
// hours-band, so Spans cuts eagerly at exchange scale and Classes remends same-class
// spans. Threshold passed explicitly: log-weights compress at small span counts, and the
// default is calibrated for repertoires of ~5+ spans.
func TestOverSegmentThenMerge(t *testing.T) {
	base := time.Date(2026, 7, 13, 11, 0, 0, 0, time.UTC).UnixMilli()
	sec := int64(1000)
	log := []Observation{
		// burst 1 — maze work
		obsAt(base, "being:michael", "asked", "thing:maze"),
		obsAt(base+5*sec, "being:skyra", "solved", "thing:maze"),
		// 120s exchange gap, then burst 2 — more maze work (same class)
		obsAt(base+125*sec, "being:michael", "asked", "thing:maze"),
		obsAt(base+130*sec, "being:skyra", "solved", "thing:maze"),
		// 300s gap, then burst 3 — unrelated greeting
		obsAt(base+430*sec, "being:michael", "greeted", "being:skyra"),
	}
	spans := Spans(log, 0) // rhythm 5s → cut 25s: over-segments into 3 spans
	if len(spans) != 3 {
		t.Fatalf("want 3 over-segmented spans, got %d", len(spans))
	}
	classes := Classes(log, spans, 0.8)
	if len(classes) != 2 {
		t.Fatalf("want 2 classes after merge, got %d: %+v", len(classes), classes)
	}
	if got := classes[0].Members; len(got) != 2 || got[0] != 0 || got[1] != 1 {
		t.Fatalf("maze spans %v should merge into one class", got)
	}
}

func TestWriteStampsWall(t *testing.T) {
	fixed := int64(1_700_000_000_000)
	old := wallNow
	wallNow = func() int64 { return fixed }
	defer func() { wallNow = old }()

	g := NewObservationGraph()
	g.Write(Observation{Triple: Triple{Subject: "a", Predicate: "b", Object: "c"}})
	g.Write(Observation{Triple: Triple{Subject: "d", Predicate: "e", Object: "f"}, Wall: 42})
	log := g.Log()
	if log[0].Wall != fixed {
		t.Fatalf("unstamped observation should get wallNow, got %d", log[0].Wall)
	}
	if log[1].Wall != 42 {
		t.Fatalf("replayed observation must keep its historical wall, got %d", log[1].Wall)
	}
}
