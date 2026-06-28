# Development Workflow

## Purpose

This document defines the development workflow for the Open Workforce Platform.

The workflow exists to keep development aligned with the architecture, engineering principles and long-term goals of the project.

---

## Core Principle

Every implementation should trace back to a decision, issue or documented objective.

Code should not appear without context.

---

## Workflow

## 1. Define the Work

Every meaningful change should begin with a clear issue or objective.

The issue should explain:

- What problem is being solved.
- Why it matters.
- What success looks like.

## 2. Check the Architecture

Before implementation begins, confirm where the change belongs.

The change should respect:

- The domain model.
- Bounded contexts.
- Architecture decisions.
- Engineering principles.

## 3. Implement Simply

Start with the smallest implementation that proves the behaviour.

Avoid unnecessary abstractions.

Avoid adding dependencies unless they clearly earn their place.

## 4. Test Behaviour

Add tests that validate the expected behaviour.

Tests should focus on business outcomes rather than implementation details.

## 5. Update Documentation

If a change affects architecture, domain language, behaviour or usage, documentation should be updated.

Documentation is part of the product.

## 6. Review

Review should challenge:

- Does this add value?
- Does this preserve the domain model?
- Does this respect architectural boundaries?
- Is it understandable?
- Is it tested?
- Is it documented where needed?

## 7. Merge

Changes should be merged when they are clear, tested, documented and aligned with the project principles.

---

## AI-Assisted Workflow

AI may assist with research, design, implementation, testing and review.

AI-generated suggestions should be treated as proposals, not decisions.

Human judgement remains responsible for final engineering decisions.

---

## Key Principle

The workflow should increase confidence, not bureaucracy.
