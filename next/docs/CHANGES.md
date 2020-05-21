## Changes in v2, already done

General:
- `--recursive` is default for some commands, notably `chezmoi add`
- only diff format is git
- remove hg support
- remove source command (use git instead)
- `--include` option to many commands
- errors output to stderr, not stdout
- all paths printed with OS-specific path separator (except `chezmoi dump`)
- `--force` now global
- `--output` now global
- diff includes scripts
- archive includes scripts
- `encrypt` -> `encrypted` in chattr
- `--format` now global, don't use toml for dump
- remove `secret` commands
- remove `keyring` support
- `y`, `yes`, `on`, `n`, `no`, `off` recognized as bools
- added `promptBool`, `promptFloat`, `promptInt` functions to `chezmoi init`

Config file:
- rename `sourceVCS` to `git`
- use `gpg.recipient` instead of `gpgRecipient`

## Changes in v2, to be done

- debug Windows support
- finish `add` command
- add `status` command
- add encryption support
- port `doctor` command
- port `merge` command
- add more tests
- update documentation with changes