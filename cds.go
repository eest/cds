package cds

import (
	"github.com/miekg/dns"
	"log"
	"strings"
)

// ServeMux handles all incoming DNS requests.
type ServeMux struct {
	zones map[string]Zone
}

// Zone describes the SOA contents for a configured zone.
type Zone struct {
	TTL     uint32
	Mname   string
	Mbox    string
	Serial  uint32
	Refresh uint32
	Retry   uint32
	Expire  uint32
	Minimum uint32
	Ns      []string
}

// NewServeMux returns a configured ServeMux
func NewServeMux(zones map[string]Zone) *ServeMux {
	return &ServeMux{zones: zones}
}

// ServeDNS makes ServeMux implement the dns.Handler interface.
func (s *ServeMux) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {

	// We will always send a reply at this point, so set up initial
	// reply message.
	m := new(dns.Msg)
	m.SetReply(r)

	// We require at least one question in the question section.
	if len(r.Question) > 0 {

		// Make sure we have a lowercase version of the query name for easy
		// comparision
		questionName := strings.ToLower(r.Question[0].Name)

		// Check that we are configured to respond to the zone.
		var parentZone string
		for zone := range s.zones {
			if dns.IsSubDomain(zone, questionName) {
				parentZone = zone
				break
			}
		}

		if parentZone != "" {
			// The query belongs to a configured zone, any response
			// should be authoritative.
			m.Authoritative = true

			switch {
			case strings.HasPrefix(questionName, "time."):
				handleTime(s.zones[parentZone], r, m)

			case strings.HasPrefix(questionName, "whoami."):
				handleWhoami(s.zones[parentZone], w, r, m)

			default:
				m.MsgHdr.Rcode = dns.RcodeNameError
			}
		} else {
			// We are not configured for the queried zone, return
			// REFUSED.
			m.MsgHdr.Rcode = dns.RcodeRefused
		}
	} else {
		// There was no question in the question section, this is not expected.
		m.MsgHdr.Rcode = dns.RcodeFormatError
	}

	// Send the reply.
	err := w.WriteMsg(m)
	if err != nil {
		log.Printf("failed sending response: %s", err)
	}
}
