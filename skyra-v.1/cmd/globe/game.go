package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"skyra-v1/src/debug"
	dnav3 "skyra-v1/src/skyra/cognitive-agent/skyra-dna-v3"
	"skyra-v1/src/skyra/spacetime"
	"skyra-v1/src/skyra/distributed-agent/world3d"
)

// The seeker duel: you and the being hunt the same snitch, first catch wins.
//
// This is deliberately a harder world than the beacon. The beacon experiment is
// deterministic — the recipe sits still while the being discovers it. The snitch MOVES,
// and there is an adversary. So the duel is a live stress test of exactly the scope
// .2001 did not cover: plans go stale between planning and acting, precedent describes a
// world that no longer holds, and the honest response is lower confidence on the same
// claims — which the expects make visible in real time. Nobody promises the curve falls
// here. That is the point of playing it.

var (
	playerStart = world3d.Cell{F: 5, X: 6, Y: 6}
	snitchStart = world3d.Cell{F: 1, X: 6, Y: 6}
)

// snitchBeat paces the snitch's flight relative to grid resolution, so its speed as a
// fraction of the world stays put when -face changes.
func (s *server) snitchBeat() time.Duration {
	n := s.world.Globe().N
	ms := 420 * 12 / n
	if ms < 160 {
		ms = 160
	}
	return time.Duration(ms) * time.Millisecond
}

type game struct {
	mu     sync.Mutex
	active bool
	stop   chan struct{}

	player *beingProxy
	snitch *beingProxy
}

func newGame() *game {
	return &game{
		player: newBeingProxy("you"),
		snitch: newBeingProxy("snitch"),
	}
}

// handleGameStart resets the stage and starts the three actors: the snitch's flight,
// the being's hunt, and the referee watching for a catch.
func (s *server) handleGameStart(rw http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		http.Error(rw, "already running", http.StatusConflict)
		return
	}
	s.running = true
	s.mu.Unlock()

	g := s.game
	g.mu.Lock()
	g.active = true
	g.stop = make(chan struct{})
	g.mu.Unlock()

	s.world.ResetEpisode()
	s.world.Place(s.proxy, startCell)
	s.world.Place(g.player, playerStart)
	s.world.Place(g.snitch, snitchStart)
	s.world.SetSnitch(g.snitch.ID())

	s.hub.publish(map[string]any{"type": "game", "phase": "start",
		"you": playerStart.ID(), "skyra": startCell.ID(), "snitch": snitchStart.ID()})
	debug.Log("─── seeker duel ─── snitch loose at", snitchStart.ID())

	go s.flySnitch(g)
	go s.hunt(g)
	go s.referee(g)
	rw.WriteHeader(http.StatusAccepted)
}

// flySnitch is the snitch's own little mind: every beat, drift one cell — biased away
// from whichever seeker is closest, with enough randomness to be catchable.
func (s *server) flySnitch(g *game) {
	globe := s.world.Globe()
	for {
		select {
		case <-g.stop:
			return
		case <-time.After(s.snitchBeat()):
		}
		if s.world.CaughtBy() != "" {
			return
		}
		pos, ok := s.world.SnitchAt()
		if !ok {
			return
		}
		options := globe.Neighbors(pos)
		best := options[rand.Intn(len(options))]
		if rand.Float64() < 0.72 { // flee: maximize distance to the nearest seeker
			bestScore := -1.0
			for _, n := range options {
				score := s.nearestSeekerDist(n)
				if score > bestScore {
					bestScore, best = score, n
				}
			}
		}
		s.kernel.Realize(g.snitch, s.world, skyra.Impulse{
			Primitive: world3d.PrimitiveMove,
			TargetID:  best.ID(),
		})
	}
}

func (s *server) nearestSeekerDist(c world3d.Cell) float64 {
	g := s.world.Globe()
	cx, cy, cz := g.Pos3D(c)
	nearest := 10.0
	for _, id := range []string{s.proxy.ID(), s.game.player.ID()} {
		if p, ok := s.world.PositionOf(id); ok {
			px, py, pz := g.Pos3D(p)
			d := (cx-px)*(cx-px) + (cy-py)*(cy-py) + (cz-pz)*(cz-pz)
			if d < nearest {
				nearest = d
			}
		}
	}
	return nearest
}

// hunt is the being's game loop: perceive, plan, chase, swipe — and when the world has
// moved on, do it again. Each Run is one full reconcile arc; between Runs the being
// re-opens its eyes, because the observation graph's snitch intel is stale by exactly
// the amount of time planning took.
func (s *server) hunt(g *game) {
	s.being.System = seekerSystem
	target := []dnav3.Triple{{Subject: "snitch:/", Predicate: "caught-by", Object: s.proxy.ID()}}

	for s.world.CaughtBy() == "" && s.gameActive(g) {
		s.being.NewEpisode(s.realm)
		seed := s.realm.Panorama()
		s.being.ObserveImpulse(seed.Effects, target)
		c, err := s.being.Run(target, 4)
		if err != nil {
			s.hub.publish(map[string]any{"type": "error", "err": err.Error()})
			return
		}
		s.hub.publish(map[string]any{"type": "hunt", "passes": c.Passes, "acts": c.Acts,
			"resolved": s.being.Metrics.Resolved, "edge": s.being.Metrics.GradeSum})
	}
}

func (s *server) gameActive(g *game) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.active
}

// referee watches for the catch and calls the game.
func (s *server) referee(g *game) {
	for {
		select {
		case <-g.stop:
			return
		case <-time.After(150 * time.Millisecond):
		}
		winner := s.world.CaughtBy()
		if winner == "" {
			continue
		}
		g.mu.Lock()
		g.active = false
		close(g.stop)
		g.mu.Unlock()
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		s.hub.publish(map[string]any{"type": "game", "phase": "over", "winner": winner})
		debug.Log("─── seeker duel over ─── caught by", winner)
		return
	}
}

// handlePlayerMove: the human's broom. Click a cell, fly there — same walk machinery,
// same tick speed, as fair as the world knows how to be.
func (s *server) handlePlayerMove(rw http.ResponseWriter, r *http.Request) {
	if !s.gameActive(s.game) || s.world.CaughtBy() != "" {
		http.Error(rw, "duel is not active", http.StatusConflict)
		return
	}
	cell := strings.TrimSpace(r.URL.Query().Get("cell"))
	c, ok := world3d.ParseCell(cell)
	if !ok || !s.world.Globe().Valid(c) {
		http.Error(rw, "bad cell", http.StatusBadRequest)
		return
	}
	s.kernel.Realize(s.game.player, s.world, skyra.Impulse{
		Primitive: world3d.PrimitiveMove,
		TargetID:  c.ID(),
	})
	rw.WriteHeader(http.StatusAccepted)
}

// handlePlayerCatch: the human's swipe, through the same world primitive the being uses.
// The outcome comes back on the event stream like everything else.
func (s *server) handlePlayerCatch(rw http.ResponseWriter, r *http.Request) {
	if !s.gameActive(s.game) || s.world.CaughtBy() != "" {
		http.Error(rw, "duel is not active", http.StatusConflict)
		return
	}
	s.kernel.Realize(s.game.player, s.world, skyra.Impulse{
		Primitive: world3d.PrimitiveCatch,
	})
	rw.WriteHeader(http.StatusAccepted)
}

func (s *server) handleGameState(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(map[string]any{
		"active":   s.gameActive(s.game),
		"caughtBy": s.world.CaughtBy(),
		"agents":   s.world.Positions(),
	})
}
