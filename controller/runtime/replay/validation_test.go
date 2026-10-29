package replay

import (
	"encoding/json"
	"skyra/controller/core/boundary"
	"skyra/controller/core/machine"
	"skyra/controller/core/record"
	"skyra/controller/runtime/tree"
	"testing"
)

func TestReplayRejectsUnfinishedInvoke(t *testing.T) {
	root := t.TempDir()
	_, echoReply, _ := buildContinuation(t, root)
	reply, err := tree.Load(root, echoReply)
	if err != nil {
		t.Fatal(err)
	}
	invoke, err := tree.Load(root, reply.Predecessor)
	if err != nil {
		t.Fatal(err)
	}
	// Pretend the echo invocation itself discharged the agent's dependency.
	obs, _ := json.Marshal(machine.Observation{Frame: invoke.Predecessor, Round: invoke.Predecessor, Discharges: []string{reply.Predecessor}})
	r, _ := machine.Resolved(reply.Predecessor, obs)
	head := commitStep(t, root, obs, r)
	_, err = Replay(root, head)
	if err == nil {
		t.Fatal("replay accepted an unfinished dependency")
	}

}

func TestReplayRejectsDuplicateDischarges(t *testing.T) {
	root := t.TempDir()
	_, echoReply, _ := buildContinuation(t, root)
	reply, _ := tree.Load(root, echoReply)
	invoke, _ := tree.Load(root, reply.Predecessor)
	obs, _ := json.Marshal(machine.Observation{Frame: invoke.Predecessor, Round: invoke.Predecessor, Discharges: []string{echoReply, echoReply}})
	r, _ := machine.Resolved(echoReply, obs)
	head := commitStep(t, root, obs, r)
	if _, err := Replay(root, head); err == nil {
		t.Fatal("replay accepted duplicate discharges")
	}

}

func TestReplayChecksCompleteRoundInEmissionOrder(t *testing.T) {
	root := t.TempDir()
	if err := tree.Init(root); err != nil {
		t.Fatal(err)
	}
	reach := map[string]boundary.Surface{"agent": agentS, "echo": echoS}
	g, _ := tree.Commit(root, record.Record{Version: 2, State: record.Invoke, Reach: reach})
	operator := boundary.Boundary{Stamp: g, Frame: g, Reach: reach}
	emit := func(under boundary.Boundary, out string, fanned bool) (string, boundary.Boundary) {
		r := intakeRecord(under, out)
		stamp := commitStep(t, root, []byte(out), r)
		var h boundary.Boundary
		var err error
		if fanned {
			h, err = machine.Fanned(under, r, stamp)
		} else {
			h, err = machine.Narrow(under, r, stamp)
		}
		if err != nil {
			t.Fatal(err)
		}
		if fanned {
			if _, err := tree.PutBoundary(root, h); err != nil {
				t.Fatal(err)
			}
		}
		return stamp, h
	}
	_, agent := emit(operator, `{"to":"agent","payload":{"text":"hi"}}`, false)
	outs := []string{`{"to":"echo","payload":{"text":"a"}}`, `{"to":"echo","payload":{"text":"b"}}`}
	full := []byte("[" + outs[0] + "," + outs[1] + "]")
	hash, err := tree.Put(root, "inputs", full)
	if err != nil {
		t.Fatal(err)
	}
	var deps []string
	for _, out := range outs {
		_, echo := emit(agent, out, true)
		d, _ := emit(echo, `{"to":"`+echo.Frame+`","payload":"done"}`, false)
		deps = append(deps, d)
	}
	for name, list := range map[string][]string{"complete": deps, "omitted": deps[:1], "reordered": {deps[1], deps[0]}, "duplicate": {deps[0], deps[0]}} {
		obs, _ := json.Marshal(machine.Observation{Frame: agent.Frame, Round: agent.Stamp, Emission: hash, Discharges: list})
		r, _ := machine.Resolved(list[0], obs)
		head := commitStep(t, root, obs, r)
		_, err := Replay(root, head)
		if (err == nil) != (name == "complete") {
			t.Errorf("%s: %v", name, err)
		}
	}
	obs, _ := json.Marshal(machine.Observation{Frame: agent.Frame, Round: agent.Stamp, Discharges: deps})
	r, _ := machine.Resolved(deps[0], obs)
	head := commitStep(t, root, obs, r)
	if _, err := Replay(root, head); err == nil {
		t.Fatal("version 2 resolve missing round evidence accepted")
	}
}
