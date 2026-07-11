# owp-sdk

Stable Go SDK for bring-your-own-domain (BYOD) integrations with the Open Workforce Platform.

**Module:** `github.com/timdodgson/open-workforce-platform/owp-sdk`  
**Version:** v0.1.0 (git tag `owp-sdk/v0.1.0` on the monorepo)

## Install

Published from the monorepo (submodule path `platform/owp-sdk`):

```bash
go get github.com/timdodgson/open-workforce-platform/owp-sdk@v0.1.0
```

Monorepo development:

```bash
# In your go.mod
require github.com/timdodgson/open-workforce-platform/owp-sdk v0.1.0
replace github.com/timdodgson/open-workforce-platform/owp-sdk => ../platform/owp-sdk
```

Go resolves the module from the repository root; the import path is `github.com/timdodgson/open-workforce-platform/owp-sdk` with sources under `platform/owp-sdk/`.

## Packages

| Package | Role |
|---------|------|
| `searchdef` | Core `Problem` interface and search assist types |
| `sdk` | `RegisterProblem` registry for domain loaders |

Search execution (`RunSearch`, policy hooks) lives in `platform/go/internal/optimisation` and is bridged via `platform/go/internal/sdk` for the `owp` CLI.

## Bring your own domain

1. Implement `searchdef.Problem` (solutions, moves, `Evaluate`).
2. Register with `sdk.RegisterProblem` — set `Title`, `PolicyDomain`, `ObjectiveLabel` for generic CLI display.
3. Blank-import your package from `cmd/owp` (or a custom binary).
4. Run: `owp solve <name> --instance <path>`

Generic solve provides display, search, and `run.json` finalize when `--run-label` is set. Rich domain-specific telemetry (e.g. `discoveries.csv`) requires platform hooks in `cmd/owp` or a custom finalize in your binary.

See `examples/byod-tsp` for a complete minimal example.

## Usage

```go
import (
    "github.com/timdodgson/open-workforce-platform/owp-sdk/sdk"
    "github.com/timdodgson/open-workforce-platform/owp-sdk/searchdef"
)

func init() {
    sdk.RegisterProblem(sdk.ProblemDescriptor{
        Name:           "my-domain",
        Title:          "My Domain Solver",
        PolicyDomain:   "my-domain",
        ObjectiveLabel: "Cost",
        Load: func(path string) (searchdef.Problem, sdk.InstanceMeta, error) {
            // load instance, return Problem implementation
            return nil, sdk.InstanceMeta{}, nil
        },
    })
}
```

Wire the package from `cmd/owp` via blank import alongside `internal/sdk/builtin`.
