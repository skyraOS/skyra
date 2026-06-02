package main

import (
	"fmt"
	"os"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"

	"skyra-v1/src/debug"
	"skyra-v1/src/debugview"
	"skyra-v1/src/inference"
	"skyra-v1/src/reality"
	"skyra-v1/src/tui"
)

func main() {
	if err := debug.Init("logs"); err != nil {
		fmt.Fprintln(os.Stderr, "debug init:", err)
		os.Exit(1)
	}
	defer debug.Close()

	debug.Log("skyra starting")

	inputCh := make(chan string)
	outputCh := make(chan tui.ResponseMsg)

	world := createWorld(inputCh, outputCh)

	debug.Log("world created, starting TUI")
	debugview.Start(":3333", world)

	p := tea.NewProgram(
		tui.New(world, inputCh, outputCh),
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		debug.Log("tui error:", err)
		fmt.Fprintln(os.Stderr, "tui:", err)
		os.Exit(1)
	}
}

func createWorld(inputCh chan string, outputCh chan tui.ResponseMsg) *reality.Relation {
	rel, _ := reality.Impress("world", "boot")

	skyra, thread := createSkyra()
	michael := createMichael(inputCh, outputCh, rel)

	rel.World["skyra"] = skyra
	rel.World["michael"] = michael
	rel.ThreadID = thread.ID()
	rel.Threads[thread.ID()] = thread

	// the world loop — michael's Realize drives it
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				buf := make([]byte, 4096)
				n := runtime.Stack(buf, false)
				debug.Log("PANIC in world loop:", rec)
				debug.Log(string(buf[:n]))
			}
		}()
		debug.Log("world loop started")
		for {
			michael.Realize(rel)
		}
	}()

	return rel
}

func createSkyra() (*reality.Skyra, *reality.Thread) {
	skyra := &reality.Skyra{}
	skyra.Init("skyra", "skyra")
	skyra.Content = "I hold the world together. I think, respond, and relate on behalf of the system."

	skyra.ObserveFns["pass-through"] = reality.ObservePassThrough(skyra)
	skyra.ObserveFns["pick-up"] = reality.ObservePickUp(skyra)
	skyra.ObserveFns["propagate"] = reality.ObservePropagate(skyra, inference.CallDeepSeek)

	skyra.ExpressFns["compress"] = reality.ExpressCompress(skyra, inference.CallDeepSeek)
	skyra.ExpressFns["speak"] = reality.ExpressSpeak(skyra, inference.CallDeepSeek)
	skyra.ExpressFns["motor"] = reality.ExpressMotor(skyra, inference.CallDeepSeek)

	skyra.Primitives["speak"] = reality.PrimitiveSpeak(skyra)

	thread := reality.NewThread("thread:skyra:main", "skyra")
	thread.Bind(skyra.ID())
	skyra.Relationships["thread:main"] = thread

	return skyra, thread
}

func createMichael(inputCh chan string, outputCh chan tui.ResponseMsg, world *reality.Relation) *reality.User {
	michael := reality.NewUser("michael", "I build Skyra. I decide what matters.")

	cmdCh := make(chan reality.CommandOutput, 1)
	go func() {
		for cmd := range cmdCh {
			outputCh <- tui.ResponseMsg{From: cmd.From, Content: cmd.Content}
		}
	}()

	// Express — block on TUI input, set impulse, trigger world routing
	michael.ExpressFns["terminal"] = func(r *reality.Relation) *reality.Relation {
		debug.Being("michael", "express", "waiting for input...")
		msg := <-inputCh
		debug.Being("michael", "express", "received:", msg)
		if handled := reality.HandleCommand(msg, r, cmdCh); handled {
			debug.Being("michael", "express", "handled as command")
			return r
		}
		r.SetImpulse("michael", msg)
		debug.Being("michael", "express", "impulse set, routing...")
		r.Observe(r)
		debug.Being("michael", "express", "routing complete")
		return r
	}

	// Observe — read from the Relation, not from other beings' Slices.
	michael.ObserveFns["terminal"] = func(r *reality.Relation) {
		r.RLock()
		origin := r.Origin
		impulse := r.Impulse
		r.RUnlock()
		debug.Being("michael", "observe", "origin:", origin, "impulse:", impulse)

		if origin != "" && origin != "michael" && impulse != "" {
			debug.Being("michael", "observe", "delivering response from", origin)
			outputCh <- tui.ResponseMsg{From: origin, Content: impulse}
		} else {
			debug.Being("michael", "observe", "nothing to deliver")
		}
	}

	// Michael's thread binding reality
	michaelThread := reality.NewThread("thread:michael:main", "michael")
	michaelThread.Bind(michael.ID())
	world.Threads[michaelThread.ID()] = michaelThread

	return michael
}
