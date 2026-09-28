package main

import (
	"io"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestStopShutsDownWriteAndObservesPeerEOF(t *testing.T) {
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		t.Fatal(err)
	}
	local := os.NewFile(uintptr(fds[0]), "local")
	peer := os.NewFile(uintptr(fds[1]), "peer")
	defer local.Close()
	defer peer.Close()
	finished := make(chan error, 1)
	go func() {
		var one [1]byte
		_, err := peer.Read(one[:])
		if err != io.EOF {
			finished <- err
			return
		}
		_, err = peer.Write([]byte("{\"method\":\"goodbye\"}\n"))
		if err == nil {
			err = peer.Close()
		}
		finished <- err
	}()
	o := operate(local, request{Incarnation: "one", Op: "stop", WaitMS: 1000})
	select {
	case err := <-finished:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("peer did not observe shutdown")
	}
	if !o.PeerEOF || o.Error != "" || o.Delivery != "shutdown" || len(o.Lines) != 1 {
		t.Fatalf("stop observation: %+v", o)
	}
	if o.ObservedAt <= time.Now().Add(-time.Second).UnixMilli() {
		t.Fatalf("stop has no observation time: %+v", o)
	}
}
