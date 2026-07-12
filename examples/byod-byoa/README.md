# BYOA — custom search mode example

The platform ships a **greedy** hill-climb mode registered via `sdk.RegisterSearch` in `platform/go/internal/sdk/byoa`.

It accepts only strictly improving moves (no simulated annealing temperature). Works on any `searchdef.Problem`, including the BYOD TSP demo.

## Run on TSP

From `platform/go`:

```bash
go run ./cmd/owp solve tsp \
  --instance ../../examples/byod-tsp/instances/tsp-5city.json \
  --mode greedy \
  --iterations 20000 \
  --seed 42 \
  --run-label demo-tsp-5city-greedy-s42
```

Expect: **Tour length 23** (baseline 28 → 23), feasible.

Compare with built-in SA:

```bash
go run ./cmd/owp solve tsp --instance ../../examples/byod-tsp/instances/tsp-5city.json \
  --mode sa --iterations 50000 --seed 42 --run-label demo-tsp-5city-sa-s42
```

## Bring your own algorithm

1. Implement `func(problem searchdef.Problem, config optimisation.SearchConfig) optimisation.SearchResult`.
2. Call `sdk.RegisterSearch("your-mode", runner)` from an `init()` in your module.
3. Blank-import that package from `cmd/owp` (or your binary).
4. Run `owp solve <domain> --mode your-mode ...`

Built-in modes (`sa`, `lahc`, `tabu`, `portfolio`, `adaptive`) still resolve via `optimisation.RunSearch` unless overridden.

See `platform/go/internal/sdk/byoa/greedy.go` for the minimal reference implementation.
