package tree

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Failure records an I/O interruption, not a terminal discharge or permission
// to execute again. Partial external effects may already have happened.
type Failure struct {
	Stamp   string `json:"stamp"`
	Message string `json:"message"`
}

func PublishFailure(root string, frame *os.File, failure Failure) error {
	b, err := json.Marshal(failure)
	if err != nil {
		return err
	}
	hash, err := Put(root, "inputs", b)
	if err != nil {
		return err
	}
	_, err = writeOnce(frame, "failure", hash)
	return err
}

func ReadFailure(root string) (*Failure, error) {
	frames, err := os.ReadDir(filepath.Join(root, "frames"))
	if err != nil {
		return nil, err
	}
	for _, frame := range frames {
		if !frame.IsDir() {
			continue
		}
		hash, err := os.ReadFile(filepath.Join(root, "frames", frame.Name(), "failure"))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		b, err := Get(root, "inputs", string(hash))
		if err != nil {
			return nil, err
		}
		var failure Failure
		if err := json.Unmarshal(b, &failure); err != nil {
			return nil, err
		}
		return &failure, nil
	}
	return nil, nil
}
