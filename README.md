# shovel

Make a lot of DNS requests and count the results!

Most certainly not in a functional state yet :)

Note that this is not in Homebrew yet.

## Use

![./demo.gif](./demo.gif)

```bash
$ go run . dig --count 10 --fqdn www.linkedin.com --rtype CNAME --subnet-map china=101.251.8.0 --subnet-map usa=100.43.128.0 --subnet china --subnet usa --ns-map ns1=198.51.44.9:53 --ns-map dyn=108.59.161.43:53 --ns-map google=8.8.8.8:53 --ns ns1 --ns dyn --ns google
╭──────────────────┬───────┬──────────────┬──────────────────┬───────────────────────────────────────┬───────╮
│ FQDN             │ RTYPE │ SUBNET       │ NAMESERVER       │ ANS/ERR                               │ COUNT │
├──────────────────┼───────┼──────────────┼──────────────────┼───────────────────────────────────────┼───────┤
│ www.linkedin.com │ CNAME │ # china      │ # ns1            │ pop-az-be.www.linkedin.com.           │     5 │
│                  │       │ 101.251.8.0  │ 198.51.44.9:53   │                                       │       │
│                  │       │              │                  │ pop-az-sh.www.linkedin.com.           │     5 │
│                  │       │              │                  │                                       │       │
│                  │       │              ├──────────────────┼───────────────────────────────────────┼───────┤
│                  │       │              │ # dyn            │ pop-az-be-lor1.www.linkedin.com.      │     4 │
│                  │       │              │ 108.59.161.43:53 │ pop-az-be.www.linkedin.com.           │       │
│                  │       │              │                  │ pop-az-sh-lor1.www.linkedin.com.      │     6 │
│                  │       │              │                  │ pop-az-sh.www.linkedin.com.           │       │
│                  │       │              ├──────────────────┼───────────────────────────────────────┼───────┤
│                  │       │              │ # google         │ non-success rcode: REFUSED            │    10 │
│                  │       │              │ 8.8.8.8:53       │                                       │       │
│                  │       ├──────────────┼──────────────────┼───────────────────────────────────────┼───────┤
│                  │       │ # usa        │ # ns1            │ www-linkedin-com.l-0005.l-msedge.net. │    10 │
│                  │       │ 100.43.128.0 │ 198.51.44.9:53   │                                       │       │
│                  │       │              ├──────────────────┼───────────────────────────────────────┼───────┤
│                  │       │              │ # dyn            │ www-linkedin-com.l-0005.l-msedge.net. │    10 │
│                  │       │              │ 108.59.161.43:53 │                                       │       │
│                  │       │              ├──────────────────┼───────────────────────────────────────┼───────┤
│                  │       │              │ # google         │ non-success rcode: REFUSED            │    10 │
│                  │       │              │ 8.8.8.8:53       │                                       │       │
╰──────────────────┴───────┴──────────────┴──────────────────┴───────────────────────────────────────┴───────╯
```

```bash
$ go run . dig --fqdn linkedin.com --rtype A --rtype AAAA --ns-map ns1=198.51.44.9:53 --ns-map dyn=108.59.161.43:53 --ns ns1 --ns dyn
╭──────────────┬───────┬────────┬──────────────────┬─────────────────┬───────╮
│ FQDN         │ RTYPE │ SUBNET │ NAMESERVER       │ ANS/ERR         │ COUNT │
├──────────────┼───────┼────────┼──────────────────┼─────────────────┼───────┤
│ linkedin.com │ A     │        │ # ns1            │ 13.107.42.14    │     1 │
│              │       │        │ 198.51.44.9:53   │                 │       │
│              │       │        ├──────────────────┼─────────────────┼───────┤
│              │       │        │ # dyn            │ 13.107.42.14    │     1 │
│              │       │        │ 108.59.161.43:53 │                 │       │
│              ├───────┤        ├──────────────────┼─────────────────┼───────┤
│              │ AAAA  │        │ # ns1            │ 2620:1ec:21::14 │     1 │
│              │       │        │ 198.51.44.9:53   │                 │       │
│              │       │        ├──────────────────┼─────────────────┼───────┤
│              │       │        │ # dyn            │ 2620:1ec:21::14 │     1 │
│              │       │        │ 108.59.161.43:53 │                 │       │
╰──────────────┴───────┴────────┴──────────────────┴─────────────────┴───────╯
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
- map value type
- golden tests

Let's make a simpler mockup

## Mockup v1

Just dig one thing :)

```bash
shovel dig \
    --fqdn linkedin.com \
    --type A \
    --ns-ip 1.2.3.4 \
    --subnet-addr 1.2.3.4 \
    --timeout 2s \
```

## Output ideas

### dig + table

```
# ns: us=1.2.3.4, #subnet: china=1.2.3.0, count=100, timeout=2s
# dig +subnet=1.2.3.0 @1.2.3.4 linkedin.com A

| Answers         | Count |
| --------------- | ----- |
| 1.2.3.4 3.4.5.6 | 4     |
```

I don't like the way answers can balloon, but newlines in markdown tablers are awkward..

### dig + yaml

```
# ns: us=1.2.3.4, #subnet: china=1.2.3.0, count=100, timeout=2s
# dig +subnet=1.2.3.0 @1.2.3.4 linkedin.com A

answers:
- answers:
  - 1.2.3.4
  - 3.4.5.6
  count: 50
- answers:
  - 8.2.3.4
  - 9.4.5.6
  count: 40
errors:
- error: timeout
  count: 10
- error: something else
  count: 5
```

Ok, looks good to start at least!

## table

Subnet, ns, fqdn, rtype, count, timeout, answer, answer_count, error, error_count

this can have multiple answer sets.... and multiple error sets... need table coalescing or some sort of error

https://github.com/jedib0t/go-pretty/tree/main/table

I can make it look good with this I'm sure :)

## Next steps

