# Changelog

All notable changes to this project will be documented in this file. The format
is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

Note that the most recent version may be unreleased. See all releases on [GitHub](https://github.com/bbkane/shovel/releases).

# v0.0.18

## Fixed

- Determine filled form URL protocol from request instead of hardcoding to HTTP

# v0.0.17

Just a non-user-visible bug fix with the release process

# v0.0.16

## Added

- Add an ARM64 release

# v0.0.15

## Added

- Add HTTPS support to serve, allowing "Copy as YAML to clipboard" to work (see https://stackoverflow.com/questions/51805395/navigator-clipboard-is-undefined)

# v0.0.14

## Added

- Add metadata to "Copy as YAML to clipboard" YAMl

## Fixed

- "Copy as YAML to clipboard" results now use strings as rtypes

# v0.0.13

## Added

- Add "Copy table as YAML to clipboard" (#46)
- Add `serve --footer` flag to put HTML at the bottom. This will let me print geoip websites/other dns sites

## Changed

- `serve --otel-provider` now defaults to `stdout` and is can be overridden by the `SHOVEL_SERVE_OTEL_PROVIDER` environment variable.

# v0.0.12

## Added

- Add `--trace-id-template` flag so I can easily format my Trace IDs into links.

## Changed

- `--motd` now requires HTML and has a default

## Removed

- Remove `--http-origin` flag, instead reading from the client's request

# v0.0.11

## Added

- Updated the README with instructions to run locally
- Add Trace ID to submit results
- Add version link to index.html
- instrument DigOne for OpenTelemetry

## Changed

- `open-observe` -> `openobserve` (mostly visible in flag names and values)
- made stdout tracer print jsonl instead of formatting. Use format_jsonl to format
- Moved README systemd notes to a link to [shovel_ansible](https://github.com/bbkane/shovel_ansible/)
- Allow `--http-origin` to contain `"request.Host"` in an experiment to see if we can let it use that instead of providing the HTTP origin as a flag. What does the client see?

## Fixed

- A nil subnet doesn't show empty parens in the web form

## Removed

- `--mock-dig-func` - tests now specify dig funcs directly
- `--otel-service-name` and `--otel-service-version` - these are read from the app automatically now
