package protocol

import (
	"encoding/json"
	"fmt"
)

func (m Move) Validate() error {
	n := 0
	if m.Invoke != nil {
		n++
	}
	if m.Return != nil {
		n++
	}
	if m.Fault != nil {
		n++
	}
	if n != 1 || m.Identity == 0 {
		return fmt.Errorf("expected one Move variant and nonzero identity")
	}
	return nil
}

func (m Move) Clone() Move {
	b, _ := json.Marshal(m)
	var copy Move
	_ = json.Unmarshal(b, &copy)
	return copy
}

func Arguments(operation, text string) []Observation {
	return []Observation{{Key: "operation", Bytes: []byte(operation)}, {Key: "text", Bytes: []byte(text)}}
}

func CheckArguments(args []Observation, operation string) error {
	if len(args) != 2 || args[0].Key != "operation" || string(args[0].Bytes) != operation || args[1].Key != "text" {
		return fmt.Errorf("expected explicit %s operation and text observation", operation)
	}
	return nil
}
