# Decision Making Steering

## Purpose

This document defines how engineering decisions should be made when multiple valid solutions exist.

The objective is not to generate the most code.

The objective is to produce the highest quality engineering outcome.

When uncertain, prefer the solution that best aligns with the project's Engineering Principles.

---

# Decision Order

When evaluating multiple solutions, consider them in the following order:

1. Does it add value?
2. Does it preserve the domain model?
3. Is it architecturally correct?
4. Is it simple to understand?
5. Does it minimise long-term maintenance?
6. Does it avoid unnecessary dependencies?
7. Does it remain easy to test?
8. Does it improve the overall system?

Never optimise for fewer lines of code alone.

---

# Dependencies

Every dependency must earn its place.

Before proposing a dependency, explain:

- What problem it solves.
- Why the standard library is insufficient.
- Why the long-term benefits outweigh the maintenance cost.
- Why introducing the dependency is preferable to writing a small amount of straightforward code.

Avoid dependencies that provide only convenience.

---

# Abstractions

Every abstraction must earn its place.

Before introducing:

- interfaces
- factories
- builders
- generic frameworks
- helper packages

consider whether the additional complexity is justified.

Prefer simple, explicit implementations until abstraction clearly improves the design.

---

# Functions

Prefer small functions with a single responsibility.

Avoid functions with more than four parameters.

If more than four parameters are required, reconsider the design.

Consider introducing:

- a domain object
- a value object
- a builder
- a smaller function

Large parameter lists usually indicate the wrong abstraction.

---

# Packages

Do not create packages simply to separate files.

Packages should represent meaningful architectural boundaries.

Avoid generic packages such as:

- utils
- helpers
- common
- misc

Every package should have a clearly defined responsibility.

---

# AI Behaviour

Do not optimise for producing more code.

Do not introduce patterns simply because they are common.

If multiple solutions are possible:

- explain the trade-offs
- recommend the preferred solution
- explain why

Challenge architectural decisions that appear inconsistent rather than silently working around them.

---

# Final Check

Before completing a task, ask:

- Is this the simplest solution?
- Does it strengthen the architecture?
- Does everything introduced earn its place?
- Would I recommend this solution during a senior engineering design review?

If not, reconsider the implementation before completing the task.