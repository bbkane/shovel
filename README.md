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

# Running with `systemd`

> ***This is very much a work in progress***

Folder hierarchy:

```
/etc/bkane/<service>/<version>
```

And that can hold things like executables, data, unit files, and config.

```bash
sudo mkdir -p /etc/bkane/shovel/v0.0.9
```

TODO: `shovel version` shows `0.0.9` instead of `v0.0.9`...

```ini
[Unit]
Description=Shovel DNS query frontend
After=syslog.target network-online.target remote-fs.target nss-lookup.target

[Service]
Type=simple
ExecStart=/home/linuxbrew/.linuxbrew/bin/shovel serve --config /etc/bkane/shovel/v0.0.9/shovel.yaml
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

```
serve:
  addr-port:  0.0.0.0:8080
  http-origin: http://bkane-ld2:8080
```

```bash
sudo systemctl enable /etc/bkane/shovel/v0.0.9/shovel.service
```

```bash
sudo systemctl start shovel
```

```bash
sudo systemctl status shovel
```

```bash
sudo journalctl -u shovel
```

[How to modify an existing systemd unit file | 2DayGeek](https://www.2daygeek.com/linux-modifying-existing-systemd-unit-file/)

```bash
sudo systemctl daemon-reload
```

```bash
sudo systemctl restart shovel
```

TODO: test reboots, add security to unit file, PR to shovel docs or blog post about systemd

TODO: https://www.freedesktop.org/software/systemd/man/systemd-analyze.html . See `systemd-analyze verify|security`

## Dev Notes

- Go doc: https://pkg.go.dev/github.com/miekg/dns
- `miekg/dns` exmple: https://github.com/miekg/exdns/blob/master/q/q.go
- Look up IP subnets for a country: http://www.nirsoft.net/countryip/ or https://ipinfo.io/countries



