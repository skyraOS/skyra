package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"skyra/controller/targets/user/checkpoint"
	"skyra/controller/targets/user/projection"
)

// operation is the user actor's finite read/projection surface. The calling
// interpreter composes these calls; it does not own the checkpoint or view.
func operation(raw json.RawMessage) (bool, error) {
	var e struct {
		Frame string
		Input map[string]json.RawMessage
	}
	if err := json.Unmarshal(raw, &e); err != nil {
		return false, err
	}
	if len(e.Input) != 1 {
		return false, nil
	}
	respond := func(v any) { emit(e.Frame, v) }
	decode := func(b json.RawMessage, v any) error {
		d := json.NewDecoder(bytes.NewReader(b))
		d.DisallowUnknownFields()
		return d.Decode(v)
	}
	for name, args := range e.Input {
		switch name {
		case "read-checkpoint":
			var ask struct {
				Conversation string `json:"conversation"`
			}
			if err := decode(args, &ask); err != nil {
				return true, err
			}
			state := os.Getenv("SKYRA_STATE")
			if state == "" {
				return true, fmt.Errorf("user actor has no checkpoint storage")
			}
			cp, err := (checkpoint.Store{Dir: filepath.Join(state, "checkpoints")}).Read(ask.Conversation)
			if err != nil {
				return true, err
			}
			respond(cp)
			return true, nil
		case "project-state":
			var ask struct {
				Conversation string                  `json:"conversation"`
				Availability projection.Availability `json:"availability"`
				Checkpoint   checkpoint.Checkpoint   `json:"checkpoint"`
				Trace        projection.Tree         `json:"trace"`
			}
			if err := decode(args, &ask); err != nil {
				return true, err
			}
			// Availability is required evidence from the resource custodian. A missing
			// object is not equivalent to a read-only lost conversation.
			var fields map[string]json.RawMessage
			_ = json.Unmarshal(args, &fields)
			var availability map[string]json.RawMessage
			if json.Unmarshal(fields["availability"], &availability) != nil || len(availability) != 3 {
				return true, fmt.Errorf("missing continuation availability")
			}
			for _, key := range []string{"input_live", "turn_live", "recovery_required"} {
				var b bool
				if json.Unmarshal(availability[key], &b) != nil {
					return true, fmt.Errorf("invalid continuation availability")
				}
			}
			view, err := projection.ConversationState(ask.Conversation, ask.Checkpoint.RootInvocation, ask.Trace, ask.Availability)
			if err != nil {
				return true, err
			}
			respond(view)
			return true, nil
		case "project-app-state":
			var ask struct {
				UIIncarnation string            `json:"uiIncarnation"`
				Revision      int               `json:"revision"`
				Views         []projection.View `json:"views"`
			}
			if err := decode(args, &ask); err != nil {
				return true, err
			}
			snap, err := projection.AppState(ask.UIIncarnation, ask.Revision, ask.Views)
			if err != nil {
				return true, err
			}
			respond(snap)
			return true, nil
		}
	}
	return false, nil
}
