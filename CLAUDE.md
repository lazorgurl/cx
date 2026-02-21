# cx

CLI tool for managing multiple Claude Code credential profiles. Built with Go and [Cobra](https://github.com/spf13/cobra).

## Build & Run

```bash
go build -o cx .       # build
go install .           # install to $GOPATH/bin
go test ./...          # run all tests
```

## Project Structure

```
main.go                          # entrypoint, calls cmd.Execute()
cmd/                             # Cobra command definitions (one file per command)
  root.go                        # root command, registers subcommands
  save.go, use.go, list.go, ...  # subcommands: save, use, list, rm, current, status, refresh
internal/
  config/paths.go                # all filesystem path resolution (Claude config, cx config, profiles)
  profile/
    profile.go                   # core logic: save/use/list/remove profiles, detect active, parse creds
    credentials.go               # atomic file read/write, SHA-256 hashing for credential comparison
    livecreds.go                 # ReadLiveCredentials/WriteLiveCredentials: platform-dispatching live cred access
    keychain_darwin.go           # macOS Keychain read/write via security CLI (build-tagged darwin)
    state.go                     # active profile state persistence (~/.config/cx/state.json)
    refresh.go                   # OAuth token refresh against Claude platform
```

## Key Conventions

- **Atomic writes**: all credential and state file writes use temp-file-then-rename to avoid corruption.
- **Permissions**: credential files are written with 0600, directories with 0700.
- **CLAUDE_CONFIG_DIR**: respected throughout for non-default Claude config locations.
- **Profile names**: alphanumeric, hyphens, underscores; max 64 chars; validated by `profile.ValidateName`.
- **Token flush on switch**: `use` flushes live credentials back to the previously active profile before switching, preserving any silent token refreshes.
- **Platform-specific live credentials**: on macOS, live credentials are read/written via the system Keychain (`"Claude Code-credentials"` service); on Linux and other platforms, the file `~/.claude/.credentials.json` is used. Profile storage is always file-based under `~/.config/cx/`.

## Adding a New Command

1. Create `cmd/<name>.go` with a `cobra.Command` var.
2. Register it in `cmd/root.go` `init()` via `rootCmd.AddCommand(...)`.
3. Core logic belongs in `internal/profile/`; commands should be thin wrappers.
