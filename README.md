# shovel

Make a lot of DNS requests and count the results! Useful for testing complex dynamic DNS records.

Pass multiple qnames, nameservers, record types, and client subnets, either via command line flags, a config, or a combo of both. shovel will dig all combinations of those and show you the results.

## Project Status (2025-06-14)

`shoel1` works well, but I don't use it anymore. I'm watching issues; please open one for any questions and especially BEFORE submitting a Pull Request.

## Use

Also see [examples.md](./examples.md)

![./demo.gif](./demo.gif)

You can save these flags to a config:

```yaml
# tmp.yaml
# see config paths with `shovel dig combine -h`
dig:
  combine:
    count: 10
    qnames:
      - example.com
      - www.example.com
    nameservers:
      - 1.1.1.1:53
    rtypes:
      - A
```

Run with:

```bash
shovel dig combine --config ./tmp.yaml
```

### Proxy DNS Traffic through a separate server with [`sshuttle`](https://sshuttle.readthedocs.io/en/stable/usage.html)

In one tab:

```bash
sshuttle --dns -r username@host --ns-hosts=8.8.8.8 0/0:53 ::/0:53
```

In another tab:

```bash
shovel dig combine -q example.com -r TXT -p tcp4 -n 8.8.8.8:53
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

Run shovel. Check `go run . serve --help` to see all flags available. Also see [format_jsonl.py]https://github.com/bbkane/dotfiles/blob/master/bin_common/bin_common/format_jsonl.py)

```bash
go run . serve | format_jsonl.py fmt
```

Open shovel at: http://127.0.0.1:8080/?count=1&nameservers=dns3.p09.nsone.net%3A53&protocol=udp&qnames=linkedin.com+www.linkedin.com&rtypes=A&subnetMap=&subnets=

Install shovel + OpenObserve as systemd services, on a local dev VM or production VM with [shovel_ansible](https://github.com/bbkane/shovel_ansible/)

## Dev Notes

- Go doc: https://pkg.go.dev/github.com/miekg/dns
- `miekg/dns` exmple: https://github.com/miekg/exdns/blob/master/q/q.go
- Look up IP subnets for a country: http://www.nirsoft.net/countryip/ or https://ipinfo.io/countries



