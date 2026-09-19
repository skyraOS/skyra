// Package dna preserves the v14 public API over the component packages.
package dna

import (
	"skyra-v1/src/skyra/skyra-dna-v14/controller"
	"skyra-v1/src/skyra/skyra-dna-v14/ctu"
	"skyra-v1/src/skyra/skyra-dna-v14/driver"
	"skyra-v1/src/skyra/skyra-dna-v14/probe"
	"skyra-v1/src/skyra/skyra-dna-v14/protocol"
	"skyra-v1/src/skyra/skyra-dna-v14/scheduler"
)

type EngramID = protocol.EngramID
type InvocationID = protocol.InvocationID
type PossibilityID = protocol.PossibilityID
type Observation = protocol.Observation
type Action = protocol.Action
type EntityID = protocol.EntityID
type HostID = protocol.HostID
type Contract = protocol.Contract
type State = protocol.State
type Move = protocol.Move
type Target = protocol.Target
type Opcode = protocol.Opcode
type Invoke = protocol.Invoke
type Return = protocol.Return
type Capability = protocol.Capability
type Horizon = protocol.Horizon
type Register = protocol.Register
type RegisterEntry = protocol.RegisterEntry
type ChildInvocation = protocol.ChildInvocation
type ChildObservation = protocol.ChildObservation
type CompiledRegister = protocol.CompiledRegister
type Assessment = protocol.Assessment
type Continuation = protocol.Continuation
type Scheduler = scheduler.Scheduler
type OpcodeEvent = protocol.OpcodeEvent
type MoveBatch = protocol.MoveBatch
type DispatchResult = protocol.DispatchResult
type AdmissionFailure = protocol.AdmissionFailure
type ExecutionFailure = protocol.ExecutionFailure
type RootResult = protocol.RootResult
type Controller = controller.Controller
type ContinuationController = controller.ContinuationController
type Driver = driver.Driver
type Authority = controller.Authority
type InputValidator = driver.InputValidator
type DriverRegistration = driver.DriverRegistration
type DriverRegistry = driver.DriverRegistry
type Resolver = driver.Resolver
type Registrar = driver.Registrar
type PrivilegedCTU = driver.PrivilegedCTU
type AdmitFunc = driver.AdmitFunc
type FaultFunc = driver.FaultFunc
type CTU = ctu.CTU
type EchoDriver = probe.EchoDriver
type RootDriver = probe.RootDriver
type ProbeOutput = probe.ProbeOutput
type ProbeInput = probe.ProbeInput
type ProbeAuthority = probe.ProbeAuthority

const Dispatched = protocol.Dispatched
const InFlight = protocol.InFlight
const Faulted = protocol.Faulted
const Waiting = protocol.Waiting
const Satisfied = protocol.Satisfied
const Mismatch = protocol.Mismatch

func NewInvocationID() (InvocationID, error) { return protocol.NewInvocationID() }

func NewDriverRegistry() *DriverRegistry { return driver.NewDriverRegistry() }

func NewContinuationController(registry *DriverRegistry, authority Authority) (*ContinuationController, error) {
	return controller.NewContinuationController(registry, authority)
}

func NewScheduler(cc Controller, complete func(RootResult) error) (*Scheduler, error) {
	return scheduler.NewScheduler(cc, complete)
}

func NewCTU(rootAddress string, root DriverRegistration, authority Authority, complete func(RootResult) error) (CTU, error) {
	return ctu.NewCTU(rootAddress, root, authority, complete)
}
