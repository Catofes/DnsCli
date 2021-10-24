package dnscli

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/miekg/dns"
)

type Rfc2136Provier struct {
	Host     string
	TsigName string
	TsigAlg  string
	Tsig     string
}

func (s *Rfc2136Provier) List(Domain string) ([]DNSRecord, error) {
	tsigSecret := map[string]string{
		s.TsigName: s.Tsig,
	}
	tr := dns.Transfer{
		TsigSecret: tsigSecret,
	}
	m := &dns.Msg{}
	m.SetAxfr(dns.Fqdn(Domain))
	m.SetTsig(s.TsigName, s.TsigAlg, 300, time.Now().Unix())
	channel, err := tr.In(m, s.Host)
	if err != nil {
		return nil, err
	}
	result := make([]DNSRecord, 0)
	for v := range channel {
		if v.Error != nil {
			return nil, err
		}
		result = append(result, RR2DNSRecord(v.RR)...)
	}
	return result, nil
}

func (s *Rfc2136Provier) query(Domain, record, recordType string) ([]dns.RR, error) {
	m := &dns.Msg{}
	if _, ok := dns.StringToType[recordType]; !ok {
		return nil, errors.New("wrong type")
	}
	m.SetQuestion(dns.Fqdn(record), dns.StringToType[recordType])
	in, err := dns.Exchange(m, s.Host)
	if err != nil {
		return nil, err
	}
	result := make([]dns.RR, 0)
	for _, v := range in.Answer {
		if v.Header().Name == dns.Fqdn(record) && dns.StringToType[recordType] == v.Header().Rrtype {
			result = append(result, v)
		}
	}
	return in.Answer, nil
}

func (s *Rfc2136Provier) Present(Domain, record, recordType, recordValue string, recordTTL int) (*RecordChanges, error) {
	r, err := s.query(Domain, record, recordType)
	if err != nil {
		return nil, err
	}
	tsigSecret := map[string]string{
		s.TsigName: s.Tsig,
	}
	c := dns.Client{
		Net:        "tcp",
		TsigSecret: tsigSecret,
	}
	m := &dns.Msg{}
	m.Id = dns.Id()
	m = m.SetUpdate(dns.Fqdn(Domain))
	m.RemoveRRset(r)
	rr, err := dns.NewRR(fmt.Sprintf("%s %d IN %s %s", dns.Fqdn(record), recordTTL, recordType, recordValue))
	if err != nil {
		return nil, err
	}
	m.Insert([]dns.RR{rr})
	m = m.SetTsig(s.TsigName, s.TsigAlg, 300, time.Now().Unix())
	in, _, err := c.Exchange(m, s.Host)
	if err != nil {
		return nil, err
	}
	if in.Rcode != dns.RcodeSuccess {
		return nil, errors.New(fmt.Sprintf("rfc2136 error, code: %d", in.Rcode))
	}
	RecordChanges := &RecordChanges{
		Delete: RR2DNSRecord(r),
		Add:    RR2DNSRecord([]dns.RR{rr}),
	}
	return RecordChanges, nil
}

func (s *Rfc2136Provier) Absent(Domain, record, recordType string) (*RecordChanges, error) {
	r, err := s.query(Domain, record, recordType)
	if err != nil {
		return nil, err
	}
	if len(r) <= 0 {
		return nil, errors.New("record not found")
	}
	tsigSecret := map[string]string{
		s.TsigName: s.Tsig,
	}
	c := dns.Client{
		Net:        "tcp",
		TsigSecret: tsigSecret,
	}
	m := &dns.Msg{}
	m.Id = dns.Id()
	m = m.SetUpdate(dns.Fqdn(Domain))
	m.RemoveRRset(r)
	m = m.SetTsig(s.TsigName, s.TsigAlg, 300, time.Now().Unix())
	in, _, err := c.Exchange(m, s.Host)
	if err != nil {
		log.Printf("%#v", in)
		return nil, err
	}
	if in.Rcode != dns.RcodeSuccess {
		return nil, errors.New(fmt.Sprintf("rfc2136 error, code: %d", in.Rcode))
	}
	RecordChanges := &RecordChanges{
		Delete: RR2DNSRecord(r),
	}
	return RecordChanges, nil
}

func NewRfc2135Provier(info map[string]string) DNSProvider {
	p := &Rfc2136Provier{}
	if v, ok := info["Tsig"]; ok {
		p.Tsig = v
	} else {
		log.Fatal("RFC2136: Tsig not found.")
	}
	if v, ok := info["Host"]; ok {
		p.Host = v
	} else {
		log.Fatal("RFC2136: Host not found.")
	}
	if v, ok := info["TsigName"]; ok {
		p.TsigName = v
	} else {
		log.Fatal("RFC2136: Tsig Name not found.")
	}
	if v, ok := info["TsigAlg"]; ok {
		p.TsigAlg = v
	} else {
		p.TsigAlg = "hmac-sha1."
	}
	return p
}
