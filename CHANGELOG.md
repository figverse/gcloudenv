# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Initial release of gcloudenv.
- `list` / `ls` — list gcloud profiles, marking the active one.
- `use <profile>` — switch the active profile per-shell (via the shim) or
  globally with `--global`.
- `current` / `status` — show the active profile, its account/project, and
  where the selection came from.
- `create <profile>` — create a profile, optionally seeding `--account` and
  `--project`.
- `local <profile>` — pin a profile to a directory via a `.gcloudenv` file.
- `init bash|zsh|fish` — print the shell integration snippet.
- Directory auto-switching on `cd` when a `.gcloudenv` file is present.

[Unreleased]: https://github.com/figverse/gcloudenv/commits/main
