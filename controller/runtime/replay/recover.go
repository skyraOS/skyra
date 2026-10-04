//go:build !js

package replay

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"skyra/controller/core/contract"
	"skyra/controller/core/record"
	"skyra/controller/runtime/tree"
)

// RecoverTree is operator-side crash reconciliation, not the strand syscall.
// The operator must supply the authorized anchor, contract and requester
// context. Discovery uses durable acceptance receipts and winning frame
// claims, never a newest record, proposed release or modification time.
// The result goes back to the owning adapter for checkpoint publication.
// It does not revive a frame or authorize repeating an external effect.
func RecoverTree(root, anchor string, leaves []string, target string, schema *contract.Schema, origin, room string) (RecordedTree, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return RecordedTree{}, err
	}
	if _, err := ReadTreeInRoom(root, anchor, leaves, target, schema, origin, room); err != nil {
		return RecordedTree{}, err
	}
	belongs := map[string]bool{anchor: true}
	active := map[string]bool{}
	var within func(string) (bool, error)
	within = func(stamp string) (bool, error) {
		if yes, ok := belongs[stamp]; ok {
			return yes, nil
		}
		if !stampName(stamp) || active[stamp] {
			return false, fmt.Errorf("invalid recovery ancestry")
		}
		active[stamp] = true
		defer delete(active, stamp)
		r, err := tree.Load(root, stamp)
		if err != nil {
			return false, err
		}
		if r.Predecessor == "" {
			belongs[stamp] = false
			return false, nil
		}
		yes, err := within(r.Predecessor)
		belongs[stamp] = yes
		return yes, err
	}
	frontier := slices.Clone(leaves)
	frames := map[string]bool{anchor: true}
	add := func(stamp string) error {
		yes, err := within(stamp)
		if err != nil || !yes {
			return err
		}
		// Validate each discovered publication before including it. Corrupt
		// evidence is unresolved recovery, never permission to retry work.
		if _, err := ReadTreeInRoom(root, anchor, []string{stamp}, target, schema, origin, room); err != nil {
			return fmt.Errorf("reconcile %s: %w", record.Short(stamp), err)
		}
		h, err := HorizonAt(root, stamp)
		if err != nil {
			return err
		}
		frames[h.Frame] = true
		frontier = append(frontier, stamp)
		return nil
	}
	entries, err := os.ReadDir(filepath.Join(root, "acceptances"))
	if err != nil {
		return RecordedTree{}, err
	}
	for _, entry := range entries {
		if !stampName(entry.Name()) {
			continue // unpublished temporary files
		}
		b, err := os.ReadFile(filepath.Join(root, "acceptances", entry.Name()))
		if err != nil {
			return RecordedTree{}, err
		}
		var receipt struct{ Stamp, Horizon string }
		if err := json.Unmarshal(b, &receipt); err != nil {
			return RecordedTree{}, err
		}
		if receipt.Stamp != entry.Name() {
			continue // content-addressed receipt object, not a published lookup
		}
		if err := add(receipt.Stamp); err != nil {
			return RecordedTree{}, err
		}
	}
	for frame := range frames {
		entries, err := os.ReadDir(filepath.Join(root, "frames", frame))
		if err != nil {
			return RecordedTree{}, err
		}
		for _, entry := range entries {
			if entry.Name() != "result" && !strings.HasPrefix(entry.Name(), "resolved-") {
				continue
			}
			b, err := os.ReadFile(filepath.Join(root, "frames", frame, entry.Name()))
			if err != nil {
				return RecordedTree{}, err
			}
			if err := add(string(b)); err != nil {
				return RecordedTree{}, err
			}
		}
	}
	return ReadTreeInRoom(root, anchor, frontier, target, schema, origin, room)
}
