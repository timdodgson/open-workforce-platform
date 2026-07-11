# Publishing owp-sdk

Module: `github.com/timdodgson/open-workforce-platform/owp-sdk`  
Sources: `platform/owp-sdk/` in the monorepo.

## Tag (from repo root)

```bash
git tag -a platform/owp-sdk/v0.1.0 -m "owp-sdk v0.1.0 — BYOD Problem registry"
git push origin platform/owp-sdk/v0.1.0
```

Go module proxy resolves subdirectory modules when the tag matches `platform/owp-sdk/v0.1.0`.

## Consumer install

```bash
go get github.com/timdodgson/open-workforce-platform/owp-sdk@v0.1.0
```

Monorepo clones should keep:

```go
replace github.com/timdodgson/open-workforce-platform/owp-sdk => ../platform/owp-sdk
```

## Release checklist

- [ ] `go test ./...` in `platform/owp-sdk` (if tests added)
- [ ] Update `CHANGELOG.md`
- [ ] Tag as above
- [ ] Verify `examples/byod-tsp` builds with tagged version
