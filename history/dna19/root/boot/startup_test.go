package boot_test

import (
	"fmt"
	"testing"

	"skyra/history/dna19/root/boot"
	"skyra/history/dna19/root/trace"
)

func TestMalformedStartup(t *testing.T) {
	for _, args := range [][]string{nil, {"{"}, {`{"echo":"unused","observations":["one"]}`}, {`{"echo":"unused","observations":["a","b"],"authority":"B"}`}, {`{"echo":"unused","observations":["a","b"]} {}`}} {
		tr := &trace.Trace{}
		_, err := boot.Root(guard(t), args, tr)
		rejected(t, err)
		if len(tr.Snapshot()) != 0 {
			t.Fatal("startup performed work before validation")
		}
	}
}

func Example_startupEncoding() {
	b, err := boot.DecodeStartup([]string{`{"echo":"./artifacts/echo","observations":["hello","again"]}`})
	fmt.Println(b.Observations, err)
	// Output: [hello again] <nil>
}
