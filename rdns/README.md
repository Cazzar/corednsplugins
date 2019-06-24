rDNS
==========

My own crafting for an rDNS generation for my own local IPv4 and IPv6 ranges.
The numbering format is simply the standard shortest form replacing any `:` or `.` with `-`

## Using

Example 
```
2001:db8::/32 ipv6.example.com {
    rdns {
        prefix "ip"
        suffix ".ipv6.example.com"
    }
}

192.168.0.0/24 ipv4.example.com {
    rdns {
        prefix "ip"
        suffix ".ipv4.example.com"
    }
}

```

Or you can even combine all of them together

```
2001:db8::/32 ipv6.example.com 192.168.0.0/24 ipv4.example.com {
    rdns {
        v6-prefix "ip"
        v6-suffix ".ipv6.example.com"
        v4-prefix "ip"
        v4-suffix ".ipv4.example.com"
    }
}
```
### Options
`suffix`: This sets both `v6-suffix` and `v4-suffix` to the specified value.
`vN-suffix`: This sets the suffix for the specified IP version, so that you can have individual suffixes

`prefix`: Like `suffix` though allowing you to set the prefix
`vN-prefix`: the same as `vN-suffix` but as the prefix

_Please note:_ the leading dot is **REQUIRED**

`ttl`: set the TTL (default 86400)