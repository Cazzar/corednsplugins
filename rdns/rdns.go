package rdns

import (
	"context"
	"net"
	"strings"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

// RDNS plugin
type RDNS struct {
	Config *config
	Next plugin.Handler
}

// ServeDNS implements the plugin.Handler interface.
func (r RDNS) ServeDNS(ctx context.Context, w dns.ResponseWriter, msg *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: msg}
	if state.QType() == dns.TypePTR && strings.HasSuffix(state.Name(), ".ip6.arpa.") {
		return servePTRv6(ctx, w, msg, state, r)
	}
	if state.QType() == dns.TypeAAAA && strings.HasPrefix(state.Name(), r.Config.IPv6Prefix) && strings.HasSuffix(state.Name(), dns.Fqdn(r.Config.IPv6Suffix)) {
		return serveAAAAForPtr(ctx, w, msg, state, r)
	}
	if state.QType() == dns.TypePTR && strings.HasSuffix(state.Name(), ".in-addr.arpa.") {
		return servePTRv4(ctx, w, msg, state, r)
	}
	if state.QType() == dns.TypeA && strings.HasPrefix(state.Name(), r.Config.IPv4Prefix) && strings.HasSuffix(state.Name(), dns.Fqdn(r.Config.IPv4Suffix)) {
		return serveAForPtr(ctx, w, msg, state, r)
	}

	return plugin.NextOrFailure(r.Name(), r.Next, ctx, w, msg)
}

func servePTRv4(ctx context.Context, w dns.ResponseWriter, msg *dns.Msg, state request.Request, r RDNS) (int, error) {
	v4parts := strings.Split(strings.TrimSuffix(state.QName(), ".in-addr.arpa."), ".")
	var v4 strings.Builder
	for i := len(v4parts) - 1; i >= 0; i-- {
		v4.WriteString(v4parts[i])
		if (i != 0) { v4.WriteRune('-')}
	}

	response := new(dns.Msg)
	response.SetReply(msg)
	response.Authoritative = true

	hdr := dns.RR_Header{Name: state.QName(), Rrtype: dns.TypePTR, Class: state.QClass(), Ttl: r.Config.TTL}
	response.Answer = []dns.RR{&dns.PTR{
		Hdr: hdr,
		Ptr: dns.Fqdn(r.Config.IPv4Prefix + v4.String() + r.Config.IPv4Suffix),
	}}

	w.WriteMsg(response)

	return dns.RcodeSuccess, nil
}

func serveAForPtr(ctx context.Context, w dns.ResponseWriter, msg *dns.Msg, state request.Request, r RDNS) (int, error) {
	parsed := net.ParseIP(strings.ReplaceAll(strings.TrimPrefix(strings.TrimSuffix(state.Name(), dns.Fqdn(r.Config.IPv4Suffix)), r.Config.IPv4Prefix), "-", "."))

	response := new(dns.Msg)
	response.SetReply(msg)
	response.Authoritative = true

	hdr := dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeA, Class: state.QClass(), Ttl: r.Config.TTL}
	response.Answer = []dns.RR{&dns.A{
		Hdr: hdr,
		A: parsed,
	}}

	w.WriteMsg(response)
	return dns.RcodeSuccess, nil
}

func serveAAAAForPtr(ctx context.Context, w dns.ResponseWriter, msg *dns.Msg, state request.Request, r RDNS) (int, error) {
	parsedv6 := net.ParseIP(strings.ReplaceAll(strings.TrimPrefix(strings.TrimSuffix(state.Name(), dns.Fqdn(r.Config.IPv6Suffix)), r.Config.IPv6Prefix), "-", ":"))

	response := new(dns.Msg)
	response.SetReply(msg)
	response.Authoritative = true

	hdr := dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeAAAA, Class: state.QClass(), Ttl: r.Config.TTL}
	response.Answer = []dns.RR{&dns.AAAA{
		Hdr: hdr,
		AAAA: parsedv6,
	}}

	w.WriteMsg(response)
	return dns.RcodeSuccess, nil
}

func servePTRv6(ctx context.Context, w dns.ResponseWriter, msg *dns.Msg, state request.Request, r RDNS) (int, error) {
	parsedv6 := parseIPv6(state.QName())

	response := new(dns.Msg)
	response.SetReply(msg)
	response.Authoritative = true

	hdr := dns.RR_Header{Name: state.QName(), Rrtype: dns.TypePTR, Class: state.QClass(), Ttl: r.Config.TTL}
	response.Answer = []dns.RR{&dns.PTR{
		Hdr: hdr,
		Ptr: dns.Fqdn(r.Config.IPv6Prefix + parsedv6 + r.Config.IPv6Suffix),
	}}

	w.WriteMsg(response)
	return dns.RcodeSuccess, nil
}

func parseIPv6(s string) string {
	parts := strings.Split(strings.TrimSuffix(s, ".ip6.arpa."), ".")
	var ipv6 strings.Builder
	for i := len(parts) - 1; i >= 0; i-- {
		ipv6.WriteString(parts[i])
		if i%4 == 0 && i != 0 {
			ipv6.WriteString(":")
		}
	}

	ip := net.ParseIP(ipv6.String())
	ret := strings.ReplaceAll(ip.String(), ":", "-")
	if strings.HasSuffix(ret, "-") {
		return ret + "0"
	}
	return ret
}

// Name implements the Handler interface.
func (r RDNS) Name() string { return "rdns" }