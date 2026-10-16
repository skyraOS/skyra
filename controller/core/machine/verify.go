package machine

import (
	"encoding/json"
	"errors"
	"fmt"

	"skyra/controller/core/horizon"
	"skyra/controller/core/record"
)

// Object is bytes the shell read from the tree under a content address, as it
// read them: unverified. The core checks every one before it uses it (41, V1),
// so the shell is one more emitter of untrusted bytes: it can fail or
// withhold, but it cannot pass off anything content-addressed.
type Object struct {
	Key string // the hash it was read under
	Raw []byte // nil: nothing was found under Key
}

// Found reports whether the shell returned bytes under the key.
func (o Object) Found() bool { return o.Raw != nil }

// HeldRecord is a record as a trusted holder presents it (replay, a test):
// its canonical encoding under its stamp.
func HeldRecord(r record.Record) Object {
	b, _ := record.Encode(r)
	return Object{Key: record.Sum(b), Raw: b}
}

// HeldHorizon is a Horizon as a trusted holder presents it: its encoding
// under its id.
func HeldHorizon(h horizon.Horizon) Object {
	b, _ := json.Marshal(h)
	return Object{Key: record.Sum(b), Raw: b}
}

// RecordOf verifies a record read by stamp: the bytes hash to the stamp, and
// they are the canonical encoding. Without the second check two byte strings
// could decode to one record, and replay could see a record that was never
// committed in that form.
func RecordOf(o Object) (record.Record, error) {
	var r record.Record
	if err := verify(o, &r, func() ([]byte, error) { return record.Encode(r) }); err != nil {
		return record.Record{}, fmt.Errorf("record %s: %w", record.Short(o.Key), err)
	}
	return r, nil
}

// HorizonOf verifies a Horizon read by id, as RecordOf verifies a record.
func HorizonOf(o Object) (horizon.Horizon, error) {
	var h horizon.Horizon
	if err := verify(o, &h, func() ([]byte, error) { return json.Marshal(h) }); err != nil {
		return horizon.Horizon{}, fmt.Errorf("horizon %s: %w", record.Short(o.Key), err)
	}
	return h, nil
}

func verify(o Object, v any, encode func() ([]byte, error)) error {
	switch {
	case !o.Found():
		return errors.New("not found")
	case record.Sum(o.Raw) != o.Key:
		return errors.New("does not hash to its key")
	}
	if err := json.Unmarshal(o.Raw, v); err != nil {
		return fmt.Errorf("does not decode: %w", err)
	}
	b, err := encode()
	if err != nil || string(b) != string(o.Raw) {
		return errors.New("is not its canonical encoding")
	}
	return nil
}
