package dnscli

import (
	"google.golang.org/api/dns/v1"
	"golang.org/x/oauth2/google"
	"log"
	"io/ioutil"
	"context"
)

type GoogleProvider struct {
	project string
	client  *dns.Service
}

func (s *GoogleProvider) SetA(Record, Value string, TLL int64) error {
	panic("implement me")
}

func (s *GoogleProvider) SetAAAA(Record, Value string, TLL int64) error {
	panic("implement me")
}

func (s *GoogleProvider) SetCNAME(Record, Value string, TLL int64) error {
	panic("implement me")
}

func (s *GoogleProvider) SetTXT(Record, Value string, TLL int64) error {
	panic("implement me")
}

func (s *GoogleProvider) List(zone string) ([]DNSRecord, error) {
	recs, err := s.client.ResourceRecordSets.List(s.project, zone).Do()
	if err != nil {
		return nil, err
	}
	found := make([]DNSRecord, 0)
	for _, r := range recs.Rrsets {
		if r.Type == "TXT" || r.Type == "A" || r.Type == "AAAA" || r.Type == "CNAME" {
			found = append(found, DNSRecord{
				Name:  r.Name,
				Type:  r.Type,
				TTL:   r.Ttl,
				Datas: r.Rrdatas,
			})
		}
	}
	return found, nil
}

func NewGoogleProvider(info map[string]string) (DNSProvider) {
	project, ok := info["Project"]
	if !ok || project == "" {
		log.Fatal("Google Cloud project name missing")
	}
	saFile, ok := info["SaFile"]
	if !ok || saFile == "" {
		log.Fatal("Google Cloud Service Account file missing")
	}
	dat, err := ioutil.ReadFile(saFile)
	if err != nil {
		log.Fatalf("Unable to read Service Account file: %v", err)
	}
	conf, err := google.JWTConfigFromJSON(dat, dns.NdevClouddnsReadwriteScope)
	if err != nil {
		log.Fatalf("Unable to acquire config: %v", err)
	}
	client := conf.Client(context.Background())

	svc, err := dns.New(client)
	if err != nil {
		log.Fatalf("Unable to create Google Cloud DNS service: %v", err)
	}
	return &GoogleProvider{
		project: project,
		client:  svc,
	}
}
