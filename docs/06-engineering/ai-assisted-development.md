# AI-Assisted Development

## Purpose

This document describes how AI is used during the development of the Open Workforce Platform.

AI is treated as an engineering collaborator rather than a code generation tool.

The objective is to combine human architectural judgement with AI implementation capabilities.

---

# Philosophy

The project follows a simple principle:

> Humans own architecture.
>
> AI owns implementation.
>
> Humans own the final decision.

Neither the human nor the AI is expected to solve every problem alone.

The best outcome comes from collaboration.

---

# Responsibilities

## Human Responsibilities

The human engineers are responsible for:

- defining the product vision
- understanding the business domain
- designing the architecture
- defining engineering principles
- maintaining architectural integrity
- reviewing all implementation
- making final engineering decisions

---

## AI Responsibilities

The AI is responsible for:

- analysing specifications
- proposing implementation approaches
- identifying ambiguities
- implementing features
- producing unit tests
- suggesting documentation updates
- explaining engineering decisions

AI is encouraged to make engineering decisions where appropriate rather than simply translating specifications into code.

---

# Engineering Workflow

Every feature follows the same workflow.

```text
Architecture
        │
        ▼
Engineering Principles
        │
        ▼
AI Steering
        │
        ▼
Specification
        │
        ▼
AI Implementation
        │
        ▼
Human Review
        │
        ▼
Merge
```

---

# Human Review

The review focuses on engineering quality rather than syntax.

Questions include:

- Does this respect the architecture?
- Does it strengthen the domain model?
- Has everything earned its place?
- Is the implementation unnecessarily complex?
- Are the tests validating behaviour?
- Would this be approved during a professional code review?

---

# Engineering Principles

AI should not optimise for producing the most code.

AI should optimise for producing the best engineering outcome.

Where multiple solutions exist, AI should explain trade-offs and recommend the preferred solution.

---

# Lessons Learned

The first implementation of `BusinessEvent` demonstrated several important observations.

Well-written steering documents produced significantly better implementation quality than detailed coding instructions.

Providing architecture, engineering principles and specifications allowed AI to make sensible engineering decisions while remaining within architectural boundaries.

The most valuable human contribution was architectural review rather than code generation.

The most valuable AI contribution was implementation and reasoning within clearly defined constraints.

---

# Continuous Improvement

This workflow is expected to evolve throughout the project.

Whenever a lesson is learned that improves collaboration between human engineers and AI, this document should be updated.