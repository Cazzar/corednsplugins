Docker
==========

This is to create automatic PTR and A/AAA records for the docker daemon

**NOTE:** it is a good idea to have caching enabled in the Corefile due to possible delays with the docker daemon

## Requirements
The CoreDNS  server needs to be able to read the unix socket for docker (typically be part of the docker group)

## Using

Example 
```
docker.example.com 2001:db8::/64 172.19.0.0/16 {
    docker {
        suffix ".docker.example.com"
    }
    cache {
        success 1024
        denial 1024
    }
}
```

### Options
`suffix`: 

This is to specify the ending of the container, such as `.docker.example.com` with the container name `mariadb` would become the hostname `mariadb.docker.example.com`.
_Please note:_ the leading dot is **REQUIRED**

`ttl`: set the TTL (default 86400)