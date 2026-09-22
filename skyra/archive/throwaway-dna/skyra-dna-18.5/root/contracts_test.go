package main

import (
	"context"
	"io"
	"sync"
)

type fixture struct {
	ctx       context.Context
	kernel    *Kernel
	execution *ExecutionContext
	external  *endpoint
	client    *endpoint
	echo      *endpoint
	peer      *endpoint
	result    *endpoint
	gate      *gateConnection
}

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
type gateConnection struct {
	io.ReadWriteCloser
	entered chan struct{}
	release chan struct{}
	ctx     context.Context
	once    sync.Once
}
