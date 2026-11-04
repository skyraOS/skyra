// treeview replays a tree's log through Step and shows the tree forming:
// in the terminal, one event at a time, or with -html as a page to step
// through.
//
//	treeview [-html FILE | -json] RUN
//
// RUN is a run's .skyra directory, or the apprentice's state under it.
package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"skyra/apprentice/shell"
)

//go:embed view.html
var page string

func main() {
	html := flag.String("html", "", "write the view as a page to this file")
	asJSON := flag.Bool("json", false, "print the view as JSON")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: treeview [-html FILE | -json] RUN")
		os.Exit(2)
	}
	dir := flag.Arg(0)
	if _, err := os.Stat(filepath.Join(dir, "apprentice", "tree")); err == nil {
		dir = filepath.Join(dir, "apprentice")
	}
	runs, err := shell.View(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "treeview:", err)
		os.Exit(1)
	}
	if *asJSON {
		b, _ := json.Marshal(runs)
		os.Stdout.Write(append(b, '\n'))
		return
	}
	if *html != "" {
		b, _ := json.Marshal(runs)
		out := strings.Replace(page, "/*RUNS*/[]", strings.ReplaceAll(string(b), "</", `<\/`), 1)
		if err := os.WriteFile(*html, []byte(out), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "treeview:", err)
			os.Exit(1)
		}
		fmt.Println(*html)
		return
	}
	for _, r := range runs {
		text(r)
	}
}

func text(r shell.Run) {
	fmt.Printf("%s root, holding %s: %s\n", r.Kind, strings.Join(r.Reach, ", "), r.Goal)
	for i, f := range r.Frames {
		fmt.Printf("\n── %d. %s %s", i+1, f.Event, f.Agent)
		if f.Detail != "" {
			fmt.Printf(": %s", clip(f.Detail, 160))
		}
		fmt.Println()
		for _, n := range f.Nodes {
			mark := "  "
			if n.ID == f.Agent {
				mark = "▸ "
			}
			fmt.Printf("%s%s%s  %s %s, holds %s, turn %d, %d names", mark, strings.Repeat("   ", n.Depth), n.ID, n.Kind, n.Phase, strings.Join(n.Holds, "+"), n.Turns, n.Names)
			if n.Why != "" {
				fmt.Printf(" (%s)", n.Why)
			}
			if n.Returned > 0 {
				fmt.Printf(", returned %d", n.Returned)
			}
			for _, l := range n.Lost {
				fmt.Printf("\n%s%s   lost %s", mark, strings.Repeat("   ", n.Depth), l)
			}
			fmt.Println()
		}
	}
}

func clip(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n]) + "…"
	}
	return s
}
