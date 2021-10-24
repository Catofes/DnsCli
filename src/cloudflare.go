package dnscli

import (
	"log"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
)

const CloudFlareAPIURL = "https://api.cloudflare.com/client/v4"

type CloudflareProvider struct {
	client *cloudflare.API
}

func (s *CloudflareProvider) List(Domain string) ([]DNSRecord, error) {
	Domain = defqdn(Domain)
	id, err := s.client.ZoneIDByName(Domain)
	if err != nil {
		return nil, err
	}
	records, err := s.client.DNSRecords(id, cloudflare.DNSRecord{})
	if err != nil {
		return nil, err
	}
	result := make([]DNSRecord, 0)
	for _, v := range records {
		result = append(result, DNSRecord{
			fqdn(v.Name), v.Type, v.TTL, []string{v.Content},
		})
	}
	return result, nil
}

func (s *CloudflareProvider) Present(Domain, Record, Type, Value string, TTL int) (*RecordChanges, error) {
	Domain = defqdn(Domain)
	Record = defqdn(Record)
	id, err := s.client.ZoneIDByName(Domain)
	if err != nil {
		return nil, err
	}
	recordChanges := &RecordChanges{}
	records, err := s.client.DNSRecords(id, cloudflare.DNSRecord{Name: Record, Type: Type})
	if err != nil {
		return nil, err
	}
	for _, v := range records {
		err := s.client.DeleteDNSRecord(id, v.ID)
		if err != nil {
			return recordChanges, err
		}
		if recordChanges.Delete == nil {
			recordChanges.Delete = make([]DNSRecord, 0)
		}
		recordChanges.Delete = append(recordChanges.Delete, DNSRecord{
			fqdn(v.Name), v.Type, v.TTL, []string{v.Content},
		})
	}
	_, err = s.client.CreateDNSRecord(id, cloudflare.DNSRecord{
		ZoneID:  id,
		Name:    Record,
		Type:    Type,
		Content: Value,
		TTL:     TTL,
	})
	if err != nil {
		return recordChanges, err
	}
	recordChanges.Add = []DNSRecord{{
		fqdn(Record), Type, TTL, []string{Value},
	}}
	return recordChanges, nil
}

func (s *CloudflareProvider) Absent(Domain, Record, Type string) (*RecordChanges, error) {
	Domain = defqdn(Domain)
	Record = defqdn(Record)
	id, err := s.client.ZoneIDByName(Domain)
	if err != nil {
		return nil, err
	}
	recordChanges := &RecordChanges{}
	records, err := s.client.DNSRecords(id, cloudflare.DNSRecord{Name: Record, Type: Type})
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, errors.New("records not found")
	}
	for _, v := range records {
		err := s.client.DeleteDNSRecord(id, v.ID)
		if err != nil {
			return recordChanges, err
		}
		if recordChanges.Delete == nil {
			recordChanges.Delete = make([]DNSRecord, 0)
		}
		recordChanges.Delete = append(recordChanges.Delete, DNSRecord{
			fqdn(v.Name), v.Type, v.TTL, []string{v.Content},
		})
	}
	return recordChanges, nil
}

func NewCloudflareProvider(info map[string]string) DNSProvider {
	email, ok := info["Email"]
	if !ok {
		log.Fatal("Cloudflare email not set.")
	}
	key, ok := info["Key"]
	if !ok {
		log.Fatal("Cloudflare key not set.")
	}
	client, err := cloudflare.New(key, email)
	if err != nil {
		log.Fatal(err.Error())
	}
	return &CloudflareProvider{client}
}
