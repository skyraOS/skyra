# Parse → AST

**Date:** 2026-08-17
**Status: EVIDENCE.** Measured with `javac 23.0.2` on macOS 24.6.0, against
[`example-syntax.md`](example-syntax.md). Commands are in §5 so every claim can be re-run.

**No conclusions.** This records what the first phase of `javac` produces and what it refuses to
know. Where it bears on the design, §6 says where it attaches and stops there.

> **Parsing is per-compilation-unit and purely syntactic. It builds a tree of SHAPES. It does not
> know what a single name refers to, and it cannot, because the input that would tell it is not
> supplied to this phase.**

---

## 1 · Where the phase sits

```text
1  PARSE        source → AST                        ← THIS DOCUMENT
2  ENTER        populate the symbol table
3  ANNOTATION PROCESSING
4  ATTRIBUTE    resolve names · check types
5  FLOW         definite assignment · reachability
6  DESUGAR      erasure · inner classes · lambdas
7  GENERATE     bytecode
```

Only phase 1 runs below. `com.sun.source.util.JavacTask.parse()` returns after it, before any
symbol table exists.

---

## 2 · The measured trees ✅

```text
=== packagex/Main.java ===
COMPILATION_UNIT       package packagex; import packagey.SomeClass; public …
  PACKAGE                package packagex;
    IDENTIFIER             packagex
  IMPORT                 import packagey.SomeClass;
    MEMBER_SELECT          packagey.SomeClass
      IDENTIFIER             packagey
  CLASS                   public class Main { public static void main(String[…
    MODIFIERS              public
    METHOD                  public static void main(String[] args) { String x =…
      MODIFIERS              public static
      PRIMITIVE_TYPE         void
      VARIABLE               String[] args
        MODIFIERS
        ARRAY_TYPE             String[]
          IDENTIFIER             String
      BLOCK                  { String x = SomeClass.someMethod(); }
        VARIABLE               String x = SomeClass.someMethod()
          MODIFIERS
          IDENTIFIER             String
          METHOD_INVOCATION      SomeClass.someMethod()
            MEMBER_SELECT          SomeClass.someMethod
              IDENTIFIER             SomeClass

=== packagey/SomeClass.java ===
COMPILATION_UNIT       package packagey; public class SomeClass { public st…
  PACKAGE                package packagey;
    IDENTIFIER             packagey
  CLASS                   public class SomeClass { public static String someM…
    MODIFIERS              public
    METHOD                  public static String someMethod() { return "hello";…
      MODIFIERS              public static
      IDENTIFIER             String
      BLOCK                  { return "hello"; }
        RETURN                 return "hello";
          STRING_LITERAL         "hello"
```

Two trees, not one. **There is no edge between them.** Nothing in `Main`'s tree points at
`SomeClass`'s tree — the connection exists only as the *text* `packagey.SomeClass`, unresolved.

Note also what the declared type is: `String x` parses to a bare `IDENTIFIER`. Not a type, not a
reference — a name that has not been looked up.

---

## 3 · Finding — the parse resolves nothing, and this is measurable ✅

`Main.java` was copied into a directory where `packagey` genuinely does not exist, with an empty
`-sourcepath` and `-classpath`. Same file, same conditions, two phases:

```text
PARSE      SUCCEEDS. full tree, no diagnostics. IMPORT and METHOD_INVOCATION both present.

COMPILE    FAILS.
           packagex/Main.java:3: error: package packagey does not exist
           packagex/Main.java:7: error: cannot find symbol
                   String x = SomeClass.someMethod();
                              ^
             symbol:   variable SomeClass
             location: class Main
```

**Parsing succeeded on a file that cannot compile.** The difference between the two is exactly
resolution, and resolution takes an input — the symbol table, built from sourcepath and classpath —
that the parse phase is never given.

```text
⚠️  A NEAR-MISS WORTH RECORDING. The first attempt at this measurement APPEARED to show compilation
    succeeding. It did not: javac's default sourcepath is the user class path, which defaults to `.`,
    so it found packagey/SomeClass.java sitting one directory up and compiled it implicitly. The
    isolation had to be made real — a separate tree plus empty -sourcepath and -classpath — before
    the measurement meant anything.
```

---

## 4 · Finding — `MEMBER_SELECT` is a shape, and the error message admits it ✅

The sharpest line in the output is the compiler's own guess:

```text
symbol:   variable SomeClass
```

`javac` calls `SomeClass` a **variable**. It is a class. The parser produced

```text
METHOD_INVOCATION
  MEMBER_SELECT          SomeClass.someMethod
    IDENTIFIER             SomeClass
```

and that shape is identical for every one of these:

```text
SomeClass.someMethod()      a static method on a class
someObject.someMethod()     an instance method on a local variable
packagey.SomeClass          a class inside a package        ← the IMPORT above, SAME NODE TYPE
a.b.c.d.e()                 any mixture of the three
```

`a.b.c` cannot be disambiguated syntactically into *package.Class.member*, *Class.field.member*, or
*variable.field.member*. Java resolves it afterward with ordering rules; the tree cannot express the
distinction because **the parser has no basis on which to make it.**

The `IMPORT` node is the same `MEMBER_SELECT` shape as the call site. One is a scope declaration and
one is an invocation, and at parse time they are the same kind of object.

---

## 5 · How it was measured

```java
// DumpAst.java — public API only: com.sun.source.tree, com.sun.source.util, javax.tools.
JavacTask task = (JavacTask) compiler.getTask(null, fm, null,
        List.of("-proc:none"), null, fm.getJavaFileObjects(args));
for (CompilationUnitTree t : task.parse()) {      // ← RETURNS AFTER PARSE. no Enter, no Attribute.
    new TreeScanner<Void,Integer>() {
        public Void scan(Tree n, Integer d) { /* print n.getKind() + n.toString() */ }
    }.scan(t, 0);
}
```

```console
$ javac DumpAst.java
$ java DumpAst packagex/Main.java packagey/SomeClass.java     # §2

$ mkdir -p iso/packagex && cp packagex/Main.java iso/packagex/ && cd iso
$ find . -type f                                              # only Main.java
$ java -cp .. DumpAst packagex/Main.java                      # parses
$ javac -sourcepath "" -classpath "" -d /tmp/out packagex/Main.java   # fails    # §3
```

---

## 6 · Where this attaches, and nothing more

Recorded as pointers, not conclusions. Each is a place the measurement bears on an existing
question; none is answered here.

```text
THE PARSE/RESOLVE SPLIT     ../semantics/terminology.md's SCOPE — "what may be NAMED… you can
                            select what you can name, you can name what you were given" — and
                            frame.Resolve. javac separates producing a tree from knowing what its
                            names mean, and supplies the second from a source the first never sees.

THE IMPORT                  is a per-compilation-unit statement of which names are resolvable.
                            It is not a dependency and not an edge; nothing outside it is nameable
                            without full qualification.

`a.b.c` IS AMBIGUOUS        and Java needs ordering rules to settle it. Any surface where a
                            possibility names things will meet the same ambiguity or must forbid
                            the shape that creates it.

WHAT A COMPILER SLOT DOES   ../structs and interfaces/aug-14-int-structs/00-principles.md P19:
                            "the causal syntax lives at an address, and all it is is a parser that
                            calls the kernel's methods." This measures what the first half of that
                            sentence actually produces — and shows it produces something that
                            cannot call anything yet, because it does not know what the names are.
```

```text
NOT MEASURED HERE   Enter · Attribute · Flow · Desugar · Generate. The symbol table, the resolution
                    rules, bytecode, and the JIT are all later phases and separate documents.
```
