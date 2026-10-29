# 48 — SQL schemas

## grants

A grant is the least privilege an invocation needs to return to the runtime
without an error. It is stored in terms no OS owns; the launcher on each
install renders it at execution (Seatbelt on macOS, Landlock and seccomp on
Linux).

A grant is a set of rules. A rule is one operation on one object:

| Kind | Object | Operations |
|---|---|---|
| `path` | an absolute path, `literal` or `subpath` | `read`, `list`, `write`, `create`, `delete`, `execute` |
| `address` | `host:port`, or `unix:` and an absolute socket path | `connect`, `bind` |
| `name` | a Mach service name | `lookup`, `register` |

Rules only allow. A grant with no rules lets the actor exec its own `path`
and nothing else; the reads every process needs to start (the dynamic loader,
system libraries) are the launcher's floor for its OS, not rules.

```sql
-- The vocabulary. Fixed by this proposal; a new operation is a new row here
-- and a rendering in every launcher that supports it.
CREATE TABLE operations (
    kind      TEXT NOT NULL CHECK (kind IN ('path', 'address', 'name')),
    operation TEXT NOT NULL,
    PRIMARY KEY (kind, operation)
);

-- Which launchers can enforce an operation. A launcher refuses to run a grant
-- holding an operation it has no row for: it never drops a rule it cannot
-- render, so a grant means the same thing on every install that runs it.
CREATE TABLE operation_support (
    kind      TEXT NOT NULL,
    operation TEXT NOT NULL,
    os        TEXT NOT NULL CHECK (os IN ('darwin', 'linux')),
    PRIMARY KEY (kind, operation, os),
    FOREIGN KEY (kind, operation) REFERENCES operations (kind, operation)
);

-- sha256 is the hash of the grant's rules in canonical form: one line per
-- rule, "kind\toperation\tobject\tscope\n", sorted bytewise. The runtime
-- computes it on insert. Equal rules, equal grant.
CREATE TABLE grants (
    sha256 CHAR(64) PRIMARY KEY
);

COMMENT ON TABLE grants IS
    'The least privilege an invocation needs to return to the runtime without an error.';

CREATE TABLE grant_rules (
    grant_sha256 CHAR(64) NOT NULL REFERENCES grants (sha256),
    kind         TEXT     NOT NULL,
    operation    TEXT     NOT NULL,
    object       TEXT     NOT NULL,
    scope        TEXT     NOT NULL DEFAULT 'literal'
                          CHECK (scope IN ('literal', 'subpath')),
    PRIMARY KEY (grant_sha256, kind, operation, object, scope),
    FOREIGN KEY (kind, operation) REFERENCES operations (kind, operation),
    CHECK (scope = 'literal' OR kind = 'path')
);
```

Grants and their rules are never updated or deleted: a narrower grant is a
new grant. The runtime's database role holds `INSERT` and `SELECT` on these
tables and nothing else.

Rows of `operation_support`, as far as is known today:

| Operation | darwin | linux |
|---|---|---|
| `path` `read`, `list`, `write`, `create`, `delete`, `execute` | yes | yes |
| `address` `connect`, `bind` on `host:port` | yes | TCP only |
| `address` `connect` on `unix:` | yes | to be verified (backlog 2e) |
| `name` `lookup`, `register` | yes | no |

`operation_support` keys on the operation alone, so TCP-only and unix-only
support on Linux cannot be told apart yet. Settle this when the Linux
launcher is written; until then a Linux row for `address` `connect` is not
inserted.

## actors

An actor is a binary at `path`, named by the hash of its bytes. Its grant is
the startup grant: what the binary touches from exec until it reads its input.
It is found by running the binary with no input and recording what the OS
denied. It is not the launcher's floor, which every process gets.

```sql
CREATE TABLE actors (
    sha256       CHAR(64) PRIMARY KEY,
    path         TEXT     NOT NULL,
    grant_sha256 CHAR(64) NOT NULL REFERENCES grants (sha256)
);

COMMENT ON COLUMN actors.grant_sha256 IS
    'What the binary touches from exec until it reads its input.';
```

## contracts

A contract's grant is what one invocation through the contract needs on top
of its actor's grant, found the same way: run one invocation and record what was
still denied. NULL adds nothing.

```sql
CREATE TABLE contracts (
    contract_sha256 CHAR(64) NOT NULL,
    actor_sha256    CHAR(64) NOT NULL REFERENCES actors (sha256),
    grant_sha256    CHAR(64)          REFERENCES grants (sha256),
    PRIMARY KEY (contract_sha256, actor_sha256)
);

COMMENT ON COLUMN contracts.grant_sha256 IS
    'What one invocation through this contract needs on top of its actor''s grant. NULL adds nothing.';
```

## boundaries

A boundary is a c-list: the contracts an invocation running under it may
invoke. It holds nothing else.
Each contract names its actor, and the actor and the contract each name a
grant, so a boundary needs no grant of its own.

A boundary is named by the hash of its entries in canonical form: one line per
entry, "contract_sha256\tactor_sha256\n", sorted bytewise. A narrowing is a
new boundary holding a subset of the parent's entries; a boundary is never
updated.

```sql
CREATE TABLE boundaries (
    sha256 CHAR(64) PRIMARY KEY
);

CREATE TABLE c_list (
    boundary_sha256  CHAR(64) NOT NULL REFERENCES boundaries (sha256),
    contract_sha256 CHAR(64) NOT NULL,
    actor_sha256    CHAR(64) NOT NULL,
    PRIMARY KEY (boundary_sha256, contract_sha256, actor_sha256),
    FOREIGN KEY (contract_sha256, actor_sha256)
        REFERENCES contracts (contract_sha256, actor_sha256)
);

COMMENT ON TABLE boundaries IS
    'A c-list: the contracts an invocation running under it may invoke, and nothing else.';
```

## blobs

Bytes, named by their hash: emissions, inputs, fault causes. The database
checks the name against the bytes, so a blob cannot be stored under a hash it
does not have.

```sql
CREATE TABLE blobs (
    sha256 CHAR(64) PRIMARY KEY,
    bytes  BYTEA    NOT NULL,
    CHECK (sha256 = encode(sha256(bytes), 'hex'))
);
```

## invocations

Invocations form a tree: every invocation but the root has exactly one parent.
`boundary_root` names the tree. A node is named by its path from the root: the
root is `{0}`; its second invocation is `{0,2}`; that invocation's first is
`{0,2,1}`. A child's number is its position among its parent's invocations,
so it is fixed by the parent's decisions and is the same on every replay,
whatever order invocations finish in. The parent of a node is its path
without the last element. Nodes are only added, never moved, so a path never
changes.

`boundary_sha256` is the boundary the invocation runs under: its parent's for an
ordinary invocation, a narrower one for an invocation through a narrowing.
`opened_by` is the parent's decision that opened it.

```sql
CREATE TABLE invocations (
    boundary_root    CHAR(64)  NOT NULL,
    path            INTEGER[] NOT NULL,
    parent_path     INTEGER[] GENERATED ALWAYS AS (trim_array(path, 1)) STORED,
    opened_by       INTEGER,
    boundary_sha256  CHAR(64)  NOT NULL REFERENCES boundaries (sha256),
    actor_sha256    CHAR(64)  NOT NULL,
    contract_sha256 CHAR(64)  NOT NULL,
    input_sha256    CHAR(64)  NOT NULL REFERENCES blobs (sha256),
    PRIMARY KEY (boundary_root, path),
    FOREIGN KEY (contract_sha256, actor_sha256)
        REFERENCES contracts (contract_sha256, actor_sha256),
    CHECK ((cardinality(path) = 1) = (opened_by IS NULL))
);

CREATE INDEX invocations_children ON invocations (boundary_root, parent_path);

COMMENT ON TABLE invocations IS
    'The invocation tree. A node is one invocation of one contract, named by its path from the root.';
```

## emissions

An emission is what a node's actor wrote to its channel before it exited.
The kernel knows which node it belongs to by the channel it arrived on, never
by anything in its bytes: it creates one pipe per launch and hands the actor
only the write end. A pipe cannot carry an fd and the launcher denies fork,
so the channel cannot be passed on; authority moves only through `narrow()`.

An emission row is that attribution made durable. A node's emissions are
numbered from 1, and emission `n` is decided by decision `n`. A node has at
most one emission awaiting its decision.

```sql
CREATE TABLE emissions (
    boundary_root    CHAR(64)  NOT NULL,
    path            INTEGER[] NOT NULL,
    n               INTEGER   NOT NULL CHECK (n >= 1),
    emission_sha256 CHAR(64)  NOT NULL REFERENCES blobs (sha256),
    PRIMARY KEY (boundary_root, path, n),
    FOREIGN KEY (boundary_root, path) REFERENCES invocations (boundary_root, path)
);

COMMENT ON TABLE emissions IS
    'What a node''s actor wrote to its channel, attributed to the node by the channel it arrived on.';
```

## decisions

A decision is what the system chose to do with one emission. An emission that
is exactly `{"to": <address>, "payload": <JSON>}` speaks the protocol: it asks
for an invocation. Anything else is a return to the parent, with the emission
as its value. Handling errors is the actor's job: an emission that almost
speaks the protocol is still a return, and its parent decides what it means.

- `invoke`: open one or more invocations, the node's children. The node waits
  until every child has returned or faulted; then its actor is launched again
  with their results, and what it emits next is the node's next emission.
- `return`: the node is done, and its emission is its value.
- `fault`: the node is done, with `cause_sha256` as its cause. The runtime
  faults a node only when it refuses what the emission asked for: a contract
  its boundary does not hold, or a payload that fits no contract of the actor.

A node emits nothing after it returns or faults, and nothing while a child is
open.

```sql
CREATE TABLE decisions (
    boundary_root CHAR(64)  NOT NULL,
    path         INTEGER[] NOT NULL,
    n            INTEGER   NOT NULL,
    kind         TEXT      NOT NULL CHECK (kind IN ('invoke', 'return', 'fault')),
    cause_sha256 CHAR(64)           REFERENCES blobs (sha256),
    PRIMARY KEY (boundary_root, path, n),
    FOREIGN KEY (boundary_root, path, n) REFERENCES emissions (boundary_root, path, n),
    CHECK ((kind = 'fault') = (cause_sha256 IS NOT NULL))
);

ALTER TABLE invocations ADD FOREIGN KEY (boundary_root, parent_path, opened_by)
    REFERENCES decisions (boundary_root, path, n);

COMMENT ON TABLE decisions IS
    'What the system chose to do with one emission: invoke, return or fault.';
```

## The cycle

Read, commit, execute. Processes are ephemeral: an actor runs, emits, and
exits, and nothing holds it open.

1. **Read.** The kernel reads a node's channel until the actor exits, stores
   the bytes in `blobs`, and records them with `emit(boundary_root, node,
   emission)`, `node` being the path the channel belongs to. It then spawns
   the successor with `(boundary_root, node)` and nothing else.
2. **Commit.** The successor reads the node, its boundary and its emission
   awaiting decision, and decides: `decide_invoke` if the emission speaks the
   protocol and the invocation is allowed, `decide_fault` if it speaks the
   protocol and is refused, `decide_return` otherwise.
3. **Execute.** For `invoke`, launch each child. For `return` or `fault`, once
   no sibling is open, launch the parent's actor again with its children's
   results.

A crash at any point leaves a state the next cycle can start from:

- before `emit`: the node has no emission; it is unknown, never retried
  (backlog 2c);
- after `emit`, before the decision: a new successor decides it. Deciding only
  reads and commits, so deciding again is safe, and `(path, n)` refuses a
  second commit;
- after the decision, before execute: execute is the step with effects, and
  2c's claim guards it.

## Capabilities only narrow

The runtime cannot insert into `boundaries`, `c_list`, `invocations`,
`emissions` or `decisions`. It calls functions that run as the schema's owner
and check before they insert:

- `narrow(parent, keep)` makes the boundary holding `keep`, every entry of which
  must be in `parent`. It returns the new boundary's hash; narrowing to entries
  that already form a boundary returns that one.
- `emit(root, node, emission)` records an emission of a node that is not done,
  not waiting on a child, and has no emission awaiting a decision.
- `decide_invoke(root, node, requested)` decides the node's emission as an
  `invoke` and opens the requested invocations as children, numbered after the
  node's earlier children. Each must name a contract the node's boundary holds,
  and run under a boundary holding nothing the node's does not. It returns the
  children's paths.
- `decide_return(root, node)` and `decide_fault(root, node, cause)` decide it
  as a `return` or a `fault`.

None of them can widen a boundary, so no sequence of functions the runtime can
call gives an invocation a capability its root did not hold. The root boundary
and the root invocation `{0}` are inserted by the owner when a tree begins.
Which node an emission belongs to comes from its channel; the database cannot
see channels, so that rule is the kernel's.

```sql
CREATE TYPE capability AS (
    contract_sha256 CHAR(64),
    actor_sha256    CHAR(64)
);

CREATE TYPE invocation_request AS (
    boundary_sha256  CHAR(64),
    contract_sha256 CHAR(64),
    actor_sha256    CHAR(64),
    input_sha256    CHAR(64)
);

CREATE FUNCTION narrow(parent CHAR(64), keep capability[])
RETURNS CHAR(64)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public
AS $$
DECLARE
    child CHAR(64);
BEGIN
    IF EXISTS (
        SELECT 1 FROM unnest(keep) k
        WHERE NOT EXISTS (
            SELECT 1 FROM c_list c
            WHERE c.boundary_sha256 = parent
              AND c.contract_sha256 = k.contract_sha256
              AND c.actor_sha256 = k.actor_sha256))
    THEN
        RAISE EXCEPTION 'capability fault: narrowing keeps a contract % does not hold', parent;
    END IF;

    SELECT encode(sha256(convert_to(coalesce(string_agg(line, '' ORDER BY line COLLATE "C"), ''), 'UTF8')), 'hex')
    INTO child
    FROM (SELECT DISTINCT k.contract_sha256 || E'\t' || k.actor_sha256 || E'\n' AS line
          FROM unnest(keep) k) lines;

    INSERT INTO boundaries VALUES (child) ON CONFLICT DO NOTHING;
    INSERT INTO c_list
        SELECT DISTINCT child, k.contract_sha256, k.actor_sha256 FROM unnest(keep) k
        ON CONFLICT DO NOTHING;
    RETURN child;
END
$$;

CREATE FUNCTION emit(root CHAR(64), node INTEGER[], emission CHAR(64))
RETURNS INTEGER
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public
AS $$
DECLARE
    next_n INTEGER;
BEGIN
    PERFORM 1 FROM invocations i WHERE i.boundary_root = root AND i.path = node FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'no invocation % in %', node, root;
    END IF;
    IF EXISTS (
        SELECT 1 FROM decisions d
        WHERE d.boundary_root = root AND d.path = node AND d.kind IN ('return', 'fault'))
    THEN
        RAISE EXCEPTION 'invocation % is done', node;
    END IF;
    IF EXISTS (
        SELECT 1 FROM invocations c
        WHERE c.boundary_root = root AND c.parent_path = node
          AND NOT EXISTS (
              SELECT 1 FROM decisions d
              WHERE d.boundary_root = root AND d.path = c.path
                AND d.kind IN ('return', 'fault')))
    THEN
        RAISE EXCEPTION 'invocation % is waiting on a child', node;
    END IF;
    IF EXISTS (
        SELECT 1 FROM emissions e
        WHERE e.boundary_root = root AND e.path = node
          AND NOT EXISTS (
              SELECT 1 FROM decisions d
              WHERE d.boundary_root = root AND d.path = node AND d.n = e.n))
    THEN
        RAISE EXCEPTION 'invocation % has an emission awaiting a decision', node;
    END IF;
    SELECT coalesce(max(e.n), 0) + 1 INTO next_n FROM emissions e
    WHERE e.boundary_root = root AND e.path = node;
    INSERT INTO emissions VALUES (root, node, next_n, emission);
    RETURN next_n;
END
$$;

-- pending locks a node and returns the number of its emission awaiting a
-- decision, or says there is none.
CREATE FUNCTION pending(root CHAR(64), node INTEGER[])
RETURNS INTEGER
LANGUAGE plpgsql SET search_path = public
AS $$
DECLARE
    awaiting INTEGER;
BEGIN
    PERFORM 1 FROM invocations i WHERE i.boundary_root = root AND i.path = node FOR UPDATE;
    SELECT e.n INTO awaiting FROM emissions e
    WHERE e.boundary_root = root AND e.path = node
      AND NOT EXISTS (
          SELECT 1 FROM decisions d
          WHERE d.boundary_root = root AND d.path = node AND d.n = e.n);
    IF awaiting IS NULL THEN
        RAISE EXCEPTION 'invocation % has no emission awaiting a decision', node;
    END IF;
    RETURN awaiting;
END
$$;

CREATE FUNCTION decide_invoke(root CHAR(64), node INTEGER[], requested invocation_request[])
RETURNS SETOF INTEGER[]
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public
AS $$
DECLARE
    held     CHAR(64);
    awaiting INTEGER;
    k        INTEGER;
    r        invocation_request;
BEGIN
    awaiting := pending(root, node);
    SELECT i.boundary_sha256 INTO held FROM invocations i WHERE i.boundary_root = root AND i.path = node;
    IF coalesce(cardinality(requested), 0) = 0 THEN
        RAISE EXCEPTION 'an invoke decision opens at least one invocation';
    END IF;
    FOREACH r IN ARRAY requested LOOP
        IF NOT EXISTS (
            SELECT 1 FROM c_list l
            WHERE l.boundary_sha256 = held
              AND l.contract_sha256 = r.contract_sha256
              AND l.actor_sha256 = r.actor_sha256)
        THEN
            RAISE EXCEPTION 'capability fault: % does not hold the contract', held;
        END IF;
        IF EXISTS (
            SELECT 1 FROM c_list l
            WHERE l.boundary_sha256 = r.boundary_sha256
              AND NOT EXISTS (
                  SELECT 1 FROM c_list p
                  WHERE p.boundary_sha256 = held
                    AND p.contract_sha256 = l.contract_sha256
                    AND p.actor_sha256 = l.actor_sha256))
        THEN
            RAISE EXCEPTION 'capability fault: % holds a contract % does not', r.boundary_sha256, held;
        END IF;
    END LOOP;

    INSERT INTO decisions VALUES (root, node, awaiting, 'invoke', NULL);

    SELECT count(*) INTO k FROM invocations i
    WHERE i.boundary_root = root AND i.parent_path = node;
    FOREACH r IN ARRAY requested LOOP
        k := k + 1;
        INSERT INTO invocations (boundary_root, path, opened_by, boundary_sha256,
                                 actor_sha256, contract_sha256, input_sha256)
        VALUES (root, node || k, awaiting, r.boundary_sha256,
                r.actor_sha256, r.contract_sha256, r.input_sha256);
        RETURN NEXT node || k;
    END LOOP;
END
$$;

CREATE FUNCTION decide_return(root CHAR(64), node INTEGER[])
RETURNS VOID
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public
AS $$
BEGIN
    INSERT INTO decisions VALUES (root, node, pending(root, node), 'return', NULL);
END
$$;

CREATE FUNCTION decide_fault(root CHAR(64), node INTEGER[], cause CHAR(64))
RETURNS VOID
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public
AS $$
BEGIN
    INSERT INTO decisions VALUES (root, node, pending(root, node), 'fault', cause);
END
$$;

CREATE ROLE skyra_runtime;
REVOKE ALL ON boundaries, c_list, invocations, emissions, decisions, blobs FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION narrow, emit, pending, decide_invoke, decide_return, decide_fault FROM PUBLIC;
GRANT USAGE ON SCHEMA public TO skyra_runtime;
GRANT SELECT ON boundaries, c_list, invocations, emissions, decisions, blobs TO skyra_runtime;
GRANT INSERT ON blobs TO skyra_runtime;
GRANT EXECUTE ON FUNCTION narrow, emit, decide_invoke, decide_return, decide_fault TO skyra_runtime;
```

## The grant an invocation runs under

An invocation names a contract. The kernel looks it up in the parent's
boundary; a contract the boundary does not hold is a capability fault, and the
actor is never launched. Holding the contract is the authority: there is no
second check against a grant of the parent's.

From the contract the kernel reads its actor, and the launcher renders the
union of two grants, the actor's and the contract's.
