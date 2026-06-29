# Time Windows

## Purpose

Introduce time window constraints for Work Items and Resources.

This allows the optimiser to schedule work within available operating hours rather than simply consuming capacity.

---

# Why

Duration-based capacity measures how long work takes.

Time windows determine when work may be performed.

Together they represent the foundation of workforce scheduling.

---

# Behaviour

Resources expose:

- shift start
- shift end

Work Items expose:

- earliest start
- latest finish

All values are expressed as minutes from midnight.

Examples:

Resource

{
    "shiftStart": 480,
    "shiftEnd": 1020
}

08:00–17:00

Work Item

{
    "duration": 90,
    "earliestStart": 540,
    "latestFinish": 900
}

09:00–15:00

---

# Scheduling

The optimiser should determine whether a Work Item can fit within the remaining schedule for a Resource.

Initially:

Use a simple sequential schedule.

Assignments occur one after another.

No travel time.

No gaps.

---

# Constraints

Existing constraints remain unchanged:

- availability
- skills
- duration
- priority

Time window validation is added.

---

# Objective

Objective scoring remains unchanged.

---

# Dataset

Update the sample dataset to include realistic shifts and work windows.

---

# Non-Goals

Do not implement:

- travel time
- route optimisation
- breaks
- lunch
- overtime
- multiple shifts
- calendars

---

# Tests

Add tests covering:

- work fits exactly
- work exceeds shift
- work outside window
- multiple jobs scheduled sequentially
- existing constraints still respected

---

# Definition of Done

- resources expose shifts
- work exposes time windows
- scheduling respects windows
- algorithms continue to work
- tests pass

---

# Open Questions

None.