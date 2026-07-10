# owp-sdk

Stable Go SDK for bring-your-own-domain (BYOD) integrations with the Open Workforce Platform.

## Packages

| Package | Role |
|---------|------|
| `searchdef` | Core `Problem` interface and search assist types |
| `sdk` | `RegisterProblem` registry for domain loaders |

Search execution (`RunSearch`, policy hooks) lives in `platform/go/internal/optimisation` and is bridged via `platform/go/internal/sdk` for the `owp` CLI.

## Usage (external domain)

```go
import (
    "github.com/timdodgson/open-workforce-platform/owp-sdk/sdk"
    "github.com/timdodgson/open-workforce-platform/owp-sdk/searchdef"
)

func init() {
    sdk.RegisterProblem(sdk.ProblemDescriptor{
        Name: "my-domain",
        Load: func(path string) (searchdef.Problem, sdk.InstanceMeta, error) {
            // load instance, return Problem implementation
        },
    })
}
```

Wire the package from `cmd/owp` via blank import alongside `internal/sdk/builtin`.
