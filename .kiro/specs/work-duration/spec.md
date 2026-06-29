# Work Duration

## Purpose

Replace unit-capacity planning with duration-based planning.

Resources expose available capacity in minutes.

Work Items expose required duration in minutes.

The optimiser should consume capacity based on duration rather than counting Work Items.

---

# Why

Treating every Work Item as equal is unrealistic.

Real workforce planning allocates time rather than job count.

This change moves the platform closer to real optimisation problems while preserving the existing architecture.

---

# Behaviour

Resources provide:

{
    "capacity": 480
}

representing available minutes.

Work Items provide:

{
    "duration": 90
}

representing required minutes.

The optimiser should only assign a Work Item when sufficient remaining minutes exist.

Remaining capacity decreases by the assigned duration.

---

# Constraints

Capacity
Availability
Skills
Priority

continue to work exactly as today.

Only the capacity calculation changes.

---

# Objective

Objective scoring remains unchanged.

The assignment objective still dominates.

Workload balance should now compare remaining minutes rather than assignment counts where appropriate.

---

# Dataset

Update the example dataset with realistic durations.

Example:

Patient Assessment: 90 minutes

Maintenance Visit: 120 minutes

Delivery: 45 minutes

---

# Non-Goals

Do not implement:

- calendars
- travel time
- start/end times
- breaks
- overtime

---

# Tests

Add tests covering:

- exact capacity fit
- insufficient remaining minutes
- multiple smaller jobs filling capacity
- duration-aware workload balancing
- existing constraints still respected

---

# Definition of Done

- capacity is measured in minutes
- Work Items consume duration
- all algorithms continue to work
- CLI continues to run
- tests pass

---

# Open Questions

None.