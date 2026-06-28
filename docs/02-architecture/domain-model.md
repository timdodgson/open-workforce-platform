# Domain Model

## Purpose

This document defines the core business concepts used throughout the Open Workforce Platform.

The objective is to create a common language that is independent of any programming language, framework, or optimisation engine.

Every component of the platform should use the terminology defined in this document.

---

# Core Domain Concepts

## Business Event

A Business Event represents something that creates a requirement for work.

Business Events are the origin of Work Items within the platform.

Examples include:

- Customer order
- Equipment failure
- Maintenance due
- Patient referral
- Emergency incident
- Planned inspection

A Business Event may generate one or more Work Items.

## Resource

A Resource represents anything capable of completing one or more Work Items.

Resources are assigned Work Items by the optimisation engine.

A Resource may represent:

- Human Worker
- Vehicle
- Robot
- AI Agent
- External Contractor
- Third-party Service

Different Resource types expose different capabilities, constraints and availability.

## Work Item

A Work Item is created as the result of a Business Event.

A Work Item represents a piece of work that must be completed.

A Work Item may be assigned to one or more Resources during its lifecycle.

It is the fundamental concept within the Open Workforce Platform.

Every optimisation performed by the platform ultimately exists to maximise the effective completion of Work Items.

A Work Item may have characteristics such as:

- Required skills
- Expected duration
- Priority
- Time window
- Location
- Dependencies
- Business constraints

A Work Item is intentionally independent of any specific industry.

Examples include:

- Visit a patient
- Repair a gas leak
- Install broadband
- Deliver a parcel
- Inspect an aircraft
- Perform a safety check