# Custom ServeMux for use with [github.com/miekg/dns](https://github.com/miekg/dns)

It curently supports the following queries (name + type):
* time.<zone> TXT: Returns the current UTC time in RFC3339 format.
* whoami.<zone> TXT: Returns the IP address and port of the connecting client.
