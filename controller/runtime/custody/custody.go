// Package custody describes non-frame descriptors carried with a lineage.
// Descriptors never enter records; this manifest travels beside them at each
// handoff and is checked before a target may receive one.
package custody

import (
	"fmt"
	"os"

	"skyra/controller/core/horizon"
)

type Resource = horizon.Resource

func Validate(resources []Resource, files []*os.File) error {
	if len(resources) != len(files) {
		return fmt.Errorf("%d resources have %d descriptors", len(resources), len(files))
	}
	for i, r := range resources {
		if r.Kind != "connection" || r.Owner == "" || r.Incarnation == "" {
			return fmt.Errorf("invalid resource %d: %+v", i, r)
		}
		st, err := files[i].Stat()
		if err != nil {
			return err
		}
		if st.Mode()&os.ModeSocket == 0 {
			return fmt.Errorf("resource %d is not a socket", i)
		}
	}
	return nil
}

func Find(resources []Resource, files []*os.File, owner, incarnation string) (*os.File, error) {
	if err := Validate(resources, files); err != nil {
		return nil, err
	}
	for i, r := range resources {
		if r.Kind == "connection" && r.Owner == owner && r.Incarnation == incarnation {
			return files[i], nil
		}
	}
	return nil, fmt.Errorf("no connection for owner %s and incarnation %s", owner, incarnation)
}
