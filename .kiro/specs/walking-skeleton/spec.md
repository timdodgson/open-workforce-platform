# Walking Skeleton

## Purpose

Create the first runnable vertical slice of the Open Workforce Platform.

The goal is not to produce a real optimisation algorithm yet.

The goal is to prove that the platform can load input, create Work Items, run an optimiser, produce an Optimised Plan, and display the result from the command line.

## Target Command

From platform/go:

go run ./cmd/owp optimise ../../examples/datasets/simple-events.json

## Expected Behaviour

The command should:

1. Load a JSON dataset containing Business Events.
2. Validate and construct BusinessEvent domain objects.
3. Convert BusinessEvents into WorkItems.
4. Pass WorkItems into a trivial optimiser.
5. Produce an OptimisedPlan.
6. Print a readable summary to the console.

## Scope

Implement the smallest useful walking skeleton.

The optimiser may be deliberately simple.

The first optimiser does not need to consider resources, constraints, objectives, travel time, skills, availability or fairness.

## Required Packages

Use existing architecture:

- domain
- application
- infrastructure
- optimisation
- cmd/owp

## Responsibilities

The CLI should:

- parse command arguments
- call application behaviour
- print output

The CLI must not contain business logic.

The infrastructure layer should:

- load JSON datasets
- convert raw input into domain objects

The application layer should:

- orchestrate the workflow
- convert BusinessEvents into WorkItems for the first simple scenario
- call the optimiser

The optimisation layer should:

- accept WorkItems
- return an OptimisedPlan

The domain layer should:

- contain BusinessEvent, WorkItem and OptimisedPlan domain objects

## Non-Goals

Do not implement:

- real optimisation
- resources
- constraints
- objectives
- scheduling
- routing
- AI
- external integrations
- persistence
- web UI

## Dataset

Create:

examples/datasets/simple-events.json

The dataset should contain a small number of Business Events.

The BusinessEvent details should be opaque JSON.

For the walking skeleton only, the application layer may use the event type and details to create simple WorkItems.

## OptimisedPlan

Create the smallest useful OptimisedPlan domain object.

It should contain the WorkItems selected by the optimiser.

It should be immutable and serialisable where practical.

## Tests

Add tests for:

- dataset loading
- BusinessEvent to WorkItem conversion
- trivial optimiser
- CLI or application workflow where practical

## Constraints

Use only the Go standard library.

Do not introduce external dependencies.

Do not introduce unnecessary interfaces or abstractions.

If an abstraction is introduced, explain why it has earned its place.

## Definition of Done

The walking skeleton is complete when:

- go test ./... passes
- the target command runs successfully
- the command prints a readable Optimised Plan
- no external dependencies are introduced
- implementation respects the steering documents

## Open Questions

None.