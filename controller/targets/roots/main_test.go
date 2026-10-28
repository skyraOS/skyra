package main

import (
	"reflect"
	"testing"
)

func TestReportsWhatItMaySee(t *testing.T) {
	// The workspace is named, never the first read grant (the registry sorts them).
	got := roots("/work", "/docs:/work", "/state/c-1/scratch")
	want := result{Workspace: "/work", Read: []string{"/docs", "/work"}, Scratch: "/state/c-1/scratch"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v", got)
	}
	if got := roots("", "", ""); got.Workspace != "" || len(got.Read) != 0 || got.Scratch != "" {
		t.Fatalf("nothing granted: %+v", got)
	}
}
