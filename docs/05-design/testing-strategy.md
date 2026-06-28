# Testing Strategy

## Purpose

This document defines the testing strategy for the Open Workforce Platform.

Testing exists to provide confidence that the platform behaves correctly, remains maintainable and can evolve safely.

The goal is not to maximise test count or coverage percentage.

The goal is to validate important behaviour.

---

## Core Principle

Test behaviour, not implementation.

Tests should describe what the platform should do in business terms.

Implementation details may change.

Expected behaviour should remain protected.

---

## Test Types

## Unit Tests

Unit tests validate small pieces of domain behaviour in isolation.

They should be fast, focused and easy to understand.

Examples include:

- A Business Event creates Work Items.
- A Work Item requires specific skills.
- A Constraint identifies an invalid assignment.
- An Objective calculates a score.

## Integration Tests

Integration tests validate collaboration between parts of the system.

Examples include:

- Loading a dataset.
- Applying constraints.
- Producing an optimised plan.
- Returning an explanation for a solution.

## Scenario Tests

Scenario tests validate full business scenarios.

They should be written in a way that a domain expert can understand.

Examples include:

- Assign engineers to jobs without violating skill constraints.
- Prefer lower travel distance when all hard constraints are satisfied.
- Avoid overtime where possible.
- Produce a plan that is defensible to an operations manager.

## Regression Tests

Every defect should result in a regression test.

The test should fail before the fix and pass after the fix.

The purpose is to prevent the same problem returning.

---

## What Not to Test

Tests should avoid depending on:

- Private implementation details.
- Internal helper functions unless they contain meaningful domain behaviour.
- Exact algorithm internals where the business outcome is what matters.
- Fragile formatting unless the format is part of the contract.

---

## Quality Bar

A good test should be:

- Easy to read.
- Fast enough to run frequently.
- Focused on one behaviour.
- Useful when it fails.
- Written in the language of the domain.

---

## Key Principle

Tests are not proof that software is perfect.

Tests are evidence that important behaviour is protected.
