package docker

import (
	"context"
	"net"
	"strings"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"


	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// Docker plugin
type Docker struct {
	Config *config
	Next plugin.Handler
}

// ServeDNS the function for the interface
func (r Docker) ServeDNS(ctx context.Context, w dns.ResponseWriter, msg *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: msg}

	if state.QType() == dns.TypeA {
		return r.serveA(ctx, w, msg, state)
	}
	if state.QType() == dns.TypeAAAA {
		return r.serveAAAA(ctx, w, msg, state)
	}
	if state.QType() == dns.TypePTR {
		return r.servePTR(ctx, w, msg, state)
	}

	return plugin.NextOrFailure(r.Name(), r.Next, ctx, w, msg)
}

func (r Docker) servePTR(ctx context.Context, w dns.ResponseWriter, msg *dns.Msg, state request.Request) (int, error) {
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.22", nil, map[string]string{"User-Agent": "engine-api-cli-1.0"})

	if err != nil {
		log.Error(err)
		return dns.RcodeServerFailure, err
	}
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Error(err)
		return dns.RcodeServerFailure, err
	}

	ipaddr, isv6 := convertQuestionToIP(state.Name())
	var rr dns.RR

	for _, container := range containers {
		for _, network := range(container.NetworkSettings.Networks) {
			if isv6 {
				if (ipaddr.Equal(net.ParseIP(network.GlobalIPv6Address))) {
					rr = &dns.PTR{
						Hdr: dns.RR_Header{Name: state.QName(), Rrtype: dns.TypePTR, Class: state.QClass(), Ttl: r.Config.TTL},
						Ptr: dns.Fqdn(strings.TrimPrefix(container.Names[0], "/") + r.Config.Suffix),
					}
					break
				}
			} else {
				if (ipaddr.Equal(net.ParseIP(network.IPAddress))) {
					rr = &dns.PTR{
						Hdr: dns.RR_Header{Name: state.QName(), Rrtype: dns.TypePTR, Class: state.QClass(), Ttl: r.Config.TTL},
						Ptr: dns.Fqdn(strings.TrimPrefix(container.Names[0], "/") + r.Config.Suffix),
					}
					break
				}
			}
		}
		if (rr != nil) { break }
	}

	resp := r.createResponse(state, []dns.RR{rr})
	w.WriteMsg(resp)

	return dns.RcodeSuccess, nil
}

func (r Docker) serveA(ctx context.Context, w dns.ResponseWriter, msg *dns.Msg, state request.Request) (int, error) {
	if !strings.HasSuffix(state.Name(), dns.Fqdn(r.Config.Suffix)) {
		return plugin.NextOrFailure(r.Name(), r.Next, ctx, w, msg)
	}

	container, err := getContainerSettings(strings.TrimSuffix(strings.TrimSuffix(state.Name(), dns.Fqdn(r.Config.Suffix)), "."))
	if err != nil {
		return dns.RcodeServerFailure, err
	}

	ips := make([]net.IP, 0)
	networkSettings := container.NetworkSettings
	for _, v := range(networkSettings.Networks) {
		ips = append(ips, net.ParseIP(v.IPAddress))
	}

	rr := make([]dns.RR, len(ips))
	for i := 0; i < len(ips); i++ {
		rr[i] = &dns.A{
			Hdr: dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeA, Class: state.QClass(), Ttl: r.Config.TTL},
			A: ips[i],
		}
	}

	resp := r.createResponse(state, rr)
	w.WriteMsg(resp)

	return dns.RcodeSuccess, nil
}

func (r Docker) serveAAAA(ctx context.Context, w dns.ResponseWriter, msg *dns.Msg, state request.Request) (int, error) {
	if !strings.HasSuffix(state.Name(), dns.Fqdn(r.Config.Suffix)) {
		return plugin.NextOrFailure(r.Name(), r.Next, ctx, w, msg)
	}

	container, err := getContainerSettings(strings.TrimSuffix(strings.TrimSuffix(state.Name(), dns.Fqdn(r.Config.Suffix)), "."))
	if err != nil {
		return dns.RcodeServerFailure, err
	}

	ips := make([]net.IP, 0)
	for _, v := range(container.NetworkSettings.Networks) {
		ips = append(ips, net.ParseIP(v.GlobalIPv6Address))
	}

	rr := make([]dns.RR, len(ips))
	for i := 0; i < len(ips); i++ {
		rr[i] = &dns.AAAA{
			Hdr: dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeAAAA, Class: state.QClass(), Ttl: r.Config.TTL},
			AAAA: ips[i],
		}
	}

	resp := r.createResponse(state, rr)
	w.WriteMsg(resp)

	return dns.RcodeSuccess, nil
}

func getContainerSettings(containerID string) (types.ContainerJSON, error) {
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.22", nil, map[string]string{"User-Agent": "engine-api-cli-1.0"})
	if err != nil {
		log.Error(err)
		return types.ContainerJSON{}, err
	}
	container, err := cli.ContainerInspect(context.Background(), containerID)
	if err != nil {
		log.Error(err)
		return types.ContainerJSON{}, err
	}
	return container, nil
}

func convertQuestionToIP(question string) (net.IP, bool) {
	if strings.HasSuffix(question, ".ip6.arpa.") {
		parts := strings.Split(strings.TrimSuffix(question, ".ip6.arpa."), ".")
		var ipv6 strings.Builder
		for i := len(parts) - 1; i >= 0; i-- {
			ipv6.WriteString(parts[i])
			if i%4 == 0 && i != 0 {
				ipv6.WriteString(":")
			}
		}

		return net.ParseIP(ipv6.String()), true
	}

	if strings.HasSuffix(question, ".in-addr.arpa.") {
		v4parts := strings.Split(strings.TrimSuffix(question, ".in-addr.arpa."), ".")
		var v4 strings.Builder
		for i := len(v4parts) - 1; i >= 0; i-- {
			v4.WriteString(v4parts[i])
			if (i != 0) { v4.WriteRune('.')}
		}

		return net.ParseIP(v4.String()), false
	}

	return nil, false
}


func (r Docker) createResponse(state request.Request, answer []dns.RR) *dns.Msg {
	response := new(dns.Msg)
	response.SetReply(state.Req)
	response.Authoritative = true
	response.Answer = answer

	return response
}

// Name implements the Handler interface.
func (r Docker) Name() string { return pluginname }