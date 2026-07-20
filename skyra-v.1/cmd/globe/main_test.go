package main

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"skyra-v1/src/skyra/relation"
	dnav3 "skyra-v1/src/skyra/skyra-dna-v3"
	"skyra-v1/src/skyra/world3d"
)

// The whole loop, headless: real Being, real globe, real kernel, real walks — scripted
// model. The script behaves like a competent planner behaves: cold it tries the obvious
// thing (walk to the beacon, activate), when the world reports `missing sigil` it plans
// the full recipe, and warm it recognizes the precedent and replays it. What the test
// asserts is the SHAPE: passes-to-closure falls to 1 and stays there, every episode
// closes, and the edges exist. If the mounting is wrong anywhere — the wave, the walk,
// the triple translation, the settlement — the curve is where it shows.
func TestGlobeLearningCurve(t *testing.T) {
	kernel := relation.NewRelation()
	w := world3d.New(kernel, 12, shrineCell, beaconCell)
	proxy := newBeingProxy("skyra")
	w.Place(proxy, startCell)
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

	being := dnav3.NewBeing(realm, scriptedPlanner(t))
	being.Aff = globeAffordances()
	being.System = globeSystem

	target := []dnav3.Triple{{Subject: "beacon:/", Predicate: "status", Object: "lit"}}
	var passes []int
	for run := 1; run <= 4; run++ {
		w.ResetEpisode()
		w.Place(proxy, startCell)
		being.NewEpisode(realm)
		seed := realm.Panorama()
		if seed.Status != "succeeded" {
			t.Fatalf("run %d: panorama failed: %s", run, seed.Err)
		}
		being.ObserveImpulse(seed.Effects, target)
		c, err := being.Run(target, 6)
		if err != nil {
			t.Fatalf("run %d: %v", run, err)
		}
		if c.Unresolved {
			t.Fatalf("run %d did not close (passes=%d rejects=%v)", run, c.Passes, c.Rejects)
		}
		if !w.Lit() {
			t.Fatalf("run %d closed but the beacon is not lit", run)
		}
		passes = append(passes, c.Passes)
	}

	t.Logf("passes-to-closure: %v", passes)
	if passes[0] < 2 {
		t.Fatalf("cold run closed in %d pass(es) — the hidden prerequisite did not force discovery", passes[0])
	}
	if passes[len(passes)-1] != 1 {
		t.Fatalf("warm run took %d passes — precedent is not doing its job", passes[len(passes)-1])
	}
	if being.Metrics.Resolved == 0 {
		t.Fatal("no bets were ever settled")
	}
	if len(being.EdgeReport()) == 0 {
		t.Fatal("no edges learned")
	}
}

var coordRe = regexp.MustCompile(`(\S+)\s+location\s+(coord:\S+)`)

// scriptedPlanner is the fake model: reads the present the way a planner would, emits
// the grammar the parser accepts. Cold → naive. Hint or precedent visible → full recipe.
func scriptedPlanner(t *testing.T) func(system, present string) (string, error) {
	return func(system, present string) (string, error) {
		locs := map[string]string{}
		for _, m := range coordRe.FindAllStringSubmatch(present, -1) {
			locs[m[1]] = m[2]
		}
		shrine, beacon := locs["shrine:/"], locs["beacon:/"]
		if shrine == "" || beacon == "" {
			return "", fmt.Errorf("present has no landmark locations:\n%s", present)
		}
		knowsRecipe := strings.Contains(present, "missing sigil") ||
			(strings.Contains(present, "PRECEDENT") && strings.Contains(present, `primitive="take"`))
		if knowsRecipe {
			return fmt.Sprintf(`<framing>the beacon needs the sigil: fetch it from the shrine, then light the beacon</framing>
<act id="a1" target="%[1]s" primitive="move"></act>
<act id="a2" target="sigil:/" primitive="take" requires="agent:self at %[1]s"></act>
<act id="a3" target="%[2]s" primitive="move" requires="sigil:/ carried true"></act>
<act id="a4" target="beacon:/" primitive="activate" requires="agent:self at %[2]s"></act>
<expect of="a1">  0.95  1  agent:self  at       %[1]s  </expect>
<expect of="a2">  0.90  1  sigil:/     carried  true   </expect>
<expect of="a3">  0.95  1  agent:self  at       %[2]s  </expect>
<expect of="a4">  0.85  1  beacon:/    status   lit    </expect>`, shrine, beacon), nil
		}
		return fmt.Sprintf(`<framing>walk to the beacon and activate it</framing>
<act id="a1" target="%[1]s" primitive="move"></act>
<act id="a2" target="beacon:/" primitive="activate" requires="agent:self at %[1]s"></act>
<expect of="a1">  0.95  1  agent:self  at      %[1]s  </expect>
<expect of="a2">  0.70  1  beacon:/    status  lit    </expect>`, beacon), nil
	}
}
