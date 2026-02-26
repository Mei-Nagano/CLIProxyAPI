# Release v6.0.0-fork.1

Date: 2026-02-26
Branch: `feature/usage-stats-persist`
Commit: `6807bbf`

## Highlights

- Persist usage statistics across restarts.
- Restore usage snapshot on startup from `auth-dir/usage-statistics.json`.
- Auto-save snapshot every 30 seconds when request count changes.
- Flush snapshot on shutdown to minimize data loss.
- Keep compatibility with legacy raw `StatisticsSnapshot` JSON payloads.

## Files Changed

- `sdk/cliproxy/service.go`
- `internal/usage/persistence.go`
- `internal/usage/persistence_test.go`
- `FORK_USAGE_STATS_GUIDE.md`

## Validation

- `go test ./internal/usage ./sdk/cliproxy ./cmd/server` passed.

## Notes

- Ensure `usage-statistics-enabled: true` in `config.yaml`.
- Docker default persistent path: `./auths/usage-statistics.json`.
