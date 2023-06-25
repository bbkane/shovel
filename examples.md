Examples (assuming BASH-like shell):

# `shovel dig`

## Minimal no-config example

```bash
shovel dig \
    --fqdn linkedin.com \
    --ns dns.google:53
```

## Maximal no-config example

```bash
shovel dig \
    --count 20 \
    --fqdn linkedin.com \
    --fqdn www.linkedin.com \
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
    count: 10
    nameservers:
    - dyn
    # - google
    - ns1
    nameserver-map:
      dyn: ns1.p43.dynect.net.:53
      google: dns.google.:53
      ns1: dns1.p09.nsone.net.:53
    subnet-map:
      china: 101.251.8.0
      usa: 100.43.128.0
```

## Override values in the config with flags

```bash
shovel dig \
    --count 20 \
    --fqdn www.linkedin.com \
    --rtype CNAME \
    --subnet 100.43.128.0
```

## Override values in the config with flags

```bash
shovel dig \
    --count 20 \
    --fqdn www.linkedin.com \
    --rtype CNAME \
    --subnet 100.43.128.0
```

## Use 'all' to pass everything in --ns-map/--subnet-map

```bash
shovel dig \
    --fqdn www.linkedin.com \
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
shovel dig -f linkedin.com -r TXT -p udp4
```

