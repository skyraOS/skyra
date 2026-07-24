package main

import (
	"fmt"
	"testing"
	"time"

	dnav3 "skyra-v1/src/skyra/cognitive-agent/skyra-dna-v3"
	"skyra-v1/src/skyra/spacetime/relation"
	"skyra-v1/src/skyra/distributed-agent/world3d"
)

// The duel, headless, with the snitch held still: the being perceives it, walks to it,
// swipes, and the catch lands — target observed as `snitch:/ caught-by skyra`. This
// verifies the whole game mounting (catch primitive, snitch triples, seeker vocabulary,
// walk + settle) without a flying snitch, whose evasion is deliberately untestable-fast.
func TestSeekerCatchesStaticSnitch(t *testing.T) {
	kernel := relation.NewRelation()
	w := world3d.New(kernel, 12, shrineCell, beaconCell)
	proxy := newBeingProxy("skyra")
	snitch := newBeingProxy("snitch")
	w.Place(proxy, startCell)
	w.Place(snitch, snitchStart)
	w.SetSnitch(snitch.ID())
	realm := newGlobeRealm(w, kernel, proxy)

	stop := make(chan struct{})
	defer close(stop)
	go func() {
		for {
			select {
			case <-stop:
				return
			default:
				w.Tick()
				time.Sleep(time.Millisecond)
			}
		}
	}()

	being := dnav3.NewBeing(realm, scriptedSeeker(t))
	being.Aff = globeAffordances()
	being.System = seekerSystem

	seed := realm.Panorama()
	target := []dnav3.Triple{{Subject: "snitch:/", Predicate: "caught-by", Object: "skyra"}}
	being.ObserveImpulse(seed.Effects, target)
	c, err := being.Run(target, 4)
	if err != nil {
		t.Fatal(err)
	}
	if c.Unresolved {
		t.Fatalf("hunt did not close (passes=%d rejects=%v)", c.Passes, c.Rejects)
	}
	if got := w.CaughtBy(); got != "skyra" {
		t.Fatalf("caughtBy=%q, want skyra", got)
	}
}

func TestOpponentCatchSupersedesQueuedExpression(t *testing.T) {
	kernel := relation.NewRelation()
	w := world3d.New(kernel, 12, shrineCell, beaconCell)
	skyra := newBeingProxy("skyra")
	player := newBeingProxy("you")
	snitch := newBeingProxy("snitch")
	cell := world3d.Cell{F: 0, X: 4, Y: 4}
	w.Place(skyra, cell)
	w.Place(player, cell)
	w.Place(snitch, cell)
	w.SetSnitch(snitch.ID())

	if ok, missing := w.Catch(player.ID()); !ok {
		t.Fatalf("opponent catch failed: missing %q", missing)
	}

	realm := newGlobeRealm(w, kernel, skyra)
	out := realm.Do(dnav3.Act{Primitive: "catch", Target: "snitch:/"})
	if out.Status != "superseded" {
		t.Fatalf("status=%q, want superseded", out.Status)
	}
	if len(out.Effects) != 2 ||
		out.Effects[0] != (dnav3.Triple{Subject: "snitch:/", Predicate: "status", Object: "caught"}) ||
		out.Effects[1] != (dnav3.Triple{Subject: "snitch:/", Predicate: "caught-by", Object: "you"}) {
		t.Fatalf("supersession effects = %#v", out.Effects)
	}
}

// scriptedSeeker chases the last observed snitch position: move there, then swipe.
func scriptedSeeker(t *testing.T) func(system, present string) (string, error) {
	return func(system, present string) (string, error) {
		// snitch position line: `snitch:/ at coord:f,x,y`
		var f, x, y int
		if _, err := fmt.Sscanf(afterMarker(present, "snitch:/ at coord:"), "%d,%d,%d", &f, &x, &y); err != nil {
			return "", fmt.Errorf("present has no snitch position:\n%s", present)
		}
		snitchAt := fmt.Sprintf("coord:%d,%d,%d", f, x, y)
		return fmt.Sprintf(`<framing>fly to the snitch's last known cell and swipe</framing>
<act id="a1" target="%[1]s" primitive="move"></act>
<act id="a2" target="snitch:/" primitive="catch" requires="agent:self at %[1]s"></act>
<expect of="a1">  0.95  1  agent:self  at      %[1]s   </expect>
<expect of="a2">  0.60  1  snitch:/    status  caught  </expect>`, snitchAt), nil
	}
}

func afterMarker(s, marker string) string {
	for i := 0; i+len(marker) <= len(s); i++ {
		if s[i:i+len(marker)] == marker {
			end := i + len(marker)
			for end < len(s) && s[end] != '\n' && s[end] != ' ' {
				end++
			}
			return s[i+len(marker) : end]
		}
	}
	return ""
}
