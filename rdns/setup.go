package rdns

import (
	"strconv"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"

	"github.com/mholt/caddy"
)

var log = clog.NewWithPlugin("rdns")

type config struct{
	IPv6Prefix string
	IPv6Suffix string
	IPv4Prefix string
	IPv4Suffix string
	TTL uint32
}

func defaultConfig() *config {
	conf := config{
		IPv6Prefix: "",
		IPv6Suffix: "",
		IPv4Prefix: "",
		IPv4Suffix: "",
		TTL: 86400,
	}
	return &conf
}

func init() {
	caddy.RegisterPlugin("rdns", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	c.Next() //Consume the directive itself.
	conf := defaultConfig()

	for c.NextBlock() {
		value := c.Val()
		switch value {
		case "suffix":
			if !c.NextArg() { return plugin.Error("rdns", c.Err("No General Suffix assigned")) }
			s := c.Val()
			conf.IPv4Suffix = s
			conf.IPv6Suffix = s
			break;
		case "prefix":
			if !c.NextArg() { return plugin.Error("rdns", c.Err("No General Prefix assigned")) }
			s := c.Val()
			conf.IPv4Prefix = s
			conf.IPv6Prefix = s
			break;
		case "v6-prefix":
			if !c.NextArg() { return plugin.Error("rdns", c.Err("No IPv6 Prefix assigned")) }
			conf.IPv6Prefix = c.Val()
			break
		case "v6-suffix":
			if !c.NextArg() { return plugin.Error("rdns", c.Err("No IPv6 Suffix assigned")) }
			conf.IPv6Suffix = c.Val()
			break
		case "v4-prefix":
			if !c.NextArg() { return plugin.Error("rdns", c.Err("No IPv4 Prefix assigned")) }
			conf.IPv6Suffix = c.Val()
			break
		case "v4-suffix":
			if !c.NextArg() { return plugin.Error("rdns", c.Err("No IPv4 Suffix assigned")) }
			conf.IPv6Suffix = c.Val()
			break
		case "ttl":
			if !c.NextArg() { return plugin.Error("rdns", c.Err("TTL was not provided"))}
			val, err := strconv.ParseUint(c.Val(), 10, 32)
			if err != nil { return plugin.Error("rdns", err) }
			conf.TTL = uint32(val)
		case "}":
			break
		case "{":
			break
		}
	}

	rdns := RDNS{Config: conf}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		rdns.Next = next
		return rdns
	})

	return nil
}
