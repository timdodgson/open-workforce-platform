# Dependency Policy

## Purpose

This document defines how external dependencies are selected, maintained and reviewed within the Open Workforce Platform.

Dependencies can provide significant value, but they also introduce maintenance, security and operational risk.

The purpose of this policy is to ensure that every dependency earns its place.

---

## Core Principle

Every dependency is a long-term maintenance commitment.

A package should only be added when it provides clear value that outweighs its cost.

---

## When to Use a Dependency

A dependency may be appropriate when it:

- Provides complex functionality that would be expensive or risky to build internally.
- Is a well-maintained ecosystem standard.
- Reduces security risk by using a mature implementation.
- Improves reliability, correctness or maintainability.
- Saves significant engineering effort without introducing unreasonable complexity.

---

## When to Avoid a Dependency

A dependency should be avoided when it:

- Replaces simple, understandable code.
- Provides only trivial helper functions.
- Introduces a large transitive dependency tree.
- Is poorly maintained or abandoned.
- Makes the system harder to understand.
- Adds complexity without clear value.

---

## Dependency Review

Before adding a dependency, consider:

- What problem does it solve?
- Could the same result be achieved clearly with standard library code?
- Is the package actively maintained?
- How many transitive dependencies does it introduce?
- What is the security and patching history?
- How difficult would it be to remove later?

---

## Patching and Maintenance

Dependencies must be reviewed and updated regularly.

Security fixes should be prioritised.

Outdated dependencies should not be ignored simply because the system still works.

If a dependency no longer earns its place, it should be removed.

---

## Key Principle

Dependencies must earn their place.

Convenience is not enough.
