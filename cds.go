package cds

import (
	"github.com/miekg/dns"
	"strings"
)

// ServeMux handles all incoming DNS requests.
type ServeMux struct {
	zones map[string]Zone
}

// Zone describes the SOA contents for a configured zone.
type Zone struct {
	TTL     uint32
	Ns      string
	Mbox    string
	Serial  uint32
	Refresh uint32
	Retry   uint32
	Expire  uint32
	Minimum uint32
}

// NewServeMux returns a configured ServeMux
func NewServeMux(zones map[string]Zone) *ServeMux {
	return &ServeMux{zones: zones}
}

// ServeDNS makes ServeMux implement the dns.Handler interface.
func (s *ServeMux) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {

	// Check that we are configured to respond to the zone.
	var parentZone string
	for zone := range s.zones {
		if dns.IsSubDomain(zone, r.Question[0].Name) {
			parentZone = zone
			break
		}
	}

	// We will always send a reply at this point, so set up initial
	// reply message.
	m := new(dns.Msg)
	m.SetReply(r)

	if parentZone != "" {
		// The query belongs to a configured zone, any response
		// should be authoritative.
		m.Authoritative = true

		switch {
		case strings.HasPrefix(r.Question[0].Name, "time."):
			handleTime(s.zones[parentZone], r, m)

		case strings.HasPrefix(r.Question[0].Name, "whoami."):
			handleWhoami(s.zones[parentZone], w, r, m)

		default:
			m.MsgHdr.Rcode = dns.RcodeNameError
		}
	} else {
		// We are not configured for the queried zone, return
		// REFUSED.
		m.MsgHdr.Rcode = dns.RcodeRefused
	}

	// Send the reply.
	w.WriteMsg(m)
}
