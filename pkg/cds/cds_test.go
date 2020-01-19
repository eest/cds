package cds_test

import (
	"fmt"
	"github.com/eest/cds/pkg/cds"
	"github.com/miekg/dns"
	"log"
	"testing"
)

func TestRequests(t *testing.T) {

	requestTests := []struct {
		desc   string
		rd     bool
		qzone  string
		qowner string
		qtype  uint16
		qclass uint16
		rauth  bool
		rerr   error
		rcode  int
	}{
		{
			desc:   "Type TXT time query for configured zone",
			qzone:  "example.com.",
			qowner: "time",
			qtype:  dns.TypeTXT,
			qclass: dns.ClassINET,
			rauth:  true,
			rerr:   nil,
			rcode:  dns.RcodeSuccess,
		},
		{
			desc:   "Type TXT TIME query for configured zone (capital letters)",
			qzone:  "example.com.",
			qowner: "TIME",
			qtype:  dns.TypeTXT,
			qclass: dns.ClassINET,
			rauth:  true,
			rerr:   nil,
			rcode:  dns.RcodeSuccess,
		},
		{
			desc:   "Type A time query for configured zone",
			qzone:  "example.com.",
			qowner: "time",
			qtype:  dns.TypeA,
			qclass: dns.ClassINET,
			rauth:  true,
			rerr:   nil,
			rcode:  dns.RcodeSuccess,
		},
		{
			desc:   "Type TXT time query for unconfigured zone",
			qzone:  "notconfigured.test.",
			qowner: "time",
			qtype:  dns.TypeTXT,
			qclass: dns.ClassINET,
			rerr:   nil,
			rcode:  dns.RcodeRefused,
		},
		{
			desc:   "Type A time query for unconfigured zone",
			qzone:  "notconfigured.test.",
			qowner: "time",
			qtype:  dns.TypeA,
			qclass: dns.ClassINET,
			rerr:   nil,
			rcode:  dns.RcodeRefused,
		},
		{
			desc:   "Type A query for nonexistant qname in configured zone",
			qzone:  "example.com.",
			qowner: "nonexistant",
			qtype:  dns.TypeA,
			qclass: dns.ClassINET,
			rerr:   nil,
			rauth:  true,
			rcode:  dns.RcodeNameError,
		},
		{
			desc:   "Type TXT whoami query for configured zone",
			qzone:  "example.com.",
			qowner: "whoami",
			qtype:  dns.TypeTXT,
			qclass: dns.ClassINET,
			rauth:  true,
			rerr:   nil,
			rcode:  dns.RcodeSuccess,
		},
		{
			desc:   "Type TXT WHOAMI query for configured zone (capital letters)",
			qzone:  "example.com.",
			qowner: "WHOAMI",
			qtype:  dns.TypeTXT,
			qclass: dns.ClassINET,
			rauth:  true,
			rerr:   nil,
			rcode:  dns.RcodeSuccess,
		},
		{
			desc:   "Type A whoami query for configured zone",
			qzone:  "example.com.",
			qowner: "whoami",
			qtype:  dns.TypeA,
			qclass: dns.ClassINET,
			rauth:  true,
			rerr:   nil,
			rcode:  dns.RcodeSuccess,
		},
	}

	// Start up an internal DNS server for test queries.
	internalDNSPort := "53535"
	internalDNSAddr := "127.0.0.1"
	internalDNSProto := "udp"

	serveMux := cds.NewServeMux(map[string]cds.Zone{
		"example.com.": {
			TTL:     300,
			Mname:   "mname.example.com.",
			Mbox:    "hostmaster.example.com.",
			Refresh: 100,
			Retry:   100,
			Expire:  100,
			Minimum: 100,
			Ns:      []string{"ns1.example.com.", "ns2.example.com."},
		},
	},
	)

	dnsServerReady := make(chan struct{})

	dnsServer := &dns.Server{
		Addr:    fmt.Sprintf("%s:%s", internalDNSAddr, internalDNSPort),
		Net:     internalDNSProto,
		Handler: serveMux,
		NotifyStartedFunc: func(dnsServerReady chan struct{}) func() {
			return func() {
				close(dnsServerReady)
			}
		}(dnsServerReady),
	}

	go dnsServer.ListenAndServe()
	// Wait until DNS server is ready to serve requests before continuing.
	<-dnsServerReady

	c := new(dns.Client)
	for _, test := range requestTests {

		// Build DNS message
		m := new(dns.Msg)
		m.Id = dns.Id()
		m.RecursionDesired = test.rd
		m.Question = make([]dns.Question, 1)
		m.Question[0] = dns.Question{Name: test.qowner + "." + test.qzone, Qtype: test.qtype, Qclass: test.qclass}

		// Create client and send message to server
		log.Printf("Sending type %s query for \"%s\"", dns.TypeToString[test.qtype], test.qowner+"."+test.qzone)
		r, _, err := c.Exchange(m, fmt.Sprintf("%s:%s", internalDNSAddr, internalDNSPort))

		if r.MsgHdr.Rcode != test.rcode {
			t.Errorf("%s: Expected %s, got %s", test.desc, dns.RcodeToString[test.rcode], dns.RcodeToString[r.MsgHdr.Rcode])
		}

		if r.MsgHdr.Authoritative != test.rauth {
			t.Errorf("%s: Expected %v, got %v", test.desc, test.rauth, r.MsgHdr.Authoritative)
		}

		if r != nil {
			if len(r.Answer) < 0 {
				if rr, ok := r.Answer[0].(*dns.TXT); ok {
					log.Printf("Content of TXT record: %s", rr.Txt[0])
				}
			}
		}

		if err != test.rerr {
			t.Errorf("%s: Expected %s, got %s", test.desc, test.rerr, err)
		}

	}
}
