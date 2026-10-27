# Testing and verify by rendering upfront

Render a team without installing it, and check the output against the committed goldens. Back to the [README](../README.md).

```bash
kirokit render [team] --out <dir>   # render only, never touches testdata/golden/
kirokit render --quick [team] --out <dir>   # render the Amazon Quick bundle only (see amazon-quick.md)
kirokit test [team]                 # render + diff against testdata/golden/<team>/ (CI entry point)
kirokit update [team]               # render + overwrite testdata/golden/<team>/
```

No `[team]` = every directory under `teams/`. `test` exits non-zero and lists each mismatched file. `update` regenerates `testdata/golden/<team>/rendered/`, `context.json` and, for a team with a `quick` section, `quick/`; review the diff before committing.

- The Quick goldens (`testdata/golden/<team>/quick/`) render with the committed, made-up ids in `testdata/golden/quick-ids.yaml`, never with your own `quick-ids.yaml`, so they are the same on every machine. A `quick/` golden for a team without a `quick` section is a mismatch.
- `go test ./...` runs every team's golden comparison too. `go test ./internal/golden` also runs `kiro-cli agent validate` on the rendered Kiro agents; it skips without a logged-in `kiro-cli` unless `KIROKIT_REQUIRE_KIRO_CLI=1`.
