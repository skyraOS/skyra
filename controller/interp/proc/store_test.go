package proc

import (
	"fmt"
	"sync"
	"testing"
)

func TestConcurrentProceduresShareNodes(t *testing.T) {
	store := Store{Dir: t.TempDir()}
	node := Node{Call: raw(`"echo"`), In: raw(`{"text":"hi"}`), After: []string{}}
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			p := Procedure{Header: Header{Contract: &Schema{Title: fmt.Sprint(i), Type: "string"}}, Return: raw(`"done"`), Body: []string{node.ID()}}
			if err := store.Put(p, map[string]Node{node.ID(): node}); err != nil {
				t.Error(err)
				return
			}
			if _, _, err := store.Procedure(p.Header.Contract); err != nil {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()
}

func TestConcurrentProcedureInstallationHasOneWinner(t *testing.T) {
	store := Store{Dir: t.TempDir()}
	var wg sync.WaitGroup
	wins := make(chan struct{}, 16)
	c := &Schema{Type: "string"}
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			p := Procedure{Header: Header{Contract: c}, Return: raw(fmt.Sprintf(`"%d"`, i))}
			if store.Put(p, nil) == nil {
				wins <- struct{}{}
			}
		}(i)
	}
	wg.Wait()
	close(wins)
	if len(wins) != 1 {
		t.Fatalf("%d winners; want one", len(wins))
	}
	if _, _, err := store.Procedure(c); err != nil {
		t.Fatal(err)
	}
}
