package main

import "testing"

func TestNoBridgeGrant(t *testing.T) {
	t.Setenv("SKYRA_CONNECTIONS", "[]")
	_, err := run(request{Operation: "context"})
	f, ok := err.(*failure)
	if !ok || f.Code != "no_bridge_grant" {
		t.Fatalf("%v", err)
	}
}
func TestUnknownOperationCannotDispatch(t *testing.T) {
	_, err := run(request{Operation: "execute"})
	f, ok := err.(*failure)
	if !ok || f.Code != "invalid_request" {
		t.Fatalf("%v", err)
	}
}
