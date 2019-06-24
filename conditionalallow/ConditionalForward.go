package conditionalallow

import (
	"net"
	"context"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

//ServeDNS implements interface
func (c ConditionalAllow) ServeDNS(ctx context.Context, w dns.ResponseWriter, msg *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: msg}

	for _, ipnet := range(c.Config.ForwardIPs) {
		if (ipnet.Contains(net.ParseIP(state.IP()))) {
			return plugin.NextOrFailure(c.Name(), c.Next, ctx, w, msg)
		}
	}
	return dns.RcodeRefused, nil
}

//Name implements interface
func (c ConditionalAllow) Name() string { return pluginname }