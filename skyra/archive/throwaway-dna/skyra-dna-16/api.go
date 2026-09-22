// Package dna is the compatibility facade over the DNA 16 component owners.
package dna

import (
	"skyra-v1/src/skyra/skyra-dna-16/codex"
	"skyra-v1/src/skyra/skyra-dna-16/controller"
	"skyra-v1/src/skyra/skyra-dna-16/ctu"
	"skyra-v1/src/skyra/skyra-dna-16/driver"
	"skyra-v1/src/skyra/skyra-dna-16/host"
	support "skyra-v1/src/skyra/skyra-dna-16/internal/support"
	"skyra-v1/src/skyra/skyra-dna-16/kernel"
	"skyra-v1/src/skyra/skyra-dna-16/probe"
	"skyra-v1/src/skyra/skyra-dna-16/protocol"
	"skyra-v1/src/skyra/skyra-dna-16/scheduler"
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
type OpcodeEvent = protocol.OpcodeEvent
type MoveBatch = protocol.MoveBatch
type DispatchResult = protocol.DispatchResult
type AdmissionFailure = protocol.AdmissionFailure
type ExecutionFailure = protocol.ExecutionFailure
type DriverTransition = protocol.DriverTransition
type RootResult = protocol.RootResult
type AuthorizationStamp = protocol.AuthorizationStamp
type Envelope = protocol.Envelope
type DriverState = protocol.DriverState
type DriverInput = protocol.DriverInput
type DriverEmission = protocol.DriverEmission
type HostCommand = protocol.HostCommand

type Transport = driver.Transport
type Driver = driver.Driver
type CommandExecutor = driver.CommandExecutor
type InputValidator = driver.InputValidator
type DriverRegistration = driver.DriverRegistration
type Resolver = driver.Resolver
type Registrar = driver.Registrar
type DriverRegistry = driver.DriverRegistry

type Controller = controller.Controller
type ContinuationController = controller.ContinuationController

type CompiledRegister = scheduler.CompiledRegister
type Assessment = scheduler.Assessment
type Continuation = scheduler.Continuation
type Scheduler = scheduler.Scheduler
type Placement = scheduler.Placement
type Handoff = scheduler.Handoff

const Waiting = scheduler.Waiting
const Satisfied = scheduler.Satisfied
const Mismatch = scheduler.Mismatch
const PlacementPending = scheduler.PlacementPending
const PlacementSucceeded = scheduler.PlacementSucceeded
const PlacementFailed = scheduler.PlacementFailed

type Kernel = kernel.Kernel
type Endpoint = kernel.Endpoint
type FixtureKernel = kernel.FixtureKernel
type CTU = ctu.CTU
type BootConfig = ctu.BootConfig

type EchoDriver = probe.EchoDriver
type RootDriver = probe.RootDriver
type ProbeOutput = probe.ProbeOutput
type ProbeInput = probe.ProbeInput

type JSONLTransport = host.JSONLTransport
type ProcessSpec = host.ProcessSpec
type ProcessHost = host.ProcessHost

type CodexPhase = codex.CodexPhase
type CodexPendingTool = codex.CodexPendingTool
type CodexInvocationState = codex.CodexInvocationState
type CodexDriver = codex.CodexDriver
type CodexInput = codex.CodexInput
type CodexLaunchPolicy = codex.CodexLaunchPolicy

const CodexStartingThread = codex.CodexStartingThread
const CodexStartingTurn = codex.CodexStartingTurn
const CodexRunning = codex.CodexRunning
const CodexWaitingChild = codex.CodexWaitingChild
const CodexCompleted = codex.CodexCompleted
const CodexFaulted = codex.CodexFaulted

var NewInvocationID = protocol.NewInvocationID
var NewDriverRegistry = driver.NewDriverRegistry
var NewContinuationController = controller.NewContinuationController
var NewScheduler = scheduler.NewScheduler
var CloneContinuation = scheduler.CloneContinuation
var ValidateEndpoint = kernel.ValidateEndpoint
var NewCTU = ctu.NewCTU
var NewJSONLTransport = host.NewJSONLTransport
var StartProcessHost = host.StartProcessHost
var NewCodexDriver = codex.NewCodexDriver
var StartCodexHost = codex.StartCodexHost

var Missing = support.Missing
var CloneObservation = support.CloneObservation
var CloneCapabilities = support.CloneCapabilities
var CloneMove = support.CloneMove
var CloneMoves = support.CloneMoves
var CloneRegister = support.CloneRegister
var CloneDriverState = support.CloneDriverState
var CloneHostCommands = support.CloneHostCommands
var EmptyRegister = support.EmptyRegister
var CompleteRegister = support.CompleteRegister
