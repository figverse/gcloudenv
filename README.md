# gcloudenv

[![CI](https://github.com/figverse/gcloudenv/actions/workflows/ci.yml/badge.svg)](https://github.com/figverse/gcloudenv/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/figverse/gcloudenv.svg)](https://pkg.go.dev/github.com/figverse/gcloudenv)
[![Go Report Card](https://goreportcard.com/badge/github.com/figverse/gcloudenv)](https://goreportcard.com/report/github.com/figverse/gcloudenv)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Manage gcloud configurations the way [`nvm`](https://github.com/nvm-sh/nvm) and
[`rbenv`](https://github.com/rbenv/rbenv) manage language versions: switch the
active profile per-shell, set a global default, and auto-switch based on a
`.gcloudenv` file in your project.

gcloudenv is a thin layer over `gcloud config configurations` — gcloud remains
the source of truth for accounts, projects, and credentials. gcloudenv just
makes switching between them ergonomic and shell-aware.

## How it works

A child process can't change its parent shell's environment, so `gcloudenv`
ships a small shell function (the *shim*) that `eval`s the `export` lines the
binary prints. Switching a profile sets `CLOUDSDK_ACTIVE_CONFIG_NAME`, which
gcloud reads to pick the active configuration for that shell — no global state
touched.

## Install

With Go:

```sh
go install github.com/figverse/gcloudenv@latest
```

Or grab a prebuilt binary for your platform from the
[releases page](https://github.com/figverse/gcloudenv/releases).

Then add the integration to your shell rc:

```sh
# ~/.zshrc or ~/.bashrc
eval "$(gcloudenv init zsh)"

# fish: ~/.config/fish/config.fish
gcloudenv init fish | source
```

## Usage

```sh
gcloudenv list                 # all profiles; * marks the active one
gcloudenv use staging          # switch this shell to "staging"
gcloudenv use prod --global    # change gcloud's global default
gcloudenv current              # show active profile + account/project + source
gcloudenv create staging \     # create a profile and seed it
    --account me@x.com --project my-stg
gcloudenv local staging        # write .gcloudenv so this dir auto-switches
```

### Directory auto-switch

`gcloudenv local <profile>` writes a `.gcloudenv` file (like `.nvmrc`). With the
shim installed, `cd`-ing into that directory (or any child) switches the shell
to that profile automatically. Resolution precedence, highest first:

1. explicit argument (`gcloudenv use X`)
2. nearest `.gcloudenv` walking up from the cwd
3. `$CLOUDSDK_ACTIVE_CONFIG_NAME` already in the environment
4. gcloud's own global default

## Known limitation

gcloud configurations share a **single** application-default-credentials (ADC)
file — switching configurations does not isolate ADC. If you need per-profile
ADC for local development against multiple identities, that requires the
isolated-`CLOUDSDK_CONFIG` model, which is not yet implemented. The `gcloud`
interop is isolated behind `internal/gcloud` to make adding it straightforward.

## Develop

```sh
make build      # build ./gcloudenv
make test       # run unit tests (no real gcloud needed — uses a stub)
make check      # fmt-check + vet + lint + test, as CI runs them
```

Tests run against a stub `gcloud`, so they're safe to run anywhere.

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for setup,
conventions, and the PR checklist, and please follow our
[Code of Conduct](CODE_OF_CONDUCT.md). Security issues should be reported per
[SECURITY.md](SECURITY.md).

## License

[MIT](LICENSE) © Figverse
