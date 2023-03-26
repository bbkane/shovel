# shovel

Make a lot of DNS requests and count the results!

Most certainly not in a functional state yet :)

Note that this is not in Homebrew yet.

## Use

![./demo.gif](./demo.gif)

```bash
shovel hello
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

Go doc: https://pkg.go.dev/github.com/miekg/dns

Using EDNS: https://github.com/miekg/exdns/blob/master/q/q.go

### Params

- FQDN
  - 1..n
- NS
  - 0..n  # not sure the DNS library knows this?
- Subnet
  - 0..n
- Type
  - 1..n # Probablyy most common for A/AAAA records
- Response
  - Err | 1..n records

## Mockup

```bash
shovel dig \
    --fqdn linkedin.com \
    --fqdn a.linkedin.com \
    --nameserver nameserver1 \
    --nameserver dyn \
    --nameserver-map ns1=1.2.3.4 \
    --nameserver-map dyn=4.3.2.1 \
    --subnet us \
    --subnet china \
    --subnet-map us=6.7.8.9 \
    --subnet-map china=9.8.7.6 \
    --type A \
    --timeout 10s \
    --count 100 \
```

then with a config I could get that down to:

```bash
shovel dig \
    --fqdn linkedin.com \
    --fqdn a.linkedin.com \
    --type A \
    --subnet us \
    --subnet china
```

Some things warg really needs to make this ergonomic:

- tab completion!
- map value type'

Let's make a simpler mockup

## Mockup v1

Just dig one thing :)

```bash
shovel dig \
    --fqdn linkedin.com \
    --type A \
    --ns-ip 1.2.3.4 \
    --subnet-ip 1.2.3.4 \
    --timeout 2s \
```
