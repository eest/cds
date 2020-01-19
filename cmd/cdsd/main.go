// Example usage of the cds ServeMux.
package main

import (
	"fmt"
	"github.com/eest/cds"
	"github.com/miekg/dns"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Set some initial values.
	dnsAddr := "127.0.0.1"
	dnsPort := "53"

	serveMux := cds.NewServeMux(map[string]cds.Zone{
		"example.com.": {
			TTL:     300,
			Mname:   "mname.example.com.",
			Mbox:    "hostmaster.example.com.",
			Refresh: 14400,
			Retry:   3600,
			Expire:  2419200,
			Minimum: 300,
			Ns:      []string{"ns1.example.com.", "ns2.example.com."},
		},
	})

	go func() {
		dnsServer := &dns.Server{
			Addr:    fmt.Sprintf("%s:%s", dnsAddr, dnsPort),
			Net:     "udp",
			Handler: serveMux,
		}
		log.Fatal(dnsServer.ListenAndServe())
	}()

	go func() {
		dnsServer := &dns.Server{
			Addr:    fmt.Sprintf("%s:%s", dnsAddr, dnsPort),
			Net:     "tcp",
			Handler: serveMux,
		}
		log.Fatal(dnsServer.ListenAndServe())
	}()

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig

	log.Fatalf("Signal (%v) received, stopping", s)
}
