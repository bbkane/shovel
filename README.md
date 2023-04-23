# shovel

Make a lot of DNS requests and count the results! Useful for testing complex dynamic DNS records.

Pass multiple FQDNs, nameservers, record types, and client subnets, either via command line flags, a config, or a combo of both. shovel will dig all combinations of those and show you the results.

## Use

### With different client subnets

![./demo.gif](./demo.gif)

### With different record types

This uses the same config as the above gif. No subnets passed, so that column is excluded from the output.

```bash
$ shovel dig --fqdn linkedin.com --rtype A --rtype AAAA
╭──────────────┬───────┬──────────────────┬─────────────────┬───────╮
│ FQDN         │ RTYPE │ NAMESERVER       │ ANS/ERR         │ COUNT │
├──────────────┼───────┼──────────────────┼─────────────────┼───────┤
│ linkedin.com │ A     │ # ns1            │ 13.107.42.14    │    10 │
│              │       │ 198.51.45.9:53   │                 │       │
│              │       ├──────────────────┼─────────────────┼───────┤
│              │       │ # dyn            │ 13.107.42.14    │    10 │
│              │       │ 108.59.161.43:53 │                 │       │
│              ├───────┼──────────────────┼─────────────────┼───────┤
│              │ AAAA  │ # ns1            │ 2620:1ec:21::14 │    10 │
│              │       │ 198.51.45.9:53   │                 │       │
│              │       ├──────────────────┼─────────────────┼───────┤
│              │       │ # dyn            │ 2620:1ec:21::14 │    10 │
│              │       │ 108.59.161.43:53 │                 │       │
╰──────────────┴───────┴──────────────────┴─────────────────┴───────╯
```

### Example config

Use `shovel dig -h` to see all possible config paths/flags:

```yaml
# ~/.config/shovel.yaml
dig:
  count: 10
  nameservers:
  - ns1
  - dyn
  # - google
  nameserver-map:
    ns1: 198.51.45.9:53
    dyn: 108.59.161.43:53
    google: 8.8.8.8:53
  # subnets:
  # - china
  # - usa
  subnet-map:
    china: 101.251.8.0
    usa: 100.43.128.0
```

## Install

- [Homebrew](https://brew.sh/): `brew install bbkane/tap/shovel`
- [Scoop](https://scoop.sh/):

```
scoop bucket add bbkane https://github.com/bbkane/scoop-bucket
scoop install bbkane/shovel
```

- Download Mac/Linux/Windows executable: [GitHub releases](https://github.com/bbkane/shovel/releases)
- Go: `go install go.bbkane.com/shovel@latest`
- Build with [goreleaser](https://goreleaser.com/) after cloning: `goreleaser --snapshot --skip-publish --clean`

## Notes

See [Go Developer Tooling](https://www.bbkane.com/blog/go-developer-tooling/) for notes on development tooling.

## Dev Notes

- Go doc: https://pkg.go.dev/github.com/miekg/dns

- `miekg/dns` exmple: https://github.com/miekg/exdns/blob/master/q/q.go



