package cds

import (
	"github.com/miekg/dns"
)

func handleWhoami(parentZone Zone, writer dns.ResponseWriter, request, response *dns.Msg) {
	switch request.Question[0].Qtype {
	case dns.TypeTXT:
		response.Answer = []dns.RR{
			&dns.TXT{
				Hdr: dns.RR_Header{
					Name:   request.Question[0].Name,
					Rrtype: dns.TypeTXT,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				Txt: []string{writer.RemoteAddr().String()},
			},
		}

	default:
		response.Ns = []dns.RR{
			&dns.SOA{
				Hdr: dns.RR_Header{
					Name:   request.Question[0].Name,
					Rrtype: dns.TypeSOA,
					Class:  dns.ClassINET,
					Ttl:    parentZone.TTL,
				},
				Ns:      parentZone.Ns,
				Mbox:    parentZone.Mbox,
				Serial:  parentZone.Serial,
				Refresh: parentZone.Refresh,
				Retry:   parentZone.Retry,
				Expire:  parentZone.Expire,
				Minttl:  parentZone.Minimum,
			},
		}
	}
}
