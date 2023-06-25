Examples (assuming BASH-like shell):

# `shovel dig combine` 

Builds dig queries based on combinations of input flags, then prints a table of results.

## Minimal no-config example

```bash
shovel dig combine \
    --qname linkedin.com \
    --ns dns.google:53
```

## Maximal no-config example

```bash
shovel dig combine \
    --count 20 \
    --qname linkedin.com \
    --qname www.linkedin.com \
    --ns cloudflare \
    --ns google \
    --ns-map cloudflare=1.1.1.1:53 \
    --ns-map google=8.8.8.8:53 \
    --rtype A \
    --rtype AAAA \
    --subnet china \
    --subnet usa \
    --subnet-map china=101.251.8.0 \
    --subnet-map usa=100.43.128.0 \
    --timeout 5s
```

## Example config

```yaml
## ~/.config/shovel.yaml
dig:
  combine:
    count: 1
    nameservers:
    - azure
    # - google
    - ns1
    nameserver-map:
      azure: ns1-42.azure-dns.com.:53
      google: dns.google.:53
      ns1: dns1.p09.nsone.net.:53
    subnet-map:
      china: 101.251.8.0
      usa: 100.43.128.0
```

## Override values in the config with flags

```bash
shovel dig combine \
    --count 20 \
    --qname www.linkedin.com \
    --rtype CNAME \
    --subnet 100.43.128.0
```

## Use 'all' to pass everything in --ns-map/--subnet-map

```bash
shovel dig combine \
    --qname www.linkedin.com \
    --rtype CNAME \
    --ns all \
    --subnet all
```

## Proxy DNS Traffic through a separate server with [`sshuttle`](https://sshuttle.readthedocs.io/en/stable/usage.html)

In one tab:
```bash
sshuttle --dns -r username@host --ns-hosts=ns1.p43.dynect.net,dns1.p09.nsone.net,ns1-42.azure-dns.com. 0/0:53 ::/0:53
```

In another tab:

```bash
shovel dig combine -q linkedin.com -r TXT -p tcp4
```

# `shovel dig list`

Pruduces digs from a list of inputs and prints the result as YAML

## Minimal no-config example

```bash
shovel dig list \
	--count 10 \
	--qname www.linkedin.com \
	--nameserver ns1-42.azure-dns.com.:53 \
	--protocol udp \
	--rtype A \
	--subnet 101.251.8.0 \
	--timeout 5s
```

## Example Config

```yaml
## /tmp/shovel.yaml
dig:
  list:
  - count: 10
    qname: www.linkedin.com
    nameserver: ns1-42.azure-dns.com.:53
    protocol: udp
    rtype: A
    subnet: 101.251.8.0
    timeout: 5s
  - count: 10
    qname: www.linkedin.com
    nameserver: ns1-42.azure-dns.com.:53
    protocol: udp
    rtype: AAAA  # different rtype
    subnet: 101.251.8.0
    timeout: 5s
```

```bash
shovel dig list --config /tmp/shovel.yaml  # read params from config
```

