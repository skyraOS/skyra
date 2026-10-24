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
of its actor's grant, found the same way: run one call and record what was
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

## The grant an invocation runs under

The launcher renders the union of two grants: the actor's and the contract's.

Neither one bounds the other. The bound is the caller: an invocation gets
only what its caller holds. When the union is not contained in the caller's
grant, the invocation is a capability fault and the actor is never launched.
Containment compares rule by rule: each rule is held by an equal rule in the
caller's grant, or, for a path, by a `subpath` rule of the same operation on
an ancestor. Paths compare by components, never by string prefix.
