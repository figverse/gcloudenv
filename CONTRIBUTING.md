# Contributing to gcloudenv

Thanks for your interest in improving gcloudenv! This document covers how to get
set up, the conventions we follow, and what to expect when you open a pull
request.

## Getting started

You'll need [Go](https://go.dev/dl/) 1.22 or newer and the `gcloud` CLI
installed for manual testing.

```sh
git clone https://github.com/figverse/gcloudenv
cd gcloudenv
make build      # build ./gcloudenv
make test       # run the unit tests
```

The test suite does **not** require a real `gcloud` — the `internal/gcloud`
tests run against a stub binary, so they're safe to run anywhere.

## Project layout

| Path                | Responsibility                                              |
| ------------------- | ----------------------------------------------------------- |
| `cmd/`              | Cobra commands (one file per command)                       |
| `internal/gcloud/`  | The only place that shells out to `gcloud`                  |
| `internal/profile/` | Profile resolution (flag > `.gcloudenv` > env > default)    |
| `internal/shell/`   | Embedded shell shims and shell-syntax helpers               |

If you're adding behaviour that touches gcloud, keep the `os/exec` calls inside
`internal/gcloud` so the rest of the code stays testable against a stub.

## Before you open a PR

Run the full local check — CI runs the same steps:

```sh
make check      # fmt-check + vet + lint + test
```

- **Formatting:** code must be `gofmt`-clean. `make fmt` fixes it.
- **Linting:** we use [golangci-lint](https://golangci-lint.run/). `make lint`.
- **Tests:** add or update tests for any behaviour change. Shell-shim changes
  should be smoke-tested in a real shell and described in the PR.
- **Commits:** we follow [Conventional Commits](https://www.conventionalcommits.org/)
  (`feat:`, `fix:`, `docs:`, `chore:`, …). This keeps the changelog tidy.

## Pull request expectations

- Keep PRs focused; one logical change per PR is easiest to review.
- Update `README.md` and `CHANGELOG.md` when you change user-facing behaviour.
- Describe how you tested the change, especially for shell integration.

## Reporting bugs and proposing features

Use the [issue templates](https://github.com/figverse/gcloudenv/issues/new/choose).
For anything security-related, please follow [SECURITY.md](SECURITY.md) instead
of opening a public issue.

## Code of Conduct

By participating you agree to abide by our
[Code of Conduct](CODE_OF_CONDUCT.md).
