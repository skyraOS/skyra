# SurfaceID, MoveID, and fd

Date: 2026-09-19  
Status: Design discussion. Separates the concepts; unresolved relationships below
are not implementation decisions.

## Why separate them

We need distinct concepts for the registered capability being addressed, the
individual Move, and the descriptor through which a caller's state-access profile
is resolved. The earlier discussion combined `moveID / fd` into one identity.
That equivalence is not settled and must not be assumed.

## SurfaceID — the registered address

`SurfaceID` is the durable identity/address of a Surface: a bounded, registered
capability that something can invoke.

A Surface can expose a leaf endpoint or a higher-level method composed from calls
to leaf Surfaces. For example, a higher-level Codex method could coordinate
thread creation, turn execution, and tool responses. Both levels can belong to
the Codex entity while remaining separately addressable capabilities.

The composition layer can reuse the Surface abstraction. Its registration and
authority rules still need definition. In particular, the idea that a Surface can
call only its registered children is not yet an agreed rule.

## MoveID — the identity of a Move

`MoveID` identifies an individual Move. Its opcode describes the kind of Move,
such as Invoke or Return.

Use the name `MoveID` instead of a separate invocation-only identity: when the
opcode is Invoke, that MoveID can serve as the invocation ID.

Repeated calls to the same Surface are distinct Moves even when their caller and
target are identical. A SurfaceID therefore does not identify an individual call.

How a Return identifies the Invoke it answers remains to be specified. Do not
assume that its own MoveID, its parent fd, or a provider session ID supplies that
correlation without an explicit rule.

## fd — the descriptor used to resolve a caller's profile

The fd participates in state access: interpretation resolves the caller's fd and
its predefined profile, so the permitted state functions are known when used.

Keep the descriptor concept separate from both the Surface's durable address and
the Move's identity while deciding its precise scope. We have not yet settled
whether an fd names a Move, an execution, a session, or another runtime record;
nor whether its encoding or lifetime matches a MoveID.

The current implementation's fd-per-invocation-session behavior is existing code,
not a decision for this revised model.

## The proposed incoming relationship

The discussion proposed two fields available before a new Move has its own fd:

1. Its parent fd.
2. The SurfaceID that was just called.

These describe the caller relationship and the addressed Surface. They do not by
themselves distinguish repeated calls by the same parent to the same Surface.
The relationship between these fields, the MoveID, and a newly allocated fd is
the next design decision.

## Operation identity is distinct from resource identity

(Folded in from the archived `syscalls-and-fds` research note.)

An operation's identity is not the identity of a resource it creates or uses. An
OS call illustrates this: `open` creates one descriptor, `pipe` creates two,
`read`/`write` create none and act on an existing one, and `dup2` adds another
reference to a resource that already exists. One resource handle participates in
many operations; one operation can create zero, one, or several handles.

The consequence for the runtime identities above: **a Move or a committed Request
need not allocate a new resource handle.** MoveID (the operation), RequestID (the
committed authorization), and any resulting resource handle (fd, process handle,
transport handle) are separate identities on separate lifetimes, and their
cardinality is not one-to-one. Do not assume that interpreting a Move mints a
handle, or that a handle names the Move that created it.

## Questions to settle

1. What exactly does an fd reference, and what establishes and ends its lifetime?
2. Is the relationship between MoveID and fd one-to-one, one-to-many, or something
   else? When is each assigned?
3. What does `parentFD` mean for Invoke versus Return: caller, answered execution,
   or another relationship? How does an external root enter without a parent fd?
4. How is a Return correlated to the Invoke it answers?
5. How do Surface registration, composition, and the caller's profile determine
   which other Surfaces can be invoked?
6. How does a long-lived provider session relate to its individual Moves and fds?

## State-phase constraint

The agreed phase rule from [improvements.md](improvements.md) still applies:
lock what the turn needs before interpret, call interpret and commit as distinct
phases in `step`, then unlock after commit.

Interpret performs all canonical state reads, including fd/profile resolution
and authority decisions. Commit applies the complete prepared write request.
Absolutely no canonical state reads past interpret.
