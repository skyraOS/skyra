// Command globe mounts the dna-v3 being on the 3D globe world, over the time kernel,
// and serves a live view of it.
//
// Same experiment as cmd/dnav3 — one target with a hidden prerequisite, fresh world and
// durable memory each episode, two numbers that must both fall — except the world is a
// sphere, moving costs ticks, and you can WATCH: the walk, the perceive wave rippling
// over the surface, the beacon lighting, and the telemetry panel with the curves.
//
//	go run ./cmd/globe                  serve on :8973, run episodes from the browser
//	go run ./cmd/globe -auto -reps 5    start running immediately
//	go run ./cmd/globe -model codex     which model plans (codex | anthropic | deepseek)
package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"skyra-v1/src/debug"
	"skyra-v1/src/inference"
	dnav3 "skyra-v1/src/skyra/cognitive-agent/skyra-dna-v3"
	"skyra-v1/src/skyra/spacetime"
	"skyra-v1/src/skyra/spacetime/relation"
	"skyra-v1/src/skyra/distributed-agent/world3d"
)

//go:embed index.html
var indexHTML []byte

// The stage. Start, shrine, and beacon sit on three different faces so every leg of the
// recipe is a real walk across a stitched edge. Defaults are for face size 12 (what the
// tests use); main rescales them to whatever -face is running.
var (
	startCell  = world3d.Cell{F: 0, X: 9, Y: 6}
	shrineCell = world3d.Cell{F: 2, X: 6, Y: 6}
	beaconCell = world3d.Cell{F: 4, X: 3, Y: 3}
)

func scaleStage(n int) {
	startCell = world3d.Cell{F: 0, X: 3 * n / 4, Y: n / 2}
	shrineCell = world3d.Cell{F: 2, X: n / 2, Y: n / 2}
	beaconCell = world3d.Cell{F: 4, X: n / 4, Y: n / 4}
	playerStart = world3d.Cell{F: 5, X: n / 2, Y: n / 2}
	snitchStart = world3d.Cell{F: 1, X: n / 2, Y: n / 2}
}

func main() {
	addr := flag.String("addr", ":8973", "listen address")
	face := flag.Int("face", 20, "cells per cube-face edge")
	tickMS := flag.Int("tick", 70, "world tick in milliseconds (walk speed)")
	reps := flag.Int("reps", 5, "episodes per run")
	auto := flag.Bool("auto", false, "start running episodes immediately")
	model := flag.String("model", "codex", "planner model: codex | anthropic | deepseek")
	verbose := flag.Bool("v", false, "print presents and emissions")
	logDir := flag.String("logs", "logs/globe", "log directory (wiped each launch; empty string disables)")
	flag.Parse()
	scaleStage(*face)

	// Wired the same way the tui wires it: one Init, and every debug.Being call in the
	// globe and the kernel starts landing in <logs>/<name>/<layer>.log. The dir is wiped
	// per launch — it is a flight recorder for THIS run, not an archive.
	if *logDir != "" {
		if err := debug.Init(*logDir); err != nil {
			log.Fatalf("logs: %v", err)
		}
	}

	kernel := relation.NewRelation()
	w := world3d.New(kernel, *face, shrineCell, beaconCell)
	proxy := newBeingProxy("skyra")
	w.Place(proxy, startCell)
	realm := newGlobeRealm(w, kernel, proxy)

	h := newHub()
	w.SetOnEvent(func(e world3d.Event) {
		h.publish(map[string]any{"type": "world", "event": e})
		// the same stream the frontend sees, durable: one JSON line per world event,
		// so a run can be replayed (or diffed) after the browser tab is long gone.
		if b, err := json.Marshal(e); err == nil {
			debug.Log("world", string(b))
		}
	})
	go func() {
		for {
			time.Sleep(time.Duration(*tickMS) * time.Millisecond)
			w.Tick()
		}
	}()

	base := pickModel(*model)
	calls := 0
	infer := func(system, present string) (string, error) {
		calls++
		h.publish(map[string]any{"type": "infer", "call": calls})
		if *verbose {
			fmt.Printf("\n─── PRESENT (call %d) ───\n%s\n", calls, present)
		}
		debug.Being("skyra", "planner", fmt.Sprintf("─── PRESENT (call %d) ───\n%s", calls, present))
		out, err := base(system, present)
		if err != nil {
			debug.Being("skyra", "planner", fmt.Sprintf("─── ERROR (call %d) ───\n%v", calls, err))
			return out, err
		}
		debug.Being("skyra", "planner", fmt.Sprintf("─── EMISSION (call %d) ───\n%s", calls, out))
		if *verbose {
			fmt.Printf("─── EMISSION ───\n%s\n", out)
		}
		return out, err
	}

	being := dnav3.NewBeing(realm, infer)
	being.Aff = globeAffordances()
	being.System = globeSystem

	srv := &server{
		hub: h, world: w, realm: realm, being: being, proxy: proxy,
		kernel: kernel, game: newGame(),
	}

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.Write(indexHTML)
	})
	http.HandleFunc("/mesh", srv.handleMesh)
	http.HandleFunc("/events", h.handleSSE)
	http.HandleFunc("/run", srv.handleRun)
	http.HandleFunc("/game/start", srv.handleGameStart)
	http.HandleFunc("/game/state", srv.handleGameState)
	http.HandleFunc("/player/move", srv.handlePlayerMove)
	http.HandleFunc("/player/catch", srv.handlePlayerCatch)

	if *auto {
		go srv.runEpisodes(*reps)
	}
	log.Printf("globe: serving on %s  (open http://localhost%s)", *addr, *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func pickModel(name string) func(string, string) (string, error) {
	switch name {
	case "anthropic":
		return inference.CallAnthropic
	case "deepseek":
		return inference.CallDeepSeek
	default:
		return inference.CallCodex
	}
}

type server struct {
	hub    *hub
	world  *world3d.World
	realm  *globeRealm
	being  *dnav3.Being
	proxy  *beingProxy
	kernel skyra.Time
	game   *game

	mu      sync.Mutex
	running bool
}

// handleMesh hands the frontend everything static: the cell geometry, the landmarks,
// and where the agents stand right now.
func (s *server) handleMesh(rw http.ResponseWriter, r *http.Request) {
	g := s.world.Globe()
	type cellJSON struct {
		ID string     `json:"id"`
		F  int        `json:"f"`
		P  [3]float64 `json:"p"`
	}
	cells := make([]cellJSON, 0, 6*g.N*g.N)
	for _, c := range g.Cells() {
		x, y, z := g.Pos3D(c)
		cells = append(cells, cellJSON{ID: c.ID(), F: c.F, P: [3]float64{x, y, z}})
	}
	agents := s.world.Positions()
	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(map[string]any{
		"n":      g.N,
		"cells":  cells,
		"shrine": s.world.ShrineCell().ID(),
		"beacon": s.world.BeaconCell().ID(),
		"lit":    s.world.Lit(),
		"agents": agents,
	})
}

func (s *server) handleRun(rw http.ResponseWriter, r *http.Request) {
	reps, err := strconv.Atoi(r.URL.Query().Get("reps"))
	if err != nil || reps <= 0 || reps > 20 {
		reps = 5
	}
	s.mu.Lock()
	busy := s.running
	if !busy {
		s.running = true
	}
	s.mu.Unlock()
	if busy {
		http.Error(rw, "already running", http.StatusConflict)
		return
	}
	go func() {
		defer func() {
			s.mu.Lock()
			s.running = false
			s.mu.Unlock()
		}()
		s.runEpisodes(reps)
	}()
	rw.WriteHeader(http.StatusAccepted)
}

// runEpisodes is cmd/dnav3's loop on the globe: fresh WORLD, same MEMORY, every episode.
func (s *server) runEpisodes(reps int) {
	s.being.System = globeSystem // the beacon vocabulary (a duel may have swapped it out)
	target := []dnav3.Triple{{Subject: "beacon:/", Predicate: "status", Object: "lit"}}
	s.hub.publish(map[string]any{"type": "target", "triples": []string{"beacon:/ status lit"}, "reps": reps})

	for i := 1; i <= reps; i++ {
		s.world.ResetEpisode()
		s.world.Place(s.proxy, startCell)
		s.being.NewEpisode(s.realm)
		s.hub.publish(map[string]any{"type": "episode", "run": i})
		debug.Log(fmt.Sprintf("─── episode %d ─── fresh world, same memory", i))

		// the being opens its eyes: one panoramic wave seeds the observation graph
		seed := s.realm.Panorama()
		s.being.ObserveImpulse(seed.Effects, target)

		before := s.being.Metrics
		c, err := s.being.Run(target, 6)
		if err != nil {
			s.hub.publish(map[string]any{"type": "error", "run": i, "err": err.Error()})
			return
		}
		resolved := s.being.Metrics.Resolved - before.Resolved
		correct := s.being.Metrics.Correct - before.Correct
		edge := 0.0
		if resolved > 0 {
			edge = (s.being.Metrics.GradeSum - before.GradeSum) / float64(resolved)
		}
		status := "closed"
		if c.Unresolved {
			status = "unresolved"
		}
		s.hub.publish(map[string]any{
			"type": "run", "run": i, "passes": c.Passes, "acts": c.Acts,
			"bets": resolved, "correct": correct, "edge": edge,
			"status": status, "lit": s.world.Lit(),
		})
		debug.Log(fmt.Sprintf("run %d  passes=%d  acts=%d  bets=%d correct=%d  edge=%+.3f  %s",
			i, c.Passes, c.Acts, resolved, correct, edge, status))
	}
	s.hub.publish(map[string]any{
		"type": "report", "text": s.being.Report(), "edges": s.being.EdgeReport(),
	})
	debug.Log("\n" + s.being.Report())
	for _, l := range s.being.EdgeReport() {
		debug.Log("edge", l)
	}
}

// hub is a plain SSE broadcaster — no dependencies, no upgrade dance.
type hub struct {
	mu   sync.Mutex
	subs map[chan []byte]bool
}

func newHub() *hub { return &hub{subs: map[chan []byte]bool{}} }

func (h *hub) publish(v any) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subs {
		select {
		case ch <- b:
		default: // a stalled viewer never stalls the world
		}
	}
}

func (h *hub) handleSSE(rw http.ResponseWriter, r *http.Request) {
	fl, ok := rw.(http.Flusher)
	if !ok {
		http.Error(rw, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")

	ch := make(chan []byte, 256)
	h.mu.Lock()
	h.subs[ch] = true
	h.mu.Unlock()
	defer func() {
		h.mu.Lock()
		delete(h.subs, ch)
		h.mu.Unlock()
	}()

	fmt.Fprintf(rw, "data: {\"type\":\"hello\"}\n\n")
	fl.Flush()
	for {
		select {
		case b := <-ch:
			fmt.Fprintf(rw, "data: %s\n\n", b)
			fl.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
