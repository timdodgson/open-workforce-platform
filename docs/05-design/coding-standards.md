# Coding Standards

## Purpose

This document defines the coding standards for the Open Workforce Platform.

These standards focus on maintainability, clarity and domain alignment rather than formatting rules.

Formatting tools may enforce style.

Engineering judgement must enforce quality.

---

## Core Principle

Code should make the domain easier to understand.

If the code obscures the business meaning, it is working against the architecture.

---

## Naming

Names should use the language of the domain.

Prefer names such as:

- BusinessEvent
- WorkItem
- Resource
- Constraint
- Objective
- OptimisedPlan

Avoid technical names that hide business meaning.

---

## Simplicity

Prefer simple, explicit code.

Avoid unnecessary cleverness.

A future engineer should be able to understand the intent of the code without needing to reverse engineer it.

---

## Boundaries

Code should respect bounded contexts.

Business rules should not leak into unrelated modules.

The optimisation engine should not own business knowledge.

The CLI should orchestrate workflows without becoming a place for hidden business logic.

---

## Dependencies

Use the standard library where it is sufficient.

External packages should only be introduced when they provide clear value.

A dependency should never be added to avoid writing a few lines of simple, understandable code.

---

## Comments

Comments should explain intent, trade-offs or non-obvious decisions.

Comments should not repeat what the code already says.

Where possible, improve the code before adding comments.

---

## Error Handling

Errors should be clear and actionable.

Domain errors should be expressed in domain language.

A user or developer should be able to understand what went wrong and what can be done next.

---

## Key Principle

Readable code is not basic code.

Readable code is professional code.
