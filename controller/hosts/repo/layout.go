package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// checkLayout is where a workspace keeps what repo.check builds and tests. The
// operator injects it when the host is provisioned (-check-layout); the model
// never supplies any of it, only which fixed action to run. The default is
// this repository's own layout, so a host started without one is unchanged.
type checkLayout struct {
	Inputs  []string      `json:"inputs"`            // repository-relative paths copied into the check snapshot; ["."] is the whole tree
	Go      *goLayout     `json:"go,omitempty"`      // absent: Go actions are unavailable
	Swift   *swiftLayout  `json:"swift,omitempty"`   // absent: Swift actions are unavailable
	Python  *pythonLayout `json:"python,omitempty"`  // absent: python.test is unavailable
	Reports *string       `json:"reports,omitempty"` // directory for check reports; absent: ideas/checks; "": none are published
}

// goLayout says where Go packages may be named: under Root, or anywhere in
// the module when Root is ".".
type goLayout struct {
	Root string `json:"root"`
}

type swiftLayout struct {
	PackagePath string `json:"package_path"`
}

type pythonLayout struct {
	Start   string `json:"start"`   // unittest discovery directory
	Pattern string `json:"pattern"` // unittest discovery file pattern, e.g. test_*.py
}

func defaultLayout() checkLayout {
	return checkLayout{
		Inputs: []string{"go.mod", "go.sum", "controller", "desktop", "docs", "skyra"},
		Go:     &goLayout{Root: "controller"},
		Swift:  &swiftLayout{PackagePath: "desktop"},
		Python: &pythonLayout{Start: "controller/blender", Pattern: "test_*.py"},
	}
}

// loadLayout reads the operator's layout, or returns the default for "".
func loadLayout(path string) (checkLayout, error) {
	if path == "" {
		return defaultLayout(), nil
	}
	f, err := os.Open(path)
	if err != nil {
		return checkLayout{}, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, 64<<10+1))
	if err != nil {
		return checkLayout{}, err
	}
	if len(data) > 64<<10 {
		return checkLayout{}, fmt.Errorf("check layout exceeds 64 KiB")
	}
	var l checkLayout
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(&l); err != nil {
		return checkLayout{}, fmt.Errorf("check layout: %w", err)
	}
	if d.More() {
		return checkLayout{}, fmt.Errorf("check layout: trailing data")
	}
	return l, l.validate()
}

func (l checkLayout) validate() error {
	if len(l.Inputs) == 0 {
		return fmt.Errorf("check layout: inputs is required")
	}
	for i, in := range l.Inputs {
		if err := layoutPath("input", in); err != nil {
			return err
		}
		if in == "." && len(l.Inputs) > 1 {
			return fmt.Errorf("check layout: input \".\" is the whole tree and stands alone")
		}
		for _, other := range l.Inputs[:i] {
			if other == in || strings.HasPrefix(in, other+"/") || strings.HasPrefix(other, in+"/") {
				return fmt.Errorf("check layout: inputs %q and %q overlap", other, in)
			}
		}
	}
	if l.Go != nil {
		if err := layoutPath("go root", l.Go.Root); err != nil {
			return err
		}
	}
	if l.Swift != nil {
		if err := layoutPath("swift package path", l.Swift.PackagePath); err != nil {
			return err
		}
	}
	if l.Python != nil {
		if err := layoutPath("python start", l.Python.Start); err != nil {
			return err
		}
		if p := l.Python.Pattern; p == "" || strings.ContainsAny(p, "/\\\x00") {
			return fmt.Errorf("check layout: python pattern must be a file pattern")
		}
		if _, err := filepath.Match(l.Python.Pattern, ""); err != nil {
			return fmt.Errorf("check layout: python pattern: %w", err)
		}
	}
	if l.Reports != nil && *l.Reports != "" {
		if err := layoutPath("reports", *l.Reports); err != nil {
			return err
		}
		if *l.Reports == "." {
			return fmt.Errorf("check layout: reports must be a directory below the root")
		}
	}
	return nil
}

// layoutPath accepts a clean, slash-separated path relative to the
// repository root that stays inside it and avoids repository and runtime
// metadata.
func layoutPath(what, p string) error {
	bad := p == "" || filepath.IsAbs(p) || strings.ContainsAny(p, "\\\x00\n") || filepath.ToSlash(filepath.Clean(p)) != p
	for _, part := range strings.Split(p, "/") {
		if part == ".." || part == ".git" || part == ".skyra" {
			bad = true
		}
	}
	if bad {
		return fmt.Errorf("check layout: %s %q must be a clean path inside the repository", what, p)
	}
	return nil
}

// goDefault is the package a Go action runs when none is named.
func (g goLayout) goDefault() string {
	if g.Root == "." {
		return "./..."
	}
	return "./" + g.Root + "/..."
}

// goAllows reports whether a requested package is the default or lies under
// the root. Escapes and shell-looking text are refused regardless.
func (g goLayout) goAllows(p string) bool {
	if strings.Contains(strings.TrimSuffix(p, "/..."), "..") || strings.ContainsAny(p, "\\\x00 \t\n:;") {
		return false
	}
	if p == g.goDefault() {
		return true
	}
	if g.Root == "." {
		return p == "." || strings.HasPrefix(p, "./")
	}
	return strings.HasPrefix(p, "./"+g.Root+"/")
}

// goPackageOf is the package path of a changed Go file, or "" if the file is
// not a Go source under the root.
func (g goLayout) goPackageOf(name string) string {
	if !strings.HasSuffix(name, ".go") || (g.Root != "." && !strings.HasPrefix(name, g.Root+"/")) {
		return ""
	}
	if dir := filepath.ToSlash(filepath.Dir(name)); dir != "." {
		return "./" + dir
	}
	return "."
}

// reportsDir is where check reports are published, or "" for none.
func (l checkLayout) reportsDir() string {
	if l.Reports == nil {
		return "ideas/checks"
	}
	return *l.Reports
}
