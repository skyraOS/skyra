package codex

import (
	"skyra-v1/src/skyra/skyra-dna-16/ctu"
	"skyra-v1/src/skyra/skyra-dna-16/driver"
	"skyra-v1/src/skyra/skyra-dna-16/host"
	"skyra-v1/src/skyra/skyra-dna-16/kernel"
	"skyra-v1/src/skyra/skyra-dna-16/probe"
	"skyra-v1/src/skyra/skyra-dna-16/protocol"
)

type Transport = driver.Transport
type DriverInput = protocol.DriverInput
type DriverEmission = protocol.DriverEmission
type DriverState = protocol.DriverState
type InvocationID = protocol.InvocationID
type Move = protocol.Move
type Invoke = protocol.Invoke
type Observation = protocol.Observation
type Target = protocol.Target
type Action = protocol.Action
type Capability = protocol.Capability
type ProcessSpec = host.ProcessSpec
type CTU = ctu.CTU
type RootResult = protocol.RootResult
type EntityID = protocol.EntityID
type PossibilityID = protocol.PossibilityID
type Horizon = protocol.Horizon
type Endpoint = kernel.Endpoint
type FixtureKernel = kernel.FixtureKernel
type Driver = driver.Driver
type DriverRegistration = driver.DriverRegistration
type EchoDriver = probe.EchoDriver
type ProbeInput = probe.ProbeInput
type BootConfig = ctu.BootConfig

const Faulted = protocol.Faulted

var NewCTU = ctu.NewCTU

func testObservation(value string) protocol.Observation {
	observation := protocol.Observation{}
	observation.Content.Bytes = []byte(value)
	return observation
}
