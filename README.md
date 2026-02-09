# cx

A CLI tool for managing multiple Claude Code credential profiles. Save, switch between, and manage different accounts on the same machine.

## Installation

```bash
go install github.com/lazorgurl/cx@latest
```

Or build from source:

```bash
git clone https://github.com/lazorgurl/cx.git
cd cx
go build
```

## Usage

```bash
# Save your current credentials as a named profile
cx save work

# Save another set of credentials
cx save personal

# Switch between profiles
cx use work

# See all saved profiles (* marks active)
cx list

# Show the active profile name
cx current

# Show credential details (subscription, rate limit tier, expiry)
cx status

# Remove a profile
cx rm old-account
```

## How it works

cx copies Claude Code credentials (`~/.claude/.credentials.json`) into named profile directories under `~/.config/cx/profiles/`. Switching profiles copies the saved credentials back. All file writes are atomic and credentials are stored with owner-only permissions (0600).

The `CLAUDE_CONFIG_DIR` environment variable is respected if you use a non-default Claude config location.

## License

MIT - see [LICENSE](LICENSE).
