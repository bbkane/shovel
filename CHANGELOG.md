# Changelog

All notable changes to this project will be documented in this file. The format
is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

# v0.0.11 (Unreleased)

## Added

- Updated the README with instructions to run locally
- Add Trace ID to submit results
- Add version link to index.html

## Changed

- `open-observe` -> `openobserve` (mostly visible in flag names and values)
- made stdout tracer print jsonl instead of formatting. Use format_jsonl to format
- Moved README systemd notes to a link to [shovel_ansible](https://github.com/bbkane/shovel_ansible/)
- Allow `--http-origin` to contain `"request.Host"` in an experiment to see if we can let it use that instead of providing the HTTP origin as a flag. What does the client see?

## Fixed

- write fixes here
- A nil subnet doesn't show empty parens in the web form

## Removed

- `--mock-dig-func` - tests now specify dig funcs directly
- `--otel-service-name` and `--otel-service-version` - these are read from the app automatically now
