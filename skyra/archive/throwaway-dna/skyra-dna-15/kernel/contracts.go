package kernel

import (
"skyra-v1/src/skyra/skyra-dna-15/driver"
)

// Kernel initialization receives registration only and must not start execution.
type Kernel interface {
	driver.Driver
	Initialize(driver.Registrar) error
}

type Endpoint struct {
	Address      string
	Registration driver.DriverRegistration
}

// Root is supplied separately; Initialize installs it and the fixed local list.
type FixtureKernel struct {
	Root  Endpoint
	Local []Endpoint
}

