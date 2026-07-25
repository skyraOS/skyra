package dnav2

import (
	"math"
	"testing"
)

// Cross lifts F from 0 and repeated crossings drive it toward 1; ticks without a
// crossing decay it. This is frequency + recency in one field.
func TestCrossRaisesAndDecaysF(t *testing.T) {
	a := newActivation()
	if a.readF(0) != 0 {
		t.Fatal("fresh F should be 0")
	}
	a.Cross(0)
	f1 := a.readF(0)
	if f1 <= 0 || f1 >= 1 {
		t.Fatalf("after one cross F=%.3f, want 0<F<1", f1)
	}
	a.Cross(0)
	if a.readF(0) <= f1 {
		t.Fatal("a second crossing should raise F")
	}
	// no crossing for 10 ticks → decayed read is lower than the stored value
	high := a.readF(0)
	if a.readF(10) >= high {
		t.Fatal("F should decay over ticks without a crossing")
	}
}

// Grade moves A toward the outcome; N drives the learning rate (1/N early). One grade of
// g=1 from A=0 with N=1 lands A exactly at 1.
func TestGradeMovesAlignmentByCount(t *testing.T) {
	a := newActivation()
	a.Grade(1, 0.5)
	if a.N != 1 {
		t.Fatalf("N=%d, want 1", a.N)
	}
	if math.Abs(a.A-1) > 1e-9 { // α = 1/1 = 1 → A = 0 + 1*(1-0) = 1
		t.Fatalf("A=%.4f after one g=1, want 1", a.A)
	}
	// a following g=0 with N=2 → α=0.5 → A = 1 + .5*(0-1) = .5
	a.Grade(0, 0.5)
	if math.Abs(a.A-0.5) > 1e-9 {
		t.Fatalf("A=%.4f after g=0, want 0.5", a.A)
	}
}

// Salience is immutable at birth: the first Grade stamps S, later grades don't move it.
func TestSalienceImmutable(t *testing.T) {
	a := newActivation()
	a.Grade(1, 0.8)
	if a.S != 0.8 {
		t.Fatalf("S=%.2f, want 0.8 (stamped on first grade)", a.S)
	}
	a.Grade(1, 0.2)
	if a.S != 0.8 {
		t.Fatalf("S=%.2f, want 0.8 (immutable)", a.S)
	}
}

// conf is a decreasing exploration bonus: 2 at N=0, →1 as N grows.
func TestConfidenceExploreBonus(t *testing.T) {
	a := newActivation()
	if a.conf() != 2 {
		t.Fatalf("conf at N=0 = %.2f, want 2", a.conf())
	}
	a.N = 100
	if a.conf() >= 1.1 || a.conf() <= 1 {
		t.Fatalf("conf at N=100 = %.3f, want just above 1", a.conf())
	}
}

// Strength composes the factors: with equal frequency, an edge graded UP outweighs one
// graded DOWN — the learned signal the frame will read. (Both crossed first so F>0.)
func TestStrengthReflectsAlignment(t *testing.T) {
	up, down := newActivation(), newActivation()
	up.Cross(0)
	down.Cross(0)
	up.Grade(1, 1)    // aligned outcome
	down.Grade(-1, 1) // failed outcome
	if up.Strength(0) <= down.Strength(0) {
		t.Fatalf("up=%.3f down=%.3f — aligned edge should outweigh failed edge",
			up.Strength(0), down.Strength(0))
	}
	// a failed edge stays barely alive (soft rect), not zero, so it can recover.
	if down.Strength(0) <= 0 {
		t.Fatalf("failed-edge strength=%.3f, want >0 (soft rect, recoverable)", down.Strength(0))
	}
}

// Bootstrap: a fresh edge has strength 0 (F=0) until it is crossed — and a born edge IS
// crossed at birth, which lifts it into contention. No deadlock.
func TestBootstrapNeedsCross(t *testing.T) {
	a := newActivation()
	if a.Strength(0) != 0 {
		t.Fatal("un-crossed fresh edge should have 0 strength")
	}
	a.Cross(0) // node birth crosses the new edge
	if a.Strength(0) <= 0 {
		t.Fatal("a crossed edge should be selectable (strength > 0)")
	}
}

// The store keys records by edge and get-or-creates on write.
func TestStoreGradeAndRead(t *testing.T) {
	s := NewActivationStore()
	if _, ok := s.Strength("a→b", 0); ok {
		t.Fatal("unknown edge should report no record")
	}
	s.Cross("a→b", 0)
	s.Grade("a→b", 1, 1)
	str, ok := s.Strength("a→b", 0)
	if !ok || str <= 0 {
		t.Fatalf("graded edge strength=%.3f ok=%v, want >0", str, ok)
	}
}
