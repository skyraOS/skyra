// Command skyra is the stateless controller. It owns only argv: which mode,
// and the operator's flags.
//
//	skyra boot -root DIR -registry FILE [-to CAPABILITY] [-payload JSON]
//	skyra instance DIR   (internal: started only by a parent holding fd 3)
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"skyra/controller/boot"
	"skyra/controller/instance"
	"skyra/controller/trace"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: skyra boot -root DIR -registry FILE [-to CAPABILITY] [-payload JSON] | skyra instance DIR")
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "boot":
		fs := flag.NewFlagSet("boot", flag.ExitOnError)
		root := fs.String("root", "", "state directory (.skyra)")
		reg := fs.String("registry", "", "the operator's registry of surfaces and their contracts")
		to := fs.String("to", "echo", "the capability the seed addresses")
		payload := fs.String("payload", `{"text":"hello"}`, "the seed's payload (JSON)")
		fs.Parse(os.Args[2:])
		if *root == "" || *reg == "" {
			fs.Usage()
			os.Exit(2)
		}
		var real string
		if real, err = realDir(*root); err == nil {
			err = boot.Run(real, *reg, *to, json.RawMessage(*payload))
		}
	case "instance":
		trace.Who = fmt.Sprintf("inst pid %d", os.Getpid())
		if len(os.Args) != 3 {
			err = fmt.Errorf("usage: skyra instance DIR")
		} else {
			err = instance.Run(os.Args[2])
		}
	default:
		err = fmt.Errorf("unknown mode %q", os.Args[1])
	}
	if err != nil {
		trace.Logf("error    %v", err)
		os.Exit(1)
	}
}

// realDir resolves symlinks so the sandbox denies the real path
// (/tmp is /private/tmp on macOS).
func realDir(p string) (string, error) {
	if err := os.MkdirAll(p, 0o755); err != nil {
		return "", err
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(abs)
}
