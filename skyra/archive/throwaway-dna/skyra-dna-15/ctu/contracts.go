package ctu

import (
"skyra-v1/src/skyra/skyra-dna-15/driver"
"skyra-v1/src/skyra/skyra-dna-15/protocol"
"skyra-v1/src/skyra/skyra-dna-15/scheduler"
"context"
"skyra-v1/src/skyra/skyra-dna-15/controller"
"skyra-v1/src/skyra/skyra-dna-15/kernel"
)

type CTU interface {
	Submit(ctx context.Context, move protocol.Move) error
	Snapshot(id protocol.InvocationID) (scheduler.Continuation, bool)
}

type ctu struct {
	scheduler  *scheduler.Scheduler
	controller controller.Controller
}

type BootConfig struct {
	KernelAddress string
	Kernel        kernel.Kernel
	KernelInput   driver.InputValidator
	Root          kernel.Endpoint
}

