package mission

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	root "skyra/controller/core/humanroot"
	"skyra/controller/core/record"
	"skyra/controller/runtime/boot"
	"skyra/controller/runtime/replay"
	"skyra/controller/runtime/tree"
)

// ExecuteLaunch is an explicit, once-only dispatch through the existing kernel
// CLI. All OS work lives here. No scratch grant or ambient registry is added.
// A crash after claiming execution leaves an unknown outcome for inspection;
// neither reopening the Mission nor calling this method again repeats it.
func (b *Bridge) ExecuteLaunch(ctx context.Context, kernel, actor string, launch *Launch) (root.Effect, error) {
	if launch == nil || launch.Dispatch.Invocation == nil {
		return root.Effect{}, fmt.Errorf("missing Launch dispatch")
	}
	call := launch.Dispatch.Invocation.ID
	if len(call) != 64 {
		return root.Effect{}, fmt.Errorf("invalid Launch call identity")
	}
	dir := filepath.Join(b.dir, "launches", string(call))
	state, err := b.Root.State()
	if err != nil {
		return root.Effect{}, err
	}
	if launch.Dispatch.Invocation.Target != b.Launch.Ref {
		return root.Effect{}, fmt.Errorf("wrong Launch endpoint")
	}
	if _, err = b.Catalogue.Typed.ValidateHandoff(launch.Dispatch, state); err != nil {
		return root.Effect{}, err
	}
	kernelBytes, err := os.ReadFile(kernel)
	if err != nil {
		return root.Effect{}, err
	}
	if "sha256:"+record.Sum(kernelBytes) != b.Launch.Contract.Binary {
		return root.Effect{}, fmt.Errorf("kernel binary differs from the pinned Launch endpoint")
	}
	if err = os.MkdirAll(dir, 0700); err != nil {
		return root.Effect{}, err
	}
	claim, err := claimExecution(dir)
	if err != nil {
		return root.Effect{}, err
	}
	reg, payload, err := b.PrepareLaunch(filepath.Join(dir, "prepared"), actor, launch)
	if err != nil {
		claim.Close()
		return root.Effect{}, err
	}
	binding, _ := json.Marshal(map[string]any{"call": call, "spec": launch.Dispatch.Passed[0].ID(), "payload": json.RawMessage(payload)})
	if err = seal(claim, dir, binding); err != nil {
		return root.Effect{}, err
	}
	runRoot := filepath.Join(dir, ".skyra")
	cmd := exec.CommandContext(ctx, kernel, "boot", "-root", runRoot, "-registry", reg, "-to", "apprentice", "-payload", string(payload))
	output, runErr := cmd.CombinedOutput()
	head, err := boot.Head(runRoot)
	if err != nil {
		return root.Effect{}, fmt.Errorf("Launch outcome unavailable: %v; kernel: %v: %s", err, runErr, output)
	}
	received, err := b.Root.Receive(call, head, func(inv root.Invocation, receipt string) ([]json.RawMessage, error) {
		if inv.ID != call || inv.Target != b.Launch.Ref || receipt != head {
			return nil, fmt.Errorf("Launch receipt belongs to another invocation")
		}
		final, err := settled(runRoot, receipt, dir, binding)
		if err != nil {
			return nil, err
		}
		switch final.State {
		case record.Return:
			return []json.RawMessage{final.Payload}, nil
		case record.Fault:
			value, _ := json.Marshal(map[string]any{"fault": final.Fault, "observation": final.Payload})
			return []json.RawMessage{value}, nil
		default:
			return nil, fmt.Errorf("Launch receipt does not settle its root invocation")
		}
	})
	if err != nil {
		return root.Effect{}, err
	}
	if runErr != nil {
		return received, fmt.Errorf("Launch returned a recorded fault: %w", runErr)
	}
	return received, nil
}

// claimExecution creates a call's claim, exclusively: a second attempt at the
// same call fails here, before any effect.
func claimExecution(dir string) (*os.File, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	claim, err := os.OpenFile(filepath.Join(dir, "execution.claim"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("already claimed or cannot be claimed: %w", err)
	}
	return claim, nil
}

// seal writes the claim's binding durably, before the kernel runs.
func seal(claim *os.File, dir string, binding []byte) error {
	_, err := claim.Write(binding)
	if err == nil {
		err = claim.Sync()
	}
	if closeErr := claim.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	folder, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer folder.Close()
	return folder.Sync()
}

// settled replays a run to its receipt and is the record that settles its
// root invocation, after checking the saved claim still binds this call.
func settled(runRoot, receipt, dir string, binding []byte) (record.Record, error) {
	if _, err := replay.Replay(runRoot, receipt); err != nil {
		return record.Record{}, err
	}
	stored, err := os.ReadFile(filepath.Join(dir, "execution.claim"))
	if err != nil || !bytes.Equal(stored, binding) {
		return record.Record{}, fmt.Errorf("execution binding changed")
	}
	final, err := tree.Load(runRoot, receipt)
	if err != nil {
		return final, err
	}
	if final.State == record.Resolve && len(final.Deps) == 1 {
		return tree.Load(runRoot, final.Deps[0])
	}
	return final, nil
}
