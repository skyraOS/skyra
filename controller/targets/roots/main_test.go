package main

import (
	"reflect"
	"testing"
)

func TestReportsWhatItMaySee(t *testing.T) {
	got := roots("/work:/docs", "/state/c-1/scratch")
	want := result{Workspace: "/work", Read: []string{"/work", "/docs"}, Scratch: "/state/c-1/scratch"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v", got)
	}
	if got := roots("", ""); got.Workspace != "" || len(got.Read) != 0 || got.Scratch != "" {
		t.Fatalf("nothing granted: %+v", got)
	}
}
