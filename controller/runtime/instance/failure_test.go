package instance

import (
	"os"
	"path/filepath"
	"skyra/controller/core/horizon"
	"skyra/controller/core/machine"
	"skyra/controller/runtime/tree"
	"testing"
)

func TestDrivePublishesIOFailureWithoutFalseDischarge(t *testing.T) {
	root := t.TempDir()
	if err := tree.Init(root); err != nil {
		t.Fatal(err)
	}
	frame, err := tree.OpenFrame(root, "frame")
	if err != nil {
		t.Fatal(err)
	}
	// The input can be stored; publishing the record cannot succeed.
	if err := os.Remove(filepath.Join(root, "records")); err != nil {
		t.Fatal(err)
	}
	_, err = Drive(root, nil, []*os.File{frame}, machine.Instance{Phase: machine.Intake}, machine.Received{Under: horizon.Horizon{Stamp: "g", Frame: "g"}, In: []byte(`{"to":"missing","payload":{}}`)})
	if err == nil {
		t.Fatal("missing record directory accepted")
	}
	failure, readErr := tree.ReadFailure(root)
	if readErr != nil || failure == nil || failure.Message == "" {
		t.Fatalf("failure was not published: %+v %v", failure, readErr)
	}
	if _, err := tree.Result(root, "frame"); !os.IsNotExist(err) {
		t.Fatal("I/O failure fabricated a discharge")
	}
}
func TestTargetDiagnosticBufferIsBounded(t *testing.T) {
	var b boundedBuffer
	data := make([]byte, 128<<10)
	n, err := b.Write(data)
	if err != nil || n != len(data) || b.Len() != 64<<10 {
		t.Fatal("diagnostics are unbounded or writer short-wrote")
	}
}
