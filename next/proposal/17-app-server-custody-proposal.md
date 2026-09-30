# 17 — The app-server in custody: proposal

Proposes changes to [17](../17-app-server-custody.md). Ids refer to the
[list](README.md).

## What changes

17's invariant ("H(t+1) = Narrow(H(t), emission(t))", execution at t+1 under
a derived Horizon) is the seed of 24, and holds. What 24 settles:

| 17 says | Now |
|---|---|
| "All of a step's authority lives in its Horizon, as structured fields" (step 1) | agreed, and taken further: the fields leave the **surface** entirely (C1). The registry row is identity plus contracts; grants are the Horizon's |
| Open, **blocking step 1's tests**: how `Narrow` bounds the new fields | **answered.** Nothing narrows by default. Every call shares its caller's Horizon (B6); a narrowing is requested through `system.narrow` (D1, D2) and may only drop what is held, or it is a horizon fault |
| "The Horizon is the Seatbelt": the profile rendered from the Horizon | agreed, and now **one profile per Horizon**, not per surface (B7) |
| the system-reads constant beside the floor, which no Horizon field can grant | one of the **exclusions** (24), with the floor and exec |
| egress: `api.openai.com` and the auth hosts in the app-server's Horizon, enforced by a proxy | the hosts are a **network grant of the Horizon**; the proxy fields leave the surface (C4) |
| a native attempt: interrupt the turn, then emit `native.<type>` for a horizon fault | interrupt the turn, then **return an observation** of the attempt (A7). Blocking at the OS stays the guarantee; the observation is the report |

## Code

17 owns no new code beyond what B7, C1 and C4 already list. Its step 7 tests
move: "`Narrow` bounds `Exec`, `Read`, `Write`, `Mach` and `Egress`" becomes
"`system.narrow` refuses to add any of them, and a Horizon's profile renders
exactly its grants plus the exclusions."

## Edits to the document

- Status: "Step 1's blocking question is answered by
  [24](../24-the-room.md): authority is shared per Horizon and narrowed only on
  request."
- In step 1, replace "the registry declares them per surface" with "the
  registry row carries none of them; they are the Horizon's".
- In step 6 (adapter), replace the `native.<type>` emission with the returned
  observation.
- Move "How `Narrow` bounds the new fields" out of Open, answered.

## Depends on

B6, B7, C1; D1 for per-thread narrowing of the app-server.
