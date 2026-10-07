// Package journal records operator-side changes per conversation. Tool input
// never supplies the conversation identity or the storage path.
package journal

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Entry struct {
	Path    string `json:"path"`
	Before  string `json:"before_sha256,omitempty"`
	After   string `json:"after_sha256"`
	Existed bool   `json:"existed"`
	Content string `json:"original"`
}
type Commit struct {
	Hash   string `json:"commit"`
	Branch string `json:"branch"`
}
type Log struct {
	Files   []Entry  `json:"files"`
	Commits []Commit `json:"commits"`
}

func Session(root string) string {
	if root == "" {
		return ""
	}
	if filepath.Base(root) == ".skyra" && filepath.Base(filepath.Dir(filepath.Dir(root))) == "runs" {
		return filepath.Dir(filepath.Dir(filepath.Dir(root)))
	}
	return filepath.Clean(root)
}
func file(gitdir, session string) string {
	sum := sha256.Sum256([]byte(Session(session)))
	return filepath.Join(gitdir, "skyra-history", hex.EncodeToString(sum[:])+".json")
}
func Load(gitdir, session string) (Log, error) {
	l := Log{Files: []Entry{}, Commits: []Commit{}}
	if session == "" {
		return l, nil
	}
	f, e := os.Open(file(gitdir, session))
	if os.IsNotExist(e) {
		return l, nil
	}
	if e != nil {
		return l, e
	}
	defer f.Close()
	data, e := io.ReadAll(io.LimitReader(f, (32<<20)+1))
	if e != nil {
		return l, e
	}
	if len(data) > 32<<20 {
		return l, fmt.Errorf("conversation history exceeds 32 MiB")
	}
	e = json.Unmarshal(data, &l)
	return l, e
}
func Save(gitdir, session string, l Log) error {
	if session == "" {
		return nil
	}
	if len(l.Files) > 256 || len(l.Commits) > 256 {
		return fmt.Errorf("conversation history is limited to 256 files and commits")
	}
	data, e := json.Marshal(l)
	if e != nil {
		return e
	}
	if len(data) > 32<<20 {
		return fmt.Errorf("conversation history exceeds 32 MiB")
	}
	name := file(gitdir, session)
	if e := os.MkdirAll(filepath.Dir(name), 0700); e != nil {
		return e
	}
	f, e := os.CreateTemp(filepath.Dir(name), ".history-")
	if e != nil {
		return e
	}
	defer os.Remove(f.Name())
	if _, e = f.Write(data); e == nil {
		e = f.Sync()
	}
	closeErr := f.Close()
	if e != nil {
		return e
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(f.Name(), name)
}
func (l *Log) Write(path string, before []byte, existed bool, after []byte) {
	hash := func(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
	for i := range l.Files {
		if l.Files[i].Path == path {
			l.Files[i].After = hash(after)
			return
		}
	}
	entry := Entry{Path: path, After: hash(after), Existed: existed, Content: string(before)}
	if existed {
		entry.Before = hash(before)
	}
	l.Files = append(l.Files, entry)
}
func (l *Log) Commit(hash, branch string) {
	for _, c := range l.Commits {
		if c.Hash == hash {
			return
		}
	}
	l.Commits = append(l.Commits, Commit{hash, branch})
}
func (l Log) Summary() []map[string]any {
	out := []map[string]any{}
	for _, e := range l.Files {
		out = append(out, map[string]any{"path": e.Path, "before_sha256": e.Before, "after_sha256": e.After, "created": !e.Existed})
	}
	return out
}
