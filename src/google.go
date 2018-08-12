package dnscli

import (
	"google.golang.org/api/dns/v1"
	"golang.org/x/oauth2/google"
	"log"
	"io/ioutil"
	"context"
	"errors"
	"time"
)

type GoogleProvider struct {
	project string
	client  *dns.Service
}

func (s *GoogleProvider) getZoneName(Domain string) string {
	if Domain[len(Domain)-1] != '.' {
		Domain = Domain + string('.')
	}
	zones, err := s.client.ManagedZones.List(s.project).Do()
	if err!=nil{
                log.Printf("Get zone failed: %s.", err)
                return ""
        }
        zoneName := ""
	if zones.ManagedZones == nil {
		return ""
	}
	for _, zone := range zones.ManagedZones {
		if zone.DnsName == Domain {
			zoneName = zone.Name
		}
	}
	return zoneName
}

func (s *GoogleProvider) parseChange(chg *dns.Change) *RecordChanges {
	recordChanges := RecordChanges{}
	if chg == nil {
		return &recordChanges
	}
	if len(chg.Additions) > 0 {
		recordChanges.Add = make([]DNSRecord, 0)
		for _, v := range chg.Additions {
			recordChanges.Add = append(recordChanges.Add, DNSRecord{
				v.Name, v.Type, int(v.Ttl), v.Rrdatas,
			})
		}
	}
	if len(chg.Deletions) > 0 {
		recordChanges.Delete = make([]DNSRecord, 0)
		for _, v := range chg.Deletions {
			recordChanges.Delete = append(recordChanges.Delete, DNSRecord{
				v.Name, v.Type, int(v.Ttl), v.Rrdatas,
			})
		}
	}
	return &recordChanges
}

func (s *GoogleProvider) findDeleteRecords(ZoneName, Record, Type string) ([]*dns.ResourceRecordSet, error) {
	recs, err := s.client.ResourceRecordSets.List(s.project, ZoneName).Do()
	if err != nil {
		return nil, err
	}
	deleteRecords := make([]*dns.ResourceRecordSet, 0)
	if recs.Rrsets == nil {
		return deleteRecords, nil
	}
	for _, v := range recs.Rrsets {
		if v.Name == Record && v.Type == Type {
			deleteRecords = append(deleteRecords, v)
		}
	}
	return deleteRecords, nil
}

func (s *GoogleProvider) Present(Domain, Record, Type, Value string, TTL int) (*RecordChanges, error) {
	zoneName := s.getZoneName(Domain)
	if zoneName == "" {
		return nil, errors.New("zone name not found")
	}
	if Record[len(Record)-1] != '.' {
		Record = Record + string('.')
	}
	if Type == "CNAME" {
		if Value[len(Value)-1] != '.' {
			Value = Value + string('.')
		}
	}
	rec := dns.ResourceRecordSet{
		Name:    Record,
		Rrdatas: []string{Value},
		Type:    Type,
		Ttl:     int64(TTL),
	}
	changes := &dns.Change{
		Additions: []*dns.ResourceRecordSet{&rec},
	}
	deleteRecords, err := s.findDeleteRecords(zoneName, Record, Type)
	if err != nil {
		return nil, err
	}
	if len(deleteRecords) > 0 {
		changes.Deletions = deleteRecords
	}
	chg, err := s.client.Changes.Create(s.project, zoneName, changes).Do()
	if err != nil {
		return nil, err
	}
	for chg.Status == "pending" {
		time.Sleep(1 * time.Second)
		chg, err = s.client.Changes.Get(s.project, zoneName, chg.Id).Do()
		if err != nil {
			return nil, err
		}
	}
	return s.parseChange(chg), nil
}

func (s *GoogleProvider) Absent(Domain, Record, Type string) (*RecordChanges, error) {
	zoneName := s.getZoneName(Domain)
	if zoneName == "" {
		return nil, errors.New("zone name not found")
	}
	if Record[len(Record)-1] != '.' {
		Record = Record + string('.')
	}
	deleteRecords, err := s.findDeleteRecords(zoneName, Record, Type)
	if err != nil {
		return nil, err
	}
	changes := &dns.Change{}
	if len(deleteRecords) > 0 {
		changes.Deletions = deleteRecords
		chg, err := s.client.Changes.Create(s.project, zoneName, changes).Do()
		if err != nil {
			return nil, err
		}
		return s.parseChange(chg), nil
	} else {
		return nil, errors.New("records not found")
	}
}

func (s *GoogleProvider) List(Domain string) ([]DNSRecord, error) {
	zoneName := s.getZoneName(Domain)
	if zoneName == "" {
		return nil, errors.New("zone name not found")
	}
	recs, err := s.client.ResourceRecordSets.List(s.project, zoneName).Do()
	if err != nil {
		return nil, err
	}
	found := make([]DNSRecord, 0)
	for _, r := range recs.Rrsets {
		if r.Type == "TXT" || r.Type == "A" || r.Type == "AAAA" || r.Type == "CNAME" {
			found = append(found, DNSRecord{
				Name:  r.Name,
				Type:  r.Type,
				TTL:   int(r.Ttl),
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
