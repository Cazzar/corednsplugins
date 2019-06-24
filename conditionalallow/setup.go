package conditionalallow

import (
	"net"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	//"github.com/coredns/coredns/plugin/pkg/parse"

	"github.com/mholt/caddy"
)

var log = clog.NewWithPlugin(pluginname)

//ConditionalForward data
type ConditionalAllow struct {
	Next plugin.Handler
	Config *config
}

type config struct{
	// Prefix string
	ForwardIPs []*net.IPNet
}

func defaultConfig() *config {
	nets := []string{
		"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "127.0.0.0/8", "169.254.0.0/16", //IPv4 private ranges, including APIPA
		"::1/128", "fe80::/10", "fc00::/7", //ipv6 private ranges, including link-local and unique local
	}
	ipnets := make([]*net.IPNet, len(nets))
	for i, network := range(nets) {
		_, ipnet, err := net.ParseCIDR(network)
		if (err != nil) {
			log.Errorf("Error parsing network [%s]: %s", network, err)
			continue
		}
		ipnets[i] = ipnet
	}

	conf := config{
		ForwardIPs: ipnets,
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
	c.Next() //consume the tag itself.
	conf := defaultConfig()

	to := c.RemainingArgs()
	if len(to) >= 1 {
		conf.ForwardIPs = make([]*net.IPNet, len(to))
		for i, network := range(to) {
			_, ipnet, err := net.ParseCIDR(network)
			if (err != nil) {
				log.Errorf("Error parsing network [%s]: %s", network, err)
			}
			conf.ForwardIPs[i] = ipnet
		}
	}

	dockerplugin := ConditionalAllow{Config: conf}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		dockerplugin.Next = next
		return dockerplugin
	})

	return nil
}

const pluginname = "limitto"