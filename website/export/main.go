// Command export turns a recorded run (a .skyra/ tree) into the bundle the
// website replays: exactly the records and inputs replay reads to reach the
// head, and nothing else from the tree. It replays the run before writing it,
// so a bundle that is written is one that replays.
//
//	go run ./website/export -root /path/.skyra -title "..." -o website/runs/name.json
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf8"

	"skyra/controller/core/record"
	"skyra/controller/runtime/replay"
	"skyra/controller/runtime/tree"
)

// Bundle is one recorded run, as the website loads it.
type Bundle struct {
	Name     string            `json:"name"`
	Title    string            `json:"title"`
	Note     string            `json:"note,omitempty"`
	Recorded string            `json:"recorded"`
	Head     string            `json:"head"`
	Records  int               `json:"records"`
	Objects  map[string]string `json:"objects"` // "records/<hash>" or "inputs/<hash>" → the exact bytes, as text
}

// reading is the tree on disk, remembering what replay read from it.
type reading struct {
	root string
	read map[string][]byte
}

func (r reading) Get(kind, h string) ([]byte, error) {
	b, err := tree.Get(r.root, kind, h)
	if err == nil {
		r.read[kind+"/"+h] = b
	}
	return b, err
}

func main() {
	root := flag.String("root", "", "the run's .skyra/ directory")
	head := flag.String("head", "", "the head record (default: the one record nothing follows)")
	name := flag.String("name", "", "a short name for the run (default: the output file's)")
	title := flag.String("title", "", "what the run shows, in a line")
	note := flag.String("note", "", "a longer note shown with the run")
	recorded := flag.String("recorded", "", "the date the run was recorded")
	out := flag.String("o", "", "the bundle to write")
	flag.Parse()
	if *root == "" || *out == "" || *title == "" || *recorded == "" {
		flag.Usage()
		os.Exit(2)
	}
	if *name == "" {
		*name = trimExt(filepath.Base(*out))
	}
	if err := run(*root, *head, *name, *title, *note, *recorded, *out); err != nil {
		fmt.Fprintln(os.Stderr, "export:", err)
		os.Exit(1)
	}
}

func run(root, head, name, title, note, recorded, out string) error {
	if head == "" {
		var err error
		if head, err = leaf(root); err != nil {
			return err
		}
	}
	r := reading{root: root, read: map[string][]byte{}}
	n, err := replay.Lineage(r, head)
	if err != nil {
		return fmt.Errorf("the run does not replay: %w", err)
	}
	b := Bundle{Name: name, Title: title, Note: note, Recorded: recorded, Head: head, Records: n, Objects: map[string]string{}}
	for k, v := range r.read {
		if !utf8.Valid(v) {
			return fmt.Errorf("%s is not text", k)
		}
		b.Objects[k] = string(v)
	}
	enc, err := json.MarshalIndent(b, "", " ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(out, append(enc, '\n'), 0o644); err != nil {
		return err
	}
	fmt.Printf("%s: %d records, head %s\n", out, n, record.Short(head))
	return nil
}

// leaf is the one record no other record follows: not as its predecessor,
// and not as a discharge a resolve names. With fan-out, a sibling's last
// record is followed only by the resolve that joins its round.
func leaf(root string) (string, error) {
	entries, err := os.ReadDir(filepath.Join(root, "records"))
	if err != nil {
		return "", err
	}
	followed := map[string]bool{}
	for _, e := range entries {
		r, err := tree.Load(root, e.Name())
		if err != nil {
			return "", err
		}
		followed[r.Predecessor] = true
		for _, d := range r.Deps {
			followed[d] = true
		}
	}
	var leaves []string
	for _, e := range entries {
		if !followed[e.Name()] {
			leaves = append(leaves, e.Name())
		}
	}
	if len(leaves) != 1 {
		return "", errors.New("the tree has more than one head: pass -head")
	}
	return leaves[0], nil
}

func trimExt(s string) string {
	return s[:len(s)-len(filepath.Ext(s))]
}
