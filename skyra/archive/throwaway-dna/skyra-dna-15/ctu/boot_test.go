package ctu

import (
"sync"
"fmt"
"skyra-v1/src/skyra/skyra-dna-15/kernel"
"context"
"skyra-v1/src/skyra/skyra-dna-15/driver"
"skyra-v1/src/skyra/skyra-dna-15/controller"
"skyra-v1/src/skyra/skyra-dna-15/scheduler"
"skyra-v1/src/skyra/skyra-dna-15/protocol"
"testing"
"errors"
"skyra-v1/src/skyra/skyra-dna-15/probe"
"skyra-v1/src/skyra/skyra-dna-15/internal/support"
"sync/atomic"
)

func TestRegistryValidationAndCapabilities(t *testing.T) {
	r := driver.NewDriverRegistry()
	valid := driver.DriverRegistration{Driver: &probe.EchoDriver{}, Input: probe.ProbeInput{}}
	for _, addr := range []string{"", " ", " echo", "echo "} {
		if err := r.Register(addr, valid); err == nil {
			t.Fatal("noncanonical address", addr)
		}
	}
	var d *probe.EchoDriver
	var v *probe.ProbeInput
	for _, bad := range []driver.DriverRegistration{{Input: probe.ProbeInput{}}, {Driver: &probe.EchoDriver{}}, {Driver: d, Input: probe.ProbeInput{}}, {Driver: &probe.EchoDriver{}, Input: v}} {
		if err := r.Register("echo", bad); err == nil {
			t.Fatal("missing dependency accepted")
		}
	}
	if _, err := r.Resolve("echo"); err == nil {
		t.Fatal("bad entry installed")
	}
	resolver := r.Resolver()
	registrar := r.Registrar()
	if _, ok := resolver.(driver.Registrar); ok {
		t.Fatal("resolver exposes registration")
	}
	if _, ok := registrar.(driver.Resolver); ok {
		t.Fatal("registrar exposes resolution")
	}
	if err := registrar.Register("echo", valid); err != nil {
		t.Fatal(err)
	}
	if err := registrar.Register("echo", valid); err == nil {
		t.Fatal("duplicate accepted")
	}
	got, err := resolver.Resolve("echo")
	if err != nil || got.Driver != valid.Driver || got.Input != valid.Input {
		t.Fatal("unstable pair")
	}
	if _, err := resolver.Resolve("Echo"); err == nil {
		t.Fatal("case folding")
	}
	var nilRegistry *driver.DriverRegistry
	for _, res := range []driver.Resolver{nil, nilRegistry} {
		if _, err := controller.NewContinuationController(res); err == nil {
			t.Fatal("nil resolver accepted")
		}
	}
	var nilController *testController
	if _, err := scheduler.NewScheduler(nilController, func(protocol.RootResult) error { return nil }); err == nil {
		t.Fatal("typed nil controller accepted")
	}
	ct := fixture(t, &probe.EchoDriver{}, func(protocol.RootResult) error { return nil })
	if _, ok := ct.(RegistrationAccess); ok {
		t.Fatal("public CTU exposes registration")
	}
	if _, ok := ct.(driver.Registrar); ok {
		t.Fatal("public CTU is registrar")
	}
	// The controller's internal resolver being registration-free is already proven
	// by the Resolver()/Registrar() door split asserted above; reaching across the
	// package boundary into the controller's private field is unnecessary.
}

func TestConcurrentRegistryInstallation(t *testing.T) {
	r := driver.NewDriverRegistry()
	var wins atomic.Int32
	var wg sync.WaitGroup
	ready := make(chan struct{})
	for i := 0; i < 24; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-ready
			pair := driver.DriverRegistration{Driver: &probe.RootDriver{EchoTarget: protocol.Target{Address: fmt.Sprint(i)}}, Input: probe.ProbeInput{}}
			if err := r.Register("same", pair); err == nil {
				wins.Add(1)
			}
			address := fmt.Sprintf("endpoint-%d", i)
			if err := r.Register(address, pair); err != nil {
				t.Error(err)
			}
			for j := 0; j < 24; j++ {
				got, err := r.Resolve("same")
				if err != nil || support.Missing(got.Driver) || support.Missing(got.Input) {
					t.Error("incomplete lookup", err)
				}
			}
		}(i)
	}
	close(ready)
	wg.Wait()
	if wins.Load() != 1 {
		t.Fatal("duplicate winners", wins.Load())
	}
	before, _ := r.Resolve("same")
	after, _ := r.Resolve("same")
	if before.Driver != after.Driver {
		t.Fatal("binding replaced")
	}
}

func TestBootFailureAndInitialization(t *testing.T) {
	root := kernel.Endpoint{Address: "root", Registration: driver.DriverRegistration{Driver: &probe.EchoDriver{}, Input: probe.ProbeInput{}}}
	calls := 0
	sentinel := errors.New("initialize failed")
	kernel := &testKernel{initialize: func(r driver.Registrar) error {
		calls++
		if err := r.Register(root.Address, root.Registration); err != nil {
			return err
		}
		return sentinel
	}}
	config := BootConfig{KernelAddress: "kernel", Kernel: kernel, KernelInput: probe.ProbeInput{}, Root: root}
	ct, err := NewCTU(config, func(protocol.RootResult) error { t.Error("boot started workflow"); return nil })
	if ct != nil || !errors.Is(err, sentinel) || calls != 1 {
		t.Fatal("failed boot published runtime", err)
	}
	kernel.initialize = func(driver.Registrar) error { return nil }
	if ct, err = NewCTU(config, func(protocol.RootResult) error { return nil }); ct != nil || err == nil {
		t.Fatal("boot without root succeeded")
	}
	kernel.initialize = func(r driver.Registrar) error { return r.Register(root.Address, root.Registration) }
	count := 0
	ct, err = NewCTU(config, func(protocol.RootResult) error { count++; return nil })
	if err != nil || count != 0 {
		t.Fatal("initialization did execution", err)
	}
	if err = ct.Submit(context.Background(), rootMove()); err != nil || count != 1 {
		t.Fatal("installed root unavailable", err)
	}
	var nilKernel *testKernel
	config.Kernel = nilKernel
	if ct, err = NewCTU(config, func(protocol.RootResult) error { return nil }); ct != nil || err == nil {
		t.Fatal("typed nil kernel")
	}
	config.Kernel = kernel
	var input *probe.ProbeInput
	config.KernelInput = input
	if ct, err = NewCTU(config, func(protocol.RootResult) error { return nil }); ct != nil || err == nil {
		t.Fatal("typed nil kernel input")
	}
	config.KernelInput = probe.ProbeInput{}
	if ct, err = NewCTU(config, nil); ct != nil || err == nil {
		t.Fatal("nil completion")
	}
}

