package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"skyra-v1/src/inference"
	"skyra-v1/src/reality"
	"skyra-v1/src/tui"
)

func main() {
	inputCh := make(chan string)
	outputCh := make(chan tui.ResponseMsg)

	world := createWorld(inputCh, outputCh)

	p := tea.NewProgram(
		tui.New(world, inputCh, outputCh),
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "tui:", err)
		os.Exit(1)
	}
}

func createWorld(inputCh chan string, outputCh chan tui.ResponseMsg) *reality.Relation {
	rel, _ := reality.Impress("world", "boot")

	skyra := createSkyra()
	michael := createMichael(inputCh, outputCh, rel)

	rel.World["skyra"] = skyra
	rel.World["michael"] = michael

	// the world loop — michael's Realize drives it
	go func() {
		for {
			michael.Realize(rel)
		}
	}()

	return rel
}

func createSkyra() *reality.Skyra {
	skyra := &reality.Skyra{}
	skyra.Init("skyra", "skyra")
	skyra.Content = "I hold the world together. I think, respond, and relate on behalf of the system."

	skyra.ObserveFns["pass-through"] = reality.ObservePassThrough(skyra)
	skyra.ObserveFns["pick-up"] = reality.ObservePickUp(skyra)
	skyra.ObserveFns["propagate"] = reality.ObservePropagate(skyra, inference.Call)

	skyra.ExpressFns["compress"] = reality.ExpressCompress(skyra, inference.Call)
	skyra.ExpressFns["motor"] = reality.ExpressMotor(skyra, func(s string) string {
		return s
	})

	return skyra
}

func createMichael(inputCh chan string, outputCh chan tui.ResponseMsg, world *reality.Relation) *reality.User {
	michael := reality.NewUser("michael", "I build Skyra. I decide what matters.")

	// Express — block on TUI input, set impulse on the Relation
	michael.ExpressFns["terminal"] = func(r *reality.Relation) *reality.Relation {
		msg := <-inputCh
		r.SetImpulse("michael", msg)
		return r
	}

	// Observe — route through the world, deliver response to TUI
	michael.ObserveFns["terminal"] = func(r *reality.Relation) {
		// world routes — activation selects the target
		r.Observe(r)

		// read the response from whoever the world routed to
		for name := range r.World {
			if name == "michael" {
				continue
			}
			sl := r.GetSlice(name)
			sl.Lock()
			var response string
			for _, t := range sl.Thoughts {
				if cell, ok := t.(*reality.Skyra); ok && cell.Content != "" {
					if response != "" {
						response += "\n"
					}
					response += cell.Content
				}
			}
			sl.Thoughts = nil
			sl.Load = 0
			sl.Unlock()

			if response != "" {
				outputCh <- tui.ResponseMsg{From: name, Content: response}
			}
		}
	}

	return michael
}
