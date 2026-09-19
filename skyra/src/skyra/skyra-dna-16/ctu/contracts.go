package ctu

import (
	"context"
	"skyra-v1/src/skyra/skyra-dna-16/controller"
	"skyra-v1/src/skyra/skyra-dna-16/driver"
	"skyra-v1/src/skyra/skyra-dna-16/kernel"
	"skyra-v1/src/skyra/skyra-dna-16/protocol"
	"skyra-v1/src/skyra/skyra-dna-16/scheduler"
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
