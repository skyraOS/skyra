package main

import (
	"testing"

	"skyra/controller/service"
)

func TestRequestsAreCheckedBeforeTheHostIsAsked(t *testing.T) {
	for name, q := range map[string]service.Request{
		"no incarnation":     {Op: "exchange", WaitMS: 1},
		"unknown op":         {Incarnation: "i", Op: "hello", WaitMS: 1},
		"no wait":            {Incarnation: "i", Op: "exchange"},
		"too long":           {Incarnation: "i", Op: "exchange", WaitMS: 30001},
		"open no program":    {Op: "open"},
		"open with a handle": {Op: "open", Program: "p", Incarnation: "i"},
		"open delivers":      {Op: "open", Program: "p", Deliver: []string{`{}`}},
		"exchange names one": {Incarnation: "i", Program: "p", Op: "exchange", WaitMS: 1},
	} {
		if _, err := perform(q, "/private/tmp/none.sock"); err == nil {
			t.Fatalf("%s was accepted", name)
		}
	}
	if _, err := perform(service.Request{Incarnation: "i", Op: "exchange", WaitMS: 1}, ""); err == nil {
		t.Fatal("an operation with no granted service host was accepted")
	}
}

func TestAnUnreachableSessionIsGone(t *testing.T) {
	o, err := perform(service.Request{Incarnation: "i", Op: "exchange", Deliver: []string{`{}`}, WaitMS: 10}, "/private/tmp/skyra-service-none/service.sock")
	if err != nil || !o.PeerEOF || o.Delivery != "not_sent" || o.Lines == nil {
		t.Fatalf("observation: %+v %v", o, err)
	}
}

// The service host is the one socket named service.sock in a skyra-service-*
// directory: other granted sockets are not it, and two of it is no grant.
func TestTheServiceSocketIsChosenByName(t *testing.T) {
	one := "/private/tmp/skyra-service-a/service.sock"
	for raw, want := range map[string]string{
		``: "",
		`["/private/tmp/skyra-cua-x/bridge.sock"]`:               "",
		`["/private/tmp/skyra-cua-x/bridge.sock","` + one + `"]`: one,
		`["/private/tmp/other/service.sock"]`:                    "",
		`["service.sock"]`:                                       "",
	} {
		if got, err := socket(raw); err != nil || got != want {
			t.Errorf("%s: %q %v", raw, got, err)
		}
	}
	if _, err := socket(`["` + one + `","/private/tmp/skyra-service-b/service.sock"]`); err == nil {
		t.Error("two service hosts were accepted")
	}
	if _, err := socket(`not json`); err == nil {
		t.Error("a malformed grant was accepted")
	}
}
