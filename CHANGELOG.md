# Changelog

All notable changes to this project will be documented in this file. The format
is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

# v0.0.11 (Unreleased)

## Added

- Updated the README with instructions to run locally

## Changed

- `open-observe` -> `openobserve` (mostly visible in flag names and values)
- made stdout tracer print jsonl instead of formatting. Use format_jsonl to format
- Moved README systemd notes to a link to [shovel_ansible](https://github.com/bbkane/shovel_ansible/)

## Fixed

- write fixes here
- A nil subnet doesn't show empty parens in the web form

## Removed

- `--mock-dig-func` - tests now specify dig funcs directly
- `--otel-service-name` and `--otel-service-version` - these are read from the app automatically now
