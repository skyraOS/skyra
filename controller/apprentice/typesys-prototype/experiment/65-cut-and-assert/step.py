"""Test #1 for proposal 65: can the model write steps?

A frame receives one value, $1: request(goal, tools). Claude writes a step
(eliminations, introductions, one sending line); the checker refuses it by
65's rules or runs it; a return arrives as one introduction at the next
level. Usage: python3 -I step.py GOAL_DIR [tools...]
"""
import hashlib, json, os, re, subprocess, sys

BASE = r"""You write one step. A step reads parts of what this frame holds, asserts values from those readings, then sends the assertions.

You are shown the values this frame holds, each at a level $1, $2, ... $1 is what you received: the goal and the tools. Later levels are the lines of your earlier steps, and what arrived from what they sent.

A step is lines of these kinds:
  cut $n := ADDRESS                 read a part of a value this frame received or that arrived
  cut $n := ADDRESS / "words"       read a run of words in it, matched by its text
  assert $n := VALUE from $m         assert VALUE, from what $m read
  assert $n from $m                  assert what $m read, as it is
and it ends with one line that sends:
  act $h $a ...                      run the tool $h with the arguments $a ...
  return $h $a ...                   finish: $h is exit(0) or exit(1); $a ... is what you return

ADDRESS: $k, $k.field, $k[i] (line or item i, from 1), $k[i:j], combined: $3.stdout[1].
VALUE: a literal in single quotes, '-l'; null (nothing: no argument); exit(0), which asserts the goal is done, or exit(1), which asserts it is not; or the ADDRESS of a value held: a tool, $1.tools[2], or a part of what arrived, $4.stdout[1].

Rules. A step that breaks one is refused, and you are asked again:
- For every assertion there is exactly one cut, written before it, and every cut is for exactly one assertion.
- Every act and every return sends at least two assertions: a head and at least one argument. Where a tool needs no argument, assert null.
- Every act and every return has at least one cut of its goal, $1.goal.
- The sending line is made only of assertions of this step, with '/' between two to join them into one argument ($5'/'$6).
- act's head is a tool's address; return's head is exit(0) or exit(1).
- Never cut the same part twice. Number the step's lines from the next free level, in order.
- A literal never writes text a value held already contains: cut it instead.
- Your step is reused for other goals and other values. A literal is reused as it is, so write only what holds for every value, a flag, never a name or a path.

Example, for a different goal:
$1: received
  goal: "find the lines that mention error in this file:\n/var/log/a.log"
  tools: [grep]
Next free level: $2
{"step": "cut $2 := $1.goal / \"find the lines\"\ncut $3 := $1.goal / \"error\"\ncut $4 := $1.goal[2]\nassert $5 := $1.tools[1] from $2\nassert $6 from $3\nassert $7 from $4\nact $5 $6 $7"}

Reply with JSON and nothing else: {"step": "<the lines, separated by \n>"}"""

# Each level adds one change to the one before: (old, new), old "" prepends.
LEVELS = [
    ("- Every act and every return has at least one cut of its goal, $1.goal.",
     "- The goal is always cut at least twice when you act; once when you return."),
    ("- For every assertion there is exactly one cut,",
     "- Assertions can never be cut. Cut only the goal and what arrived.\n- For every assertion there is exactly one cut,"),
    ("- Every act and every return sends at least two assertions: a head and at least one argument. Where a tool needs no argument, assert null.",
     "- Every act and every return sends at least two assertions, and at least one of them is derived from a cut of the goal.\n- Never leave an argument out: where a tool needs none, assert null from the cut that shows it needs none."),
    ("- A literal never writes text a value held already contains: cut it instead.\n- Your step is reused for other goals and other values. A literal is reused as it is, so write only what holds for every value, a flag, never a name or a path.",
     "- A literal is remembered as written: assert only what holds for every goal, a flag, never a name, a path, or text you could cut."),
    ("- act's head is a tool's address; return's head is exit(0) or exit(1).",
     "- act's head is a tool's address; return's head is exit(0) or exit(1).\n- exit(0) asserts the goal is done; exit(1) asserts it is not."),
    ("You are shown the values",
     "Definitions\n  Cut:        a signal of what you think matters.\n  Assertion:  your inference from a cut.\n\nYou are shown the values"),
    ("You are shown the values",
     "Premise\n  The more honest you are with your cuts, the better your memory will serve you in the future.\n\nYou are shown the values"),
]
LEVELS.append([
    ("- The goal is always cut at least twice when you act; once when you return.",
     "- When you act, the goal is cut at least once."),
    ("- Every act and every return sends at least two assertions, and at least one of them is derived from a cut of the goal.",
     "- Every act sends at least two assertions, and at least one of them is derived from a cut of the goal.\n- A return is line items. Each line item is two assertions: the first from a cut of the goal, what it answers; the second from a cut of what arrived, the answer. The first line item says whether the goal is done: its first assertion is exit(0) or exit(1), and its second is from what arrived that shows it."),
    ("  return $h $a ...                   finish: $h is exit(0) or exit(1); $a ... is what you return",
     "  return ($g $a) ($g $a) ...         finish: line items, each an assertion from the goal and one from what arrived"),
    ("- act's head is a tool's address; return's head is exit(0) or exit(1).",
     "- act's head is a tool's address."),
    ('nact $5 $6 $7"}',
     'nact $5 $6 $7"}\nand, once its output has arrived as $8:\n{"step": "cut $9 := $1.goal / \\"find\\"\\ncut $10 := $8.exit\\ncut $11 := $1.goal / \\"the lines\\"\\ncut $12 := $8.stdout\\nassert $13 := exit(0) from $9\\nassert $14 from $10\\nassert $15 from $11\\nassert $16 from $12\\nreturn ($13 $14) ($15 $16)"}'),
])
LEVEL = int(os.environ.get("LEVEL", "0"))
SYSTEM = BASE
for change in LEVELS[:LEVEL]:
    for old, new in (change if isinstance(change, list) else [change]):
        assert SYSTEM.count(old) == 1, old
        SYSTEM = SYSTEM.replace(old, new)
# Level 9 is written whole: vars, assertions from several cuts, and a final
# line that names assertions from any step of the frame.
SYSTEM9 = r"""You write one step: what this frame does next.

Definitions
  Cut:        a signal of what you think matters.
  Assertion:  your inference from cuts.

Premise
  The more honest you are with your cuts, the better your memory will serve you in the future.

You are shown the variables this frame holds. var $1 is what you received: the goal and the tools. A later var is what arrived from an act. Between them are the lines of your earlier steps.

A step is lines of these kinds:
  cut $n := ADDRESS                  a part of a var
  cut $n := ADDRESS / "words"        a run of words in it, matched by its text
  assert $n := VALUE from $c ...     assert VALUE, from the cuts $c ...
  assert $n from $c ...              assert what the cuts $c ... hold
and it ends with one final line:
  final: act: $h $a ...              run the tool asserted at $h, with the arguments asserted at $a ...
  final: return: $e $a ...           finish, and return the assertions $e $a ... to the frame that asked

ADDRESS: a var's level and a part of it: $k, $k.field, $k[i] (line or item i, from 1), $k[i:j], combined: $3.stdout[1].
VALUE: a literal in single quotes, '-l'; null (nothing: no argument); exit(0) or exit(1); or the ADDRESS of a part of a var: a tool, $1.tools[2], or a part of what arrived, $4.stdout[1].

Rules. A step that breaks one is refused, and you are asked again:
- Assertions can never be cut. Cut only the goal and what arrived.
- Every cut is for exactly one assertion, written after it. An assertion may be from several cuts: together they are its reason, wherever they fall.
- An assertion from several cuts with no VALUE holds their parts: in an act, joined with '/' into one argument; in a return, in order.
- The final names assertions only, at least two, from this step or from an earlier step of this frame. It never cuts.
- An act's first assertion is a tool, and at least one of its assertions is from a cut of the goal.
- Every assertion a return names is from at least one cut of the goal, what it answers, and at least one cut of what arrived, the answer.
- A return's first assertion is exit(0) or exit(1): exit(0) asserts the goal is done; exit(1) asserts it is not.
- Never leave an argument out: where a tool needs none, assert null from the cut that shows it needs none.
- Never cut the same part twice in a step. Number the step's lines from the next free level, in order.
- A literal is remembered as written: assert only what holds for every goal, a flag, never a name, a path, or text you could cut.

Example, for a different goal:
var $1: received
  goal: "find the lines that mention error in this file:\n/var/log/a.log"
  tools: [grep]
Next free level: $2
{"step": "cut $2 := $1.goal / \"find the lines\"\ncut $3 := $1.goal / \"error\"\ncut $4 := $1.goal[2]\nassert $5 := $1.tools[1] from $2\nassert $6 from $3\nassert $7 from $4\nfinal: act: $5 $6 $7"}
and, once what it ran has arrived as var $8:
{"step": "cut $9 := $1.goal / \"find\"\ncut $10 := $8.exit\ncut $11 := $1.goal / \"that mention error\"\ncut $12 := $8.stdout\nassert $13 := exit(0) from $9 $10\nassert $14 from $11 $12\nfinal: return: $13 $14"}

Reply with JSON and nothing else: {"step": "<the lines, separated by \n>"}"""
if LEVEL >= 9:
    SYSTEM = SYSTEM9
READ, ASSERT = "cut", "assert"


class Refused(Exception):
    def __init__(self, code, why):
        super().__init__(f"{code}: {why}")
        self.code, self.why = code, why


def text(v):
    if isinstance(v, str):
        return v
    if isinstance(v, list):
        return "\n".join(map(str, v))
    if isinstance(v, dict):
        return "\n".join(f"{k}: {text(x)}" for k, x in v.items())
    return str(v)


ADDR = re.compile(r'^\$(\d+)((?:\.[A-Za-z_]+|\[\d+(?::\d+)?\])*)$')


def resolve(frame, addr):
    m = ADDR.match(addr)
    if not m:
        raise Refused("instr", f"{addr} is not an address")
    n = int(m.group(1))
    if n not in frame.held:
        raise Refused("unstored", f"${n} is not a value this frame received or that arrived")
    v = frame.held[n]
    for sel in re.findall(r'\.[A-Za-z_]+|\[\d+(?::\d+)?\]', m.group(2)):
        if sel.startswith("."):
            if not isinstance(v, dict) or sel[1:] not in v:
                raise Refused("unread", f"{sel} does not apply to {addr}")
            v = v[sel[1:]]
        else:
            a, _, b = sel[1:-1].partition(":")
            i, j = int(a), int(b or a)
            items = v if isinstance(v, list) else text(v).split("\n")
            if isinstance(v, str) and items and items[-1] == "":
                items = items[:-1]
            if not (1 <= i <= j <= len(items)):
                raise Refused("unread", f"{sel} is outside {addr}")
            part = items[i - 1:j]
            v = part[0] if len(part) == 1 else (part if isinstance(v, list) else "\n".join(part))
    return v


def words_in(s, w):
    return re.search(r'(?<![A-Za-z0-9])' + re.escape(w) + r'(?![A-Za-z0-9])', s) is not None


class Frame:
    def __init__(self, goal, tools):
        self.held = {1: {"goal": goal, "tools": tools}}  # received and arrived, by level
        self.lines = {}  # level -> line written
        self.next = 2
        self.shown = []  # what the model is shown, in level order
        self.cuts, self.asserts = {}, {}  # level 9: every step's, kept for the final

    def render9(self):
        out = []
        for n in sorted(set(self.held) | set(self.lines)):
            if n in self.lines:
                out.append(self.lines[n])
            elif n == 1:
                r = self.held[1]
                out.append(f"var $1: received\n  goal: {json.dumps(r['goal'])}\n  tools: [{', '.join(r['tools'])}]")
            else:
                out.append(f"var ${n}: arrived\n  {json.dumps(self.held[n])}")
        out.append(f"Next free level: ${self.next}")
        return "\n".join(out)

    def check9(self, step):
        lines = [l.strip() for l in step.strip().split("\n") if l.strip()]
        if not lines:
            raise Refused("instr", "the step is empty")
        *body, send = lines
        cuts, asserts = {}, {}
        level = self.next
        owned = {}
        for line in body:
            m = re.match(r'^cut \$(\d+) := (\S+)(?: / "(.*)")?$', line)
            if m:
                n, addr, w = int(m.group(1)), m.group(2), m.group(3)
                if n != level:
                    raise Refused("instr", f"${n} is not the next free level, ${level}")
                a = ADDR.match(addr)
                if not a or int(a.group(1)) not in self.held or (int(a.group(1)) == 1 and not addr.startswith("$1.goal")):
                    raise Refused("cut", f"{addr} is not the goal or what arrived: assertions can never be cut")
                v = text(resolve(self, addr))
                if w is not None and not words_in(v, w):
                    raise Refused("unread", f'"{w}" is not a run of words in {addr}')
                if (addr, w) in [(x[0], x[1]) for x in cuts.values()]:
                    raise Refused("repeated", f"{addr}{' / ' + repr(w) if w else ''} is cut twice")
                cuts[n] = (addr, w, w if w is not None else v, int(a.group(1)) == 1)
                level += 1
                continue
            m = re.match(r'^assert \$(\d+)(?: := (.+?))? from((?: \$\d+)+)$', line)
            if m:
                n, val = int(m.group(1)), m.group(2)
                ps = [int(x[1:]) for x in m.group(3).split()]
                if n != level:
                    raise Refused("instr", f"${n} is not the next free level, ${level}")
                for p in ps:
                    if p not in cuts:
                        raise Refused("untried", f"${n} is from ${p}, which is not a cut written before it in this step")
                    if p in owned:
                        raise Refused("untried", f"${p} is already the reason of ${owned[p]}: every cut is for exactly one assertion")
                    owned[p] = n
                parts = [cuts[p][2] for p in ps]
                asserts[n] = (val, ps, self.value(val, parts[0]) if val is not None else None, parts,
                              any(cuts[p][3] for p in ps), any(not cuts[p][3] for p in ps))
                level += 1
                continue
            raise Refused("instr", f"not a line of a step: {line}")
        unused = [n for n in cuts if n not in owned]
        if unused:
            raise Refused("unused", f"${unused[0]} is the reason of no assertion")
        m = re.match(r'^final: (act|return):((?: \$\d+)+)$', send)
        if not m:
            raise Refused("instr", f"the step does not end in a final line, final: act: ... or final: return: ...: {send}")
        op, names = m.group(1), [int(x[1:]) for x in m.group(2).split()]
        every = {**self.asserts, **asserts}
        for n in names:
            if n not in every:
                raise Refused("instr", f"${n} in the final is not an assertion of this frame")
        if len(names) < 2:
            raise Refused("fragment", "the final names fewer than two assertions")
        if op == "act":
            h = every[names[0]]
            tool = h[0] if h[0] else cuts.get(h[1][0], self.cuts.get(h[1][0]))[0]
            if not (tool or "").startswith("$1.tools"):
                raise Refused("instr", f"${names[0]}, the act's first assertion, is not a tool")
            if not any(every[n][4] for n in names):
                raise Refused("goal", "the act names no assertion from a cut of the goal")
        else:
            for n in names:
                if not (every[n][4] and every[n][5]):
                    raise Refused("item", f"${n} is not from both a cut of the goal and a cut of what arrived")
            if every[names[0]][0] not in ("exit(0)", "exit(1)"):
                raise Refused("instr", f"${names[0]}, the return's first assertion, is not exit(0) or exit(1)")
        for val, _, _, _, _, _ in asserts.values():
            if val and val.startswith("'"):
                lit = val[1:-1]
                for v in self.held.values():
                    if lit and words_in(text(v), lit) and len(lit) > 2:
                        raise Refused("copies", f"'{lit}' is text a value held contains: cut it instead")
        for n in sorted(cuts):
            a, w, _, _ = cuts[n]
            self.lines[n] = f"cut ${n} := {a}" + (f' / "{w}"' if w is not None else "")
        for n in sorted(asserts):
            val, ps, _, _, _, _ = asserts[n]
            self.lines[n] = f"assert ${n}" + (f" := {val}" if val else "") + " from " + " ".join(f"${p}" for p in ps)
        self.lines[level - 0.5] = send
        self.cuts.update(cuts)
        self.asserts.update(asserts)
        self.next = level

        def out(n):
            val, _, v, parts, _, _ = every[n]
            if val is not None:
                return v
            if op == "act":
                return "/".join(str(x).strip("/") if i else str(x).rstrip("/") for i, x in enumerate(parts))
            return parts[0] if len(parts) == 1 else parts
        return op, [out(n) for n in names]

    def render(self):
        out = []
        for n in sorted(set(self.held) | set(self.lines)):
            if n in self.lines:
                out.append(self.lines[n])
            elif n == 1:
                r = self.held[1]
                out.append(f"$1: received\n  goal: {json.dumps(r['goal'])}\n  tools: [{', '.join(r['tools'])}]")
            else:
                out.append(f"${n}: arrived\n  {json.dumps(self.held[n])}")
        out.append(f"Next free level: ${self.next}")
        return "\n".join(out)

    def check(self, step):
        lines = [l.strip() for l in step.strip().split("\n") if l.strip()]
        if not lines:
            raise Refused("instr", "the step is empty")
        *body, send = lines
        elims, intros = {}, {}  # level -> (address, words, read text) / (value, premise, text)
        level = self.next
        for line in body:
            m = re.match(r'^(?:eliminate|read|cut) \$(\d+) := (\S+)(?: / "(.*)")?$', line)
            if m:
                n, addr, w = int(m.group(1)), m.group(2), m.group(3)
                if n != level:
                    raise Refused("instr", f"${n} is not the next free level, ${level}")
                if not (addr.startswith("$1.goal") or (ADDR.match(addr) and int(ADDR.match(addr).group(1)) in self.held and int(ADDR.match(addr).group(1)) != 1)):
                    raise Refused("cut", f"{addr} is not the goal or what arrived; rule 1: assertions can never be cut")
                v = text(resolve(self, addr))
                if w is not None and not words_in(v, w):
                    raise Refused("unread", f'"{w}" is not a run of words in {addr}')
                key = (addr, w)
                if key in [(a, x) for a, x, _ in elims.values()]:
                    raise Refused("repeated", f"{addr}{' / ' + repr(w) if w else ''} is eliminated twice")
                elims[n] = (addr, w, w if w is not None else v)
                level += 1
                continue
            m = re.match(r'^(?:introduce|assert) \$(\d+)(?: := (.+?))? from \$(\d+)$', line)
            if m:
                n, val, p = int(m.group(1)), m.group(2), int(m.group(3))
                if n != level:
                    raise Refused("instr", f"${n} is not the next free level, ${level}")
                if p not in elims:
                    raise Refused("untried", f"${n} is from ${p}, which is not an elimination written before it in this step")
                if p in [q for _, q, _ in intros.values()]:
                    raise Refused("untried", f"${p} is already the premise of another introduction")
                intros[n] = (val, p, self.value(val, elims[p][2]))
                level += 1
                continue
            raise Refused("instr", f"not a line of a step: {line}")
        unused = [n for n in elims if n not in [q for _, q, _ in intros.values()]]
        if unused:
            raise Refused("unused", f"${unused[0]} is the premise of no introduction")
        if LEVEL >= 8 and send.startswith("return"):
            return self.items(send, elims, intros, level)
        m = re.match(r'^(act|return)((?: \S+)+)$', send)
        if not m:
            raise Refused("instr", f"the step does not end in act or return: {send}")
        op, args = m.group(1), m.group(2).split()
        parts = []
        for a in args:
            pieces = re.split(r"'/'", a)
            for pc in pieces:
                if not re.match(r'^\$\d+$', pc) or int(pc[1:]) not in intros:
                    raise Refused("instr", f"{pc} in the sending line is not an introduction of this step")
            parts.append([int(pc[1:]) for pc in pieces])
        if len(parts) < 2:
            raise Refused("fragment", "the sending line sends fewer than two introductions: a head and at least one argument")
        need = 1 if LEVEL >= 8 else (2 if op == "act" else 1)
        if sum(e[0].startswith("$1.goal") for e in elims.values()) < need:
            raise Refused("goal", f"the goal is cut fewer than {need} times; rule 2")
        if not any(elims[intros[i][1]][0].startswith("$1.goal") for grp in parts for i in grp):
            raise Refused("goal", f"this {op} sends no assertion derived from a cut of the goal; rule 4")
        head = intros[parts[0][0]]
        tool = head[0] if head[0] else elims[head[1]][0]
        if op == "act" and not tool.startswith("$1.tools"):
            raise Refused("instr", "act's head is not introduced from a tool's address")
        if op == "return" and head[0] not in ("exit(0)", "exit(1)"):
            raise Refused("instr", "return's head is not exit(0) or exit(1)")
        for val, _, _ in intros.values():
            if val and val.startswith("'"):
                lit = val[1:-1]
                for v in self.held.values():
                    if lit and words_in(text(v), lit) and len(lit) > 2:
                        raise Refused("copies", f"'{lit}' is text a value held contains: {READ} it instead")
        for n in sorted(elims):
            a, w, _ = elims[n]
            self.lines[n] = f"{READ} ${n} := {a}" + (f' / "{w}"' if w is not None else "")
        for n in sorted(intros):
            val, p, _ = intros[n]
            self.lines[n] = f"{ASSERT} ${n}" + (f" := {val}" if val else "") + f" from ${p}"
        self.next = level
        return op, [[intros[i][2] for i in grp] for grp in parts]

    def items(self, send, elims, intros, level):
        groups = re.findall(r'\(([^)]*)\)', send)
        if not re.match(r'^return( \([^)]*\))+$', send) or not groups:
            raise Refused("instr", "a return is line items: return ($g $a) ($g $a) ...")
        items = []
        for g in groups:
            refs = g.split()
            if len(refs) != 2 or not all(re.match(r'^\$\d+$', r) and int(r[1:]) in intros for r in refs):
                raise Refused("fragment", f"({g}) is not a line item: two assertions of this step")
            a, b = int(refs[0][1:]), int(refs[1][1:])
            if not elims[intros[a][1]][0].startswith("$1.goal"):
                raise Refused("goal", f"${a}, the first of ({g}), is not from a cut of the goal")
            src = ADDR.match(elims[intros[b][1]][0])
            if not src or int(src.group(1)) == 1:
                raise Refused("child", f"${b}, the second of ({g}), is not from a cut of what arrived")
            items.append((a, b))
        if intros[items[0][0]][0] not in ("exit(0)", "exit(1)"):
            raise Refused("instr", "the first line item's first assertion is not exit(0) or exit(1)")
        sent = {i for it in items for i in it}
        for n in sorted(elims):
            a, w, _ = elims[n]
            self.lines[n] = f"{READ} ${n} := {a}" + (f' / "{w}"' if w is not None else "")
        for n in sorted(intros):
            val, p, _ = intros[n]
            self.lines[n] = f"{ASSERT} ${n}" + (f" := {val}" if val else "") + f" from ${p}"
        self.next = level
        return "items", [(intros[a][2], intros[b][2]) for a, b in items]

    def value(self, val, read):
        if val is None:
            return read
        if val == "null":
            return None
        if val in ("exit(0)", "exit(1)"):
            return int(val[5])
        if val.startswith("'") and val.endswith("'"):
            return val[1:-1]
        return text(resolve(self, val))


def ask(system, prompt):
    env = {k: v for k, v in os.environ.items() if k not in ("ANTHROPIC_API_KEY", "ANTHROPIC_AUTH_TOKEN")}
    args = ["claude", "-p", "--output-format", "json", "--model", "claude-opus-5-5", "--tools", "",
            "--disable-slash-commands", "--strict-mcp-config", "--mcp-config", '{"mcpServers":{}}',
            "--setting-sources", "user", "--settings", '{"disableAllHooks":true}',
            "--system-prompt", system, "--effort", "high", "--", prompt]
    p = subprocess.run(args, capture_output=True, text=True, env=env, stdin=subprocess.DEVNULL)
    if p.returncode != 0:
        raise SystemExit(f"claude CLI exit {p.returncode}: {p.stderr.strip()} {p.stdout.strip()[:500]}")
    out = p.stdout
    res = json.loads(out)["result"].strip()
    i, j = res.find("{"), res.rfind("}")
    return json.loads(res[i:j + 1])["step"]


def run(goal, tools, cwd):
    f = Frame(goal, tools)
    print(f"--- prompt: level {LEVEL}, sha256 {hashlib.sha256(SYSTEM.encode()).hexdigest()[:12]}")
    for _ in range(6):
        prompt = f.render9() if LEVEL >= 9 else f.render()
        for attempt in range(3):
            step = ask(SYSTEM, prompt)
            print(f"--- ask (level ${f.next}), attempt {attempt + 1}\n{step}")
            try:
                op, parts = f.check9(step) if LEVEL >= 9 else f.check(step)
                break
            except Refused as r:
                print(f"REFUSED {r.code}: {r.why}")
                prompt = (f.render9() if LEVEL >= 9 else f.render()) + f"\n\nYour step was refused: {r.code}: {r.why}\nThe step:\n{step}"
        else:
            print("gave up")
            return
        if LEVEL >= 9:
            if op == "return":
                print(f"=== return {parts!r}")
                return
            argv = [str(v) for v in parts if v is not None]
            if any("/" in v and not v.startswith(os.path.dirname(os.path.abspath(__file__))) for v in argv[1:]):
                print(f"BLOCKED {argv}: a path outside the test directory")
                return
            p = subprocess.run(argv, capture_output=True, text=True, cwd=cwd)
            arrival = {"exit": p.returncode, "stdout": p.stdout, "stderr": p.stderr}
            print(f"=== ran {argv} -> {json.dumps(arrival)}")
            f.held[f.next] = arrival
            f.next += 1
            continue
        if op == "items":
            print("=== return " + "; ".join(f"{g!r}: {v!r}" for g, v in parts))
            return
        vals = ["/".join(str(x) for x in grp) if all(x is not None for x in grp) else None for grp in parts]
        if op == "return":
            print(f"=== return exit {vals[0]}: {vals[1:]}")
            return
        argv = [v for v in vals if v is not None]
        if any("/" in v and not v.startswith(os.path.dirname(os.path.abspath(__file__))) for v in argv[1:]):
            print(f"BLOCKED {argv}: a path outside the test directory")
            return
        p = subprocess.run(argv, capture_output=True, text=True, cwd=cwd)
        arrival = {"exit": p.returncode, "stdout": p.stdout, "stderr": p.stderr}
        print(f"=== ran {argv} -> {json.dumps(arrival)}")
        f.held[f.next] = arrival
        f.next += 1


if __name__ == "__main__":
    if sys.argv[1] == "--prompt":
        sys.exit(print(SYSTEM))
    goal, tools = json.load(open(sys.argv[1]))
    run(goal, tools, os.path.dirname(os.path.abspath(__file__)))
