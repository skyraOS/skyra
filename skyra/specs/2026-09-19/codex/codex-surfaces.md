# Codex Surfaces

Date: 2026-09-19  
Status: Initial Surface set for integration design.  
Transport: stdio.

These four Surfaces are the initial set for getting the integration running.
Names below use consistent lowercase, hyphenated spelling.

## start-app-server

Start a long-lived `codex app-server --stdio` process and establish the connection.

- Input: process configuration.
- Complete the `initialize` request/response and then send `initialized`.
- Output: a connection handle for subsequent Surface calls.
- Write newline-delimited JSON requests to stdin and read responses, server
  requests, and notifications from stdout.

Reuse the process and connection across threads and turns.

## check-app-server-health

Check whether the initialized App Server connection is responding.

- Input: connection handle and response timeout.
- Send `thread/loaded/list` with `{"limit": 1}` and a fresh RPC request ID.
- Correlate the response by ID; report success, an RPC error, a transport failure,
  or a timeout.
- A successful probe confirms server responsiveness, not model-provider inference
  availability.

Codex 0.154.0 has no dedicated stdio `ping` method; this uses a read-only request.

## continue-thread

Send the next input to an existing conversation.

- Input: connection handle, thread ID, next input, and applicable settings.
- Call `thread/resume` when the thread needs to be rejoined.
- Call `turn/start` with the thread ID and input.
- Handle streamed events and server requests until the turn completes.
- Output: thread identity and turn outcome.

This Surface continues a conversation with another turn. Answering a pending
`item/tool/call` is a separate response within a turn.

## start-thread

Create a new conversation and run its initial input.

- Input: connection handle, thread settings, initial input, and dynamic-tool set.
- Call `thread/start` with the thread settings and complete tool set.
- Call `turn/start` with the returned thread ID and initial input.
- Handle streamed events and server requests until the turn completes.
- Output: new thread identity and turn outcome.

Dynamic tools require experimental API opt-in during connection initialization.
Every tool addition or removal starts a new thread with the updated complete tool
set, as recorded in [codex-integration.md](codex-integration.md).

## Design references

Detailed Surface contracts and ingress/egress profiles remain to be defined.
This initial list does not remove the previously agreed fork capability; it scopes
the starting set to these four Surfaces.

- [Surface abstraction and profiles](../surface.md)
- [Codex integration decisions](codex-integration.md)
- [Versioned endpoint inventory and parameters](codex-endpoints.md)
- [Official App Server protocol](https://learn.chatgpt.com/docs/app-server#protocol)
