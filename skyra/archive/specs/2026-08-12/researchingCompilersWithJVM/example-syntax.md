# Example syntax

**Date:** 2026-08-17
**Status: MATERIAL.** The worked example this directory reasons over. No conclusions.

Two compilation units, two packages, one cross-package static call.

---

## `packagex/Main.java`

```java
package packagex;

import packagey.SomeClass;

public class Main {
    public static void main(String[] args) {
        String x = SomeClass.someMethod();
    }
}
```

## `packagey/SomeClass.java`

```java
package packagey;

public class SomeClass {
    public static String someMethod() {
        return "hello";
    }
}
```
