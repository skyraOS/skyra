package debug

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	systemFile *os.File
	beingFiles map[string]*os.File
	baseDir    string
	mu         sync.Mutex
)

func Init(dir string) error {
	baseDir = dir
	beingFiles = make(map[string]*os.File)
	os.RemoveAll(dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(dir, "system.log"))
	if err != nil {
		return err
	}
	systemFile = f
	return nil
}

func Log(args ...any) {
	mu.Lock()
	defer mu.Unlock()
	if systemFile == nil {
		return
	}
	fmt.Fprintln(systemFile, args...)
}

func Being(name, layer string, args ...any) {
	mu.Lock()
	defer mu.Unlock()
	key := name + "/" + layer
	f, ok := beingFiles[key]
	if !ok {
		dir := filepath.Join(baseDir, name)
		os.MkdirAll(dir, 0755)
		var err error
		f, err = os.Create(filepath.Join(dir, layer+".log"))
		if err != nil {
			return
		}
		beingFiles[key] = f
	}
	fmt.Fprintln(f, args...)
}

func Close() {
	mu.Lock()
	defer mu.Unlock()
	if systemFile != nil {
		systemFile.Close()
	}
	for _, f := range beingFiles {
		f.Close()
	}
}
