# CliDiff

CLI interface contract breaking change detector. Catches deleted flags, changed defaults, and removed subcommands **before** your downstream scripts silently break.

## Install

```bash
go install github.com/clidiff/clidiff@latest
# or
go build -o clidiff .
```

## Usage

### Snapshot a CLI tool
```bash
clidiff snapshot kubectl              # → kubectl.snap
clidiff snapshot ./mytool out.snap    # custom output path
```

### Compare two snapshots
```bash
clidiff compare v1.0.0.snap v2.0.0.snap
# [BREAKING] flag-removed: --dry-run
# [BREAKING] default-changed: --output: "text"→"json"
# [MINOR] flag-added: --format
```

### CI guard mode
```bash
clidiff guard baseline.snap ./mytool
# exits non-zero if BREAKING changes detected
```

## Detection Matrix

| Change | Severity |
|---|---|
| Flag removed | BREAKING |
| Flag type changed | BREAKING |
| Default value changed | BREAKING |
| Subcommand removed | BREAKING |
| Flag added | MINOR |
| Subcommand added | MINOR |

## CI Integration

```yaml
- run: clidiff guard baseline.snap ./mytool
```

Check `.snap` files into your repo to track CLI contract over time.

## License

MIT
