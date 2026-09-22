// Package dna is the compatibility facade over the DNA 15 component owners.
package dna
import (
support "skyra-v1/src/skyra/skyra-dna-15/internal/support"
"skyra-v1/src/skyra/skyra-dna-15/protocol"
"skyra-v1/src/skyra/skyra-dna-15/scheduler"
"skyra-v1/src/skyra/skyra-dna-15/driver"
"skyra-v1/src/skyra/skyra-dna-15/controller"
"skyra-v1/src/skyra/skyra-dna-15/ctu"
"skyra-v1/src/skyra/skyra-dna-15/kernel"
"skyra-v1/src/skyra/skyra-dna-15/probe"
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
const Dispatched = protocol.Dispatched
const InFlight = protocol.InFlight
const Faulted = protocol.Faulted
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
type CompiledRegister = scheduler.CompiledRegister
type Assessment = scheduler.Assessment
const Waiting = scheduler.Waiting
const Satisfied = scheduler.Satisfied
const Mismatch = scheduler.Mismatch
type Continuation = scheduler.Continuation
type OpcodeEvent = protocol.OpcodeEvent
type MoveBatch = protocol.MoveBatch
type DispatchResult = protocol.DispatchResult
type AdmissionFailure = protocol.AdmissionFailure
type ExecutionFailure = protocol.ExecutionFailure
type RootResult = protocol.RootResult
type Driver = driver.Driver
type InputValidator = driver.InputValidator
type DriverRegistration = driver.DriverRegistration
type Resolver = driver.Resolver
type Registrar = driver.Registrar
type DriverRegistry = driver.DriverRegistry
type Controller = controller.Controller
type ContinuationController = controller.ContinuationController
type Scheduler = scheduler.Scheduler
type CTU = ctu.CTU
type EchoDriver = probe.EchoDriver
type RootDriver = probe.RootDriver
type ProbeOutput = probe.ProbeOutput
type ProbeInput = probe.ProbeInput
type Placement = scheduler.Placement
const PlacementPending = scheduler.PlacementPending
const PlacementSucceeded = scheduler.PlacementSucceeded
const PlacementFailed = scheduler.PlacementFailed
type Handoff = scheduler.Handoff
type AuthorizationStamp = protocol.AuthorizationStamp
type Envelope = protocol.Envelope
type Kernel = kernel.Kernel
type Endpoint = kernel.Endpoint
type FixtureKernel = kernel.FixtureKernel
type BootConfig = ctu.BootConfig
var CloneContinuation = scheduler.CloneContinuation
var NewScheduler = scheduler.NewScheduler
var NewContinuationController = controller.NewContinuationController
var CloneObservation = support.CloneObservation
var EmptyRegister = support.EmptyRegister
var CompleteRegister = support.CompleteRegister
var CloneRegister = support.CloneRegister
var NewCTU = ctu.NewCTU
var NewInvocationID = protocol.NewInvocationID
var Missing = support.Missing
var CloneMove = support.CloneMove
var NewDriverRegistry = driver.NewDriverRegistry
var CloneCapabilities = support.CloneCapabilities
