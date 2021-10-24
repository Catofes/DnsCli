package dnscli

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

func checkPreReq(req dns.RR, records []DNSRecord) int {
	if req.Header().Class == dns.ClassANY {
		if req.Header().Rrtype == dns.TypeANY {
			for _, v := range records {
				if req.Header().Name == fqdn(v.Name) {
					return dns.RcodeSuccess
				}
			}
			return dns.RcodeNXRrset
		} else {
			for _, v := range records {
				if req.Header().Name == dns.Fqdn(v.Name) && dns.TypeToString[req.Header().Rrtype] == v.Type {
					return dns.RcodeSuccess
				}
			}
			return dns.RcodeNXRrset
		}
	} else if req.Header().Class == dns.ClassNONE {
		if req.Header().Rrtype == dns.TypeANY {
			for _, v := range records {
				if req.Header().Name == fqdn(v.Name) {
					return dns.RcodeYXDomain
				}
			}
			return dns.RcodeSuccess
		} else {
			for _, v := range records {
				if req.Header().Name == dns.Fqdn(v.Name) && dns.TypeToString[req.Header().Rrtype] == v.Type {
					return dns.RcodeYXRrset
				}
			}
			return dns.RcodeSuccess
		}
	} else {
		return dns.RcodeNotImplemented
	}
}

func (s *Cli) handleUpdate(r *dns.Msg, m *dns.Msg) {
	d := r.Question[0].Name
	domain := s.findDomain(d)
	if domain == "" {
		m.SetRcode(r, dns.RcodeNameError)
		return
	}
	p := s.dnsProviders[domain]
	if len(r.Answer) > 0 {
		records, err := p.List(domain)
		if err != nil {
			log.Print(err)
			m.SetRcode(r, dns.RcodeServerFailure)
			return
		}
		for _, rr := range r.Answer {
			if rcode := checkPreReq(rr, records); rcode != dns.RcodeSuccess {
				m.SetRcode(r, rcode)
				return
			}
		}
	}
	if len(r.Ns) > 0 {
		for _, rr := range r.Ns {
			if rr.Header().Class == dns.ClassANY || rr.Header().Class == dns.ClassNONE {
				if rr.Header().Rrtype == dns.TypeANY {
					m.SetRcode(r, dns.RcodeNotImplemented)
					return
				} else {
					if _, err := p.Absent(domain, rr.Header().Name, dns.TypeToString[rr.Header().Rrtype]); err != nil {
						log.Print(err)
						m.SetRcode(r, dns.RcodeServerFailure)
						return
					}
				}
			} else if rr.Header().Class == dns.ClassINET {
				record := RR2DNSRecord([]dns.RR{rr})[0]
				if _, err := p.Present(domain, record.Name, record.Type, strings.Join(record.Datas, " "), record.TTL); err != nil {
					log.Print(err)
					m.SetRcode(r, dns.RcodeServerFailure)
					return
				}
			}
		}
	}
	m.SetRcode(r, dns.RcodeSuccess)
}

func (s *Cli) handleQuery(r *dns.Msg, m *dns.Msg) {
	if len(r.Question) != 1 {
		m.SetRcode(r, dns.RcodeNotImplemented)
		return
	}
	record := r.Question[0].Name
	recordType := dns.TypeToString[r.Question[0].Qtype]
	domain := s.findDomain(record)
	if domain == "" {
		m.SetRcode(r, dns.RcodeNameError)
		return
	}
	p := s.dnsProviders[domain]
	records, err := p.List(domain)
	if err != nil {
		log.Print(err)
		m.SetRcode(r, dns.RcodeServerFailure)
		return
	}
	if recordType == "ANY" {
		for _, v := range records {
			if record == v.Name {
				data := fmt.Sprintf("%s %d %s %s %s", v.Name, v.TTL, "IN", v.Type, strings.Join(v.Datas, " "))
				ans, err := dns.NewRR(data)
				if err != nil {
					log.Print(err)
					m.SetRcode(r, dns.RcodeServerFailure)
					return
				}
				m.Answer = append(m.Answer, ans)
			}
		}
	} else {
		for _, v := range records {
			if record == v.Name && recordType == v.Type {
				data := fmt.Sprintf("%s %d %s %s %s", v.Name, v.TTL, "IN", v.Type, strings.Join(v.Datas, " "))
				ans, err := dns.NewRR(data)
				if err != nil {
					log.Print(err)
					m.SetRcode(r, dns.RcodeServerFailure)
					return
				}
				m.Answer = append(m.Answer, ans)
			}
		}
	}
	m.SetRcode(r, dns.RcodeSuccess)
}

func (s *Cli) handler(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	//log.Printf("------REQ-----\n%s\n", r.String())
	if r.IsTsig() != nil {
		if w.TsigStatus() == nil {
			if r.Opcode == dns.OpcodeUpdate {
				s.handleUpdate(r, m)
				m.SetTsig(s.tsigName, s.tsigAlg, 300, time.Now().Unix())
			} else if r.Opcode == dns.OpcodeQuery {
				s.handleQuery(r, m)
				m.SetTsig(s.tsigName, s.tsigAlg, 300, time.Now().Unix())
			}
		} else {
			m.SetRcode(r, dns.RcodeNotAuth)
		}
	} else {
		if r.Opcode == dns.OpcodeUpdate {
			m.SetRcode(r, dns.RcodeNotAuth)
		}
		m.SetRcode(r, dns.RcodeRefused)
	}
	//log.Printf("-----RSP------\n%s\n", m.String())
	defer w.WriteMsg(m)
}

func (s *Cli) Listen() {
	var err error
	s.tsigAlg, s.tsigName, s.tsigSecret, err = s.Config.parseTsig()
	if err != nil {
		log.Fatal(err)
	}
	acceptFunc := func(dh dns.Header) dns.MsgAcceptAction {
		if isResponse := dh.Bits&32768 != 0; isResponse {
			return dns.MsgIgnore
		}

		// Don't allow dynamic updates, because then the sections can contain a whole bunch of RRs.
		opcode := int(dh.Bits>>11) & 0xF
		if opcode != dns.OpcodeQuery && opcode != dns.OpcodeUpdate {
			return dns.MsgRejectNotImplemented
		}
		return dns.MsgAccept
	}
	serverTCP := &dns.Server{
		Net:  "tcp",
		Addr: s.Config.Listen,
		TsigSecret: map[string]string{
			s.tsigName: s.tsigSecret,
		},
		MsgAcceptFunc: acceptFunc,
	}
	serverUDP := &dns.Server{
		Net:  "udp",
		Addr: s.Config.Listen,
		TsigSecret: map[string]string{
			s.tsigName: s.tsigSecret,
		},
		MsgAcceptFunc: acceptFunc,
	}
	handler := dns.HandlerFunc(s.handler)
	wg := sync.WaitGroup{}
	run := func(server *dns.Server) {
		server.Handler = handler
		err = server.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
		wg.Done()
	}
	wg.Add(1)
	go run(serverTCP)
	wg.Add(1)
	go run(serverUDP)
	wg.Wait()
}
