# Next: capability

Where the controller goes after the prototype. Agreed on paper, not built.

1. **[`14-interpreted-not-stated.md`](14-interpreted-not-stated.md)**: the
   model never emits a control operator. It supplies only operands (which
   capability, and a payload), and the runtime derives invoke, return, fault or
   interrupt. A return is invoking a single-use `reply` capability.
2. **[`research-question.md`](research-question.md)**: the question, the
   metrics, the baseline, the September–December plan, and the open questions
   to decide together.
3. **[`agenda.md`](agenda.md)**: the immediate work. Define input contracts,
   write `interpret` as a total pure function, define `reply`'s lifecycle, and
   fill in the checklist for every transition.
