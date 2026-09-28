package handshake

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"testing"
)

// TestInheritedHelper is re-executed as the three processes of a resume: the
// host granting a frame fd, the resumed instance, and the target it runs.
func TestInheritedHelper(t *testing.T) {
	switch os.Getenv("HANDSHAKE_ROLE") {
	case "instance":
		var frames []*os.File
		if os.Getenv("HANDSHAKE_WRAP") == "inherited" {
			frames = Inherited(3, 1)
		} else {
			frames = []*os.File{os.NewFile(3, "frame")} // what skyra resume did before
		}
		cmd := helper("target", "")
		cmd.Stdout = os.Stdout
		// Make finalization eligible during the probe if the wrappers are
		// ever accidentally discarded again.
		runtime.GC()
		err := cmd.Run()
		runtime.KeepAlive(frames)
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	case "target":
		f := os.NewFile(3, "probe")
		if _, err := f.Stat(); err == nil {
			fmt.Print("leaked")
		} else {
			fmt.Print("clean")
		}
		os.Exit(0)
	}
}

func helper(role, wrap string) *exec.Cmd {
	cmd := exec.Command(os.Args[0], "-test.run=^TestInheritedHelper$")
	cmd.Env = []string{"HANDSHAKE_ROLE=" + role, "HANDSHAKE_WRAP=" + wrap}
	cmd.Stderr = os.Stderr
	return cmd
}

// A frame fd granted to a resumed instance must not reach the targets it runs.
// The os.NewFile case shows the probe detects a leak, so the test is not vacuous.
func TestInheritedFramesDoNotReachTargets(t *testing.T) {
	for wrap, want := range map[string]string{"inherited": "clean", "newfile": "leaked"} {
		t.Run(wrap, func(t *testing.T) {
			dir, err := os.Open(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			defer dir.Close()
			cmd := helper("instance", wrap)
			cmd.ExtraFiles = []*os.File{dir}
			out, err := cmd.Output()
			if err != nil {
				t.Fatal(err)
			}
			if string(out) != want {
				t.Fatalf("target reported %q, want %q", out, want)
			}
		})
	}
}
