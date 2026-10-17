package tree

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"skyra/controller/core/horizon"
)

// Receipt publishes the acceptance of stamp under the Horizon id, after the
// successor has validated its release and before it acknowledges custody or
// performs any target effect: a committed Horizon alone is a proposed release,
// not proof of acceptance. The caller has already checked the id (the core
// does, 41 V1). It reports whether the same receipt was already there; a
// different one is ErrConflict.
func Receipt(root, stamp, id string) (existed bool, err error) {
	dir := filepath.Join(root, "acceptances")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return false, err
	}
	// Put first makes the complete receipt durable, then link publishes its
	// lookup by accepted record. Concurrent publishers cannot replace it.
	b, _ := json.Marshal(struct{ Stamp, Horizon string }{stamp, id})
	hash, err := Put(root, "acceptances", b)
	if err != nil {
		return false, err
	}
	name := filepath.Join(dir, stamp)
	if err := os.Link(filepath.Join(dir, hash), name); err != nil {
		if !os.IsExist(err) {
			return false, err
		}
		old, err := os.ReadFile(name)
		if err != nil {
			return true, err
		}
		if string(old) != string(b) {
			return true, fmt.Errorf("%w: acceptance for %s", ErrConflict, stamp)
		}
		existed = true
	}
	return existed, syncDir(dir, "acceptances dir")
}

// AcceptedHorizon reads a receipt and its content-addressed release. Records,
// releases and receipts are separate: receipts never supply new authority.
func AcceptedHorizon(root, stamp string) (horizon.Horizon, error) {
	if !validHash(stamp) {
		return horizon.Horizon{}, fmt.Errorf("invalid accepted stamp")
	}
	b, err := os.ReadFile(filepath.Join(root, "acceptances", stamp))
	if err != nil {
		return horizon.Horizon{}, err
	}
	var receipt struct{ Stamp, Horizon string }
	if err := json.Unmarshal(b, &receipt); err != nil {
		return horizon.Horizon{}, err
	}
	if receipt.Stamp != stamp || !validHash(receipt.Horizon) {
		return horizon.Horizon{}, fmt.Errorf("invalid acceptance receipt")
	}
	h, err := LoadHorizon(root, receipt.Horizon)
	if err != nil {
		return h, err
	}
	if h.Stamp != stamp {
		return h, fmt.Errorf("acceptance Horizon belongs to another record")
	}
	return h, nil
}

func validHash(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
