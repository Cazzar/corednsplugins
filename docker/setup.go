package docker

import (
	"strconv"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"

	"github.com/mholt/caddy"
)

var log = clog.NewWithPlugin("rdns")


type config struct{
	// Prefix string
	Suffix string
	TTL uint32
}

func defaultConfig() *config {
	conf := config{
		// Prefix: "",
		Suffix: "",
		TTL: 86400,
	}
	return &conf
}

func init() {
	caddy.RegisterPlugin(pluginname, caddy.Plugin{
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
			if !c.NextArg() { return plugin.Error(pluginname, c.Err("No General Suffix assigned")) }
			s := c.Val()
			conf.Suffix = s
			break;
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

	dockerplugin := Docker{Config: conf}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		dockerplugin.Next = next
		return dockerplugin
	})

	return nil
}

const pluginname = "docker"