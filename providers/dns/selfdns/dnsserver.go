package selfdns

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type record struct {
	hostname string
	fqdn     string
	value    string
}

var rec record

func request(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	if r.Opcode == dns.OpcodeQuery {
		for _, q := range r.Question {
			if strings.ToLower(q.Name) != rec.fqdn {
				continue
			}
			switch q.Qtype {
			case dns.TypeSOA:
				rr, err := dns.NewRR(fmt.Sprintf("%s 10 IN SOA %s. admin.%s. %d %d %d %d %d", q.Name, rec.hostname, rec.hostname, time.Now().Unix(), 10, 10, 10, 10))
				if err != nil {
					log.Fatalf("Failed to create RR: %v", err)
				}
				m.Answer = append(m.Answer, rr)
			case dns.TypeNS:
				rr, err := dns.NewRR(fmt.Sprintf("%s 10 IN NS %s.", q.Name, rec.hostname))
				if err != nil {
					log.Fatalf("Failed to create RR: %v", err)
				}
				m.Answer = append(m.Answer, rr)
			case dns.TypeTXT:
				rr, err := dns.NewRR(fmt.Sprintf("%s 10 IN TXT %s", q.Name, rec.value))
				if err != nil {
					log.Fatalf("Failed to create RR: %v", err)
				}
				m.Answer = append(m.Answer, rr)
			}
		}
	}
	w.WriteMsg(m)
}

func (d *DNSProvider) Run() error {
	// レコードの設定
	rec = record{
		hostname: d.config.ServerHostname,
		fqdn:     d.config.fqdn,
		value:    d.config.value,
	}
	// DNSサーバの起動
	d.config.serverUDP = dns.Server{Addr: d.config.ListenAddress + ":53", Net: "udp"}
	d.config.serverTCP = dns.Server{Addr: d.config.ListenAddress + ":53", Net: "tcp"}

	dns.HandleFunc(".", request)
	go func() {
		err := d.config.serverUDP.ListenAndServe()
	}()

	go func() {
		d.config.serverTCP.ListenAndServe()
	}()

	return nil
}

func (d *DNSProvider) Stop() error {
	err := d.config.serverUDP.Shutdown()
	if err != nil {
		return err
	}
	return d.config.serverTCP.Shutdown()
}
