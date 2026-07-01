package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"skyra-v1/src/debug"
	"skyra-v1/src/inference"
	"skyra-v1/src/skyra"
	"skyra-v1/src/skyra/relation"
	dna "skyra-v1/src/skyra/skyra-dna"
	userdna "skyra-v1/src/skyra/user-dna"
	"skyra-v1/src/skyra/world"
)

const (
	beingID   = "skyra"
	rootDrive = "be helpful"
)

func main() {
	if err := debug.Init("logs"); err != nil {
		fmt.Fprintln(os.Stderr, "log init:", err)
		os.Exit(1)
	}
	defer debug.Close()

	kernel := relation.NewRelation()
	realm := world.New(kernel)
	being := dna.NewMembraneCell(rootDrive, dna.BuildContext{
		BeingID:  beingID,
		NewSlice: func(id string) skyra.Slice { return world.NewSlice(id, beingID) },
		Kernel:   kernel,
		Infer:    inference.CallCodex,
	})
	user := userdna.NewUserCell("michael", func(id string) skyra.Slice {
		return world.NewSlice(id, "michael")
	})
	realm.Place(user, 500, 500)
	realm.Place(being, 501, 500)
	user.Seed(being) // michael knows skyra exists — they share the world from boot

	if len(os.Args) > 1 {
		runTurn(realm, user, being, strings.Join(os.Args[1:], " "))
		return
	}

	fmt.Println("skyra ready. Type a message, or Ctrl-C to exit.")
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		utterance := strings.TrimSpace(scanner.Text())
		if utterance == "" {
			continue
		}
		runTurn(realm, user, being, utterance)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "stdin:", err)
		os.Exit(1)
	}
}

func runTurn(realm *world.World, user *userdna.UserCell, being *dna.MembraneCell, utterance string) {
	fmt.Printf("michael: %s\n", utterance)

	imp := skyra.Impulse{
		Primitive: world.PrimitivePerceive,
		ResultMap: world.MapPerception,
		Payload:   utterance,
		Amplitude: 1.0,
		Source: &skyra.ImpulseSource{
			SpaceID: user.ID(),
			Coord:   skyra.Coord{X: 500, Y: 500},
		},
	}
	user.Send(imp)

	before := lockedActionCount(being.Motor())
	realm.Observe(being, imp)

	if !worldSees(realm, being.ID()) {
		fmt.Println("skyra: not visible from michael's world position")
		return
	}

	action, ok := waitForAction(being.Motor(), before, 3*time.Minute)
	if !ok {
		fmt.Println("skyra: timed out waiting for motor")
		return
	}

	mid := being.Sensory().MidLevel()
	selfPred, otherPred := predictions(mid)
	fmt.Println("framing:", mid.Gist())
	fmt.Println("other:  ", otherPred)
	fmt.Println("self:   ", selfPred)
	fmt.Println("motor:  ", action)
}

func worldSees(realm *world.World, spaceID string) bool {
	result, ok := realm.Slice().Result().(map[string]any)
	if !ok {
		return false
	}
	perception, ok := result[world.MapPerception].(map[string]any)
	if !ok {
		return false
	}
	for _, value := range perception {
		point, ok := value.(world.PointResult)
		if !ok {
			continue
		}
		if occupant, ok := point.Content.(skyra.Space); ok && occupant.ID() == spaceID {
			return true
		}
	}
	return false
}

func waitForAction(m *dna.MotorCell, previousCount int, within time.Duration) (string, bool) {
	deadline := time.Now().Add(within)
	for time.Now().Before(deadline) {
		if lockedActionCount(m) > previousCount {
			action := lockedLastAction(m)
			return action, true
		}
		time.Sleep(150 * time.Millisecond)
	}
	return "", false
}

func lockedLastAction(m *dna.MotorCell) string {
	sl := m.Slice()
	sl.Lock()
	defer sl.Unlock()
	return m.LastAction()
}

func lockedActionCount(m *dna.MotorCell) int {
	sl := m.Slice()
	sl.Lock()
	defer sl.Unlock()
	return m.ActionCount()
}

func predictions(mid *dna.MidLevelCell) (self, other string) {
	sl := mid.Slice()
	sl.Lock()
	defer sl.Unlock()

	bestTS := -1
	var best skyra.Space
	for _, gr := range mid.Groupings() {
		mc, ok := gr.(*dna.MemoryCell)
		if !ok {
			continue
		}
		if g, ok := mc.Content().(dna.Gist); ok && g.Timestamp > bestTS {
			bestTS, best = g.Timestamp, gr
		}
	}
	mc, ok := best.(*dna.MemoryCell)
	if !ok {
		return "", ""
	}
	for _, child := range mc.Children() {
		if p, ok := child.Content().(dna.Prediction); ok {
			switch p.Role {
			case "self":
				self = p.Text
			case "other":
				other = p.Text
			}
		}
	}
	return self, other
}
