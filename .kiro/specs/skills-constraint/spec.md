# Skills Constraint

## Purpose

Introduce skill-based assignment into the optimiser.

The optimiser should only assign Work Items to Resources that have the required skill.

This is the platform's third optimisation constraint after capacity and availability.

---

# Why

Capacity determines how much work a Resource can take.

Availability determines whether a Resource can take work at all.

Skills determine whether a Resource is capable of performing a specific Work Item.

This moves the optimiser closer to real workforce planning.

---

# Problem Statement

Given:

- a collection of Work Items
- a collection of Resources
- resource capacity
- resource availability
- resource skills
- work item required skills

The optimiser should assign Work Items only to Resources that:

- are available
- have remaining capacity
- have the required skill

---

# Skills

Skills remain business information.

The Resource and WorkItem domain objects remain generic.

Resource skills are supplied within Resource details JSON.

Example:

{
  "capacity": 2,
  "available": true,
  "skills": ["clinical", "assessment"]
}

Required skill is supplied within WorkItem details JSON.

Example:

{
  "priority": 100,
  "requiredSkill": "clinical"
}

The application layer is responsible for interpreting skills.

The optimisation layer receives skills as structured optimisation input.

If a WorkItem has no required skill, it may be assigned to any available Resource with capacity.

If a Resource has no skills, treat it as having no skills.

Skill matching is exact and case-sensitive for now.

---

# Behaviour

The optimiser should:

- continue to respect capacity
- continue to respect availability
- continue to prioritise higher-priority Work Items
- assign Work Items only to Resources with matching skills
- remain deterministic

If no available Resource has the required skill:

- the Work Item should remain unassigned
- the score should reflect the number of assigned Work Items

---

# Non-Goals

Do not implement:

- multiple required skills
- skill levels
- skill synonyms
- skill hierarchies
- certifications
- calendars
- routing
- weighted objectives
- generic constraint framework

---

# Optimised Plan

The Optimised Plan should continue to report:

- assignments
- assigned work item count
- unassigned work item count
- total resource capacity
- utilisation percentage
- optimisation score

Work Items that cannot be matched to a skilled Resource should remain unassigned.

---

# Tests

Add tests covering:

- Work Item assigned when Resource has required skill
- Work Item unassigned when no Resource has required skill
- Work Item with no required skill can be assigned to any available Resource with capacity
- Resource with no skills cannot satisfy a required skill
- capacity is still respected
- availability is still respected
- priority ordering is still respected
- optimiser remains deterministic

---

# Constraints

Use only the Go standard library.

Do not introduce external dependencies.

Do not introduce interfaces or a generic constraint framework.

Do not add skill fields to Resource or WorkItem domain objects.

Keep the implementation intentionally simple.

---

# Definition of Done

The implementation is complete when:

- Work Items are only assigned to Resources with matching skills where a required skill exists
- Work Items without required skills can still be assigned
- capacity constraints are still respected
- availability constraints are still respected
- priority ordering is still respected
- all tests pass
- the command-line example still runs
- no external dependencies are introduced
- architecture remains consistent with steering documents

---

# Open Questions

None.