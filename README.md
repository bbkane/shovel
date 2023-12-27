# shovel

Make a lot of DNS requests and count the results! Useful for testing complex dynamic DNS records.

Pass multiple qnames, nameservers, record types, and client subnets, either via command line flags, a config, or a combo of both. shovel will dig all combinations of those and show you the results.

## Use

Also see [examples.md](./examples.md)

### With different client subnets

![./demo.gif](./demo.gif)

### With different record types

This uses the same config as the above gif. No subnets passed, so that column is excluded from the output.

```bash
$ shovel dig --qname linkedin.com --rtype A --rtype AAAA
╭──────────────┬───────┬──────────────────┬─────────────────┬───────╮
│ QNAME         │ RTYPE │ NAMESERVER       │ ANS/ERR         │ COUNT │
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

## Run the webapp locally with [OpenObserve](https://openobserve.ai/)

Export env vars:

```bash
export SHOVEL_SERVE_OPENOBSERVE_PASS='...';
export SHOVEL_SERVE_OPENOBSERVE_USER='...';
export ZO_ROOT_USER_EMAIL='...';
export ZO_ROOT_USER_PASSWORD='...';
```

Run OpenObserve (in another terminal) after downloading:

```bash
./openobserve
```

Open OpenObserve at: http://localhost:5080/web/traces?period=15m&query=&org_identifier=default

Run shovel. Check `go run . serve --help` to see all flags available

```bash
go run . serve \
  --otel-provider openobserve
```

Open shovel at: http://127.0.0.1:8080/?count=1&nameservers=dns3.p09.nsone.net%3A53&protocol=udp&qnames=linkedin.com+www.linkedin.com&rtypes=A&subnetMap=&subnets=

Install shovel + OpenObserve as systemd services, on a local dev VM or production VM with [shovel_ansible](https://github.com/bbkane/shovel_ansible/)

## Dev Notes

- Go doc: https://pkg.go.dev/github.com/miekg/dns
- `miekg/dns` exmple: https://github.com/miekg/exdns/blob/master/q/q.go
- Look up IP subnets for a country: http://www.nirsoft.net/countryip/ or https://ipinfo.io/countries



