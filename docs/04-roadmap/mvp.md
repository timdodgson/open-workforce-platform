# MVP Definition

## Purpose

This document defines the first working version of the Open Workforce Platform.

The MVP exists to prove the core platform concept with the smallest useful implementation.

It should validate the domain model, bounded contexts and system architecture without attempting to deliver the full long-term roadmap.

---

# MVP Goal

The MVP should demonstrate that the platform can:

1. Accept business demand.
2. Represent that demand as Work Items.
3. Represent available Resources.
4. Apply Constraints.
5. Apply Objectives.
6. Produce an optimised plan.

---

# Non-Goals

The MVP will not include:

- Full web application
- Multi-tenancy
- Authentication
- AI assistant
- Cloud deployment
- External system integrations
- Real-time execution tracking
- Mobile application

---

# MVP Approach

The MVP will be command-line only.

This keeps the first implementation focused on proving the core optimisation workflow without introducing user interface, API, authentication or deployment concerns.

The command-line approach is intentional. It allows the project to validate the domain model and optimisation logic before adding platform layers.

---

# MVP v0.1

## Input

- Business Events
- Work Items
- Resources
- Constraints
- Objectives

## Processing

- Load input data
- Create Work Items
- Match Work Items to Resources
- Apply Constraints
- Optimise against Objectives

## Output

- Optimised Plan
- Constraint results
- Objective score
- Human-readable explanation

---

# Initial Dataset

The MVP will use a small, hand-crafted dataset.

The purpose of this dataset is to validate the optimisation workflow and domain model rather than benchmark algorithm performance.

The platform must not be designed around a single dataset or industry.

Different organisations will naturally have different data models, business events, resources and constraints.

The initial dataset exists purely as a simple, understandable example.

Larger datasets, including the original dissertation dataset and industry benchmarks, will be introduced later for validation and performance testing.

---

# Definition of Success

The MVP is considered successful when, given a valid dataset, it can:

- Produce a valid optimised plan.
- Satisfy all hard constraints.
- Optimise against the configured objectives.
- Explain why the proposed plan was selected.
- Produce results that satisfy a reasonable "common sense" review by a domain expert.

The platform should not only produce mathematically valid solutions but also solutions that are understandable and practical in real-world operational environments.