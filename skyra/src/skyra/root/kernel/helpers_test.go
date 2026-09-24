package kernel_test

import (
	"context"
	"io"
	"sync"

	kt "skyra-v1/src/skyra/root/kernel/kerneltest"
)

var (
	guard      = kt.Guard
	must       = kt.Must
	rejected   = kt.Rejected
	await      = kt.Await
	newFixture = kt.New
	invoke     = kt.Invoke
	exchange   = kt.Exchange
	runAsync   = kt.RunAsync
	returned   = kt.Returned
	traceCount = kt.TraceCount
	liveLoop   = kt.LiveLoop
)

type readSignalConnection struct {
	io.ReadWriteCloser
	entered chan struct{}
	once    sync.Once
}
type waitSignalContext struct {
	context.Context
	entered chan struct{}
	once    sync.Once
}

func (r *readSignalConnection) Read(data []byte) (int, error) {
	r.once.Do(func() { close(r.entered) })
	return r.ReadWriteCloser.Read(data)
}
func (c *waitSignalContext) Done() <-chan struct{} {
	c.once.Do(func() { close(c.entered) })
	return c.Context.Done()
}
