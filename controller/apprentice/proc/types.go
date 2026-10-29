package proc

import (
	"encoding/json"

	"skyra/apprentice/types"
)

// EdgeType is the base type of what a node's call returned: the name it
// called and the shape of the value (decision 10). Remember records it at
// every edge where the model was resumed; a run compares against it.
func EdgeType(call json.RawMessage, value any) string {
	return types.Of(CallName(call), nil, value).Hash
}

// InputType is the base type of a procedure's input: its contract and the
// input's shape.
func InputType(c *Schema, input any) string {
	return types.Input(c.Canonical(), input).Hash
}

// InputShown is the same, as a model reads it.
func InputShown(c *Schema, input any) string {
	return types.Input(c.Canonical(), input).Presentation
}

// Present is a value's base type as a model reads it, never a hash.
func Present(call json.RawMessage, value any) string {
	return types.Of(CallName(call), nil, value).Presentation
}
