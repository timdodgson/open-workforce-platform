# BYOD TSP example

Minimal bring-your-own-domain (BYOD) example for the Open Workforce Platform.

Implements a symmetric TSP with `searchdef.Problem` and registers via `owp-sdk`.

## Layout

```
tsp/           Problem + JSON loader
register.go    sdk.RegisterProblem in init()
instances/     Sample instances
```

## Wire into `owp`

In your `cmd/owp/main.go` (or a custom binary):

```go
import (
    _ "github.com/timdodgson/open-workforce-platform/platform/go/internal/sdk/builtin"
    _ "github.com/timdodgson/open-workforce-platform/examples/byod-tsp"
)
```

Add solve hooks in `cmd/owp` only if you need domain-specific telemetry CSVs beyond generic `run.json`.

## Run

From `platform/go` (copy-paste):

```bash
cd platform/go
go run ./cmd/owp list-solvers
go run ./cmd/owp solve tsp --instance ../../examples/byod-tsp/instances/tsp-5city.json --mode sa --iterations 50000 --seed 42
```

Expect: **Tour length 23** (constructive baseline 28 → 23), feasible.

Optional label for the lab UI:

```bash
go run ./cmd/owp solve tsp --instance ../../examples/byod-tsp/instances/tsp-5city.json --mode sa --iterations 50000 --seed 42 --run-label demo-tsp-5city-sa-s42
```

## Extend

1. Implement `searchdef.Problem` for your domain.
2. Call `sdk.RegisterProblem` from `init()`.
3. Blank-import your package from `cmd/owp`.
4. Run `go run ./cmd/owp solve <name> ...` — no `solve_hooks.go` changes needed (generic display + finalize via `ProblemDescriptor` fields).

### Custom search mode (BYOA)

The platform also registers a **greedy** hill-climb mode (`platform/go/internal/sdk/byoa`) — from `platform/go`:

```bash
go run ./cmd/owp solve tsp --instance ../../examples/byod-tsp/instances/tsp-5city.json --mode greedy --iterations 20000 --seed 42
```

Expect the same tour length **23** on this tiny instance. See `examples/byod-byoa/README.md`.

Search execution uses the platform `optimisation` engine via `internal/sdk.RunSearch`.
