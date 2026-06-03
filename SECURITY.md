# Security Policy

## Supported versions

gcloudenv is pre-1.0. Security fixes are applied to the latest released version.
Please make sure you're on the most recent release before reporting.

## Reporting a vulnerability

Please **do not** open a public issue for security problems.

Report privately via GitHub's
[private vulnerability reporting](https://github.com/figverse/gcloudenv/security/advisories/new).

Include, where possible:

- a description of the issue and its impact,
- steps to reproduce or a proof of concept,
- the gcloudenv version (`gcloudenv --version`) and your OS/shell.

We aim to acknowledge reports within 5 business days and to provide a remediation
timeline after triage. We'll credit reporters in the release notes unless you
prefer to remain anonymous.

## Scope notes

gcloudenv shells out to the `gcloud` CLI and never stores credentials itself —
accounts, projects, and ADC remain managed by `gcloud`. Reports about how
gcloudenv invokes `gcloud`, handles `.gcloudenv` files, or emits shell
statements for `eval` are in scope.
