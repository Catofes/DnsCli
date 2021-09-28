package dnscli

import (
	"errors"
	"log"
	"strings"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	dns "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/model"
	region "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/region"
)

type HuaweiProvider struct {
	Region string
	AK     string
	SK     string
	client *dns.DnsClient
}

func (s *HuaweiProvider) List(Domain string) ([]DNSRecord, error) {
	zone, err := s.client.ListPublicZones(&model.ListPublicZonesRequest{
		Name: &Domain,
	})
	if err != nil {
		return nil, err
	}
	if len(*zone.Zones) == 0 {
		return nil, errors.New("zone not found")
	}
	records, err := s.client.ListRecordSetsByZone(&model.ListRecordSetsByZoneRequest{
		ZoneId: *(*zone.Zones)[0].Id,
	})
	if err != nil {
		return nil, err
	}
	result := make([]DNSRecord, 0)
	for _, v := range *records.Recordsets {
		record := DNSRecord{
			Name:  *v.Name,
			TTL:   int(*v.Ttl),
			Type:  *v.Type,
			Datas: *v.Records,
		}
		result = append(result, record)
	}
	return result, nil
}

func (s *HuaweiProvider) Present(Domain, Record, Type, Value string, TTL int) (*RecordChanges, error) {
	zone, err := s.client.ListPublicZones(&model.ListPublicZonesRequest{
		Name: &Domain,
	})
	if err != nil {
		return nil, err
	}
	if len(*zone.Zones) == 0 {
		return nil, errors.New("zone not found")
	}
	records, err := s.client.ListRecordSetsByZone(&model.ListRecordSetsByZoneRequest{
		ZoneId: *(*zone.Zones)[0].Id,
		Name:   &Record,
	})
	if err != nil {
		return nil, err
	}
	recordChanges := &RecordChanges{}
	for _, v := range *records.Recordsets {
		if strings.Compare(Record, strings.Trim(*v.Name, ".")) == 0 && strings.Compare(Type, *v.Type) == 0 {
			_, err := s.client.DeleteRecordSet(&model.DeleteRecordSetRequest{
				ZoneId:      *v.ZoneId,
				RecordsetId: *v.Id,
			})
			log.Printf("delete")
			if err != nil {
				return recordChanges, err
			}
			if recordChanges.Delete == nil {
				recordChanges.Delete = make([]DNSRecord, 0)
			}
			recordChanges.Delete = append(recordChanges.Delete, DNSRecord{
				*v.Name, *v.Type, int(*v.Ttl), *v.Records,
			})
		}
	}
	ttl := int32(TTL)
	_, err = s.client.CreateRecordSet(&model.CreateRecordSetRequest{
		ZoneId: *(*zone.Zones)[0].Id,
		Body: &model.CreateRecordSetReq{
			Name:    Record,
			Type:    Type,
			Ttl:     &ttl,
			Records: []string{Value},
		}})
	if err != nil {
		return recordChanges, err
	}
	recordChanges.Add = []DNSRecord{DNSRecord{
		Record, Type, TTL, []string{Value},
	}}
	return recordChanges, nil
}

func (s *HuaweiProvider) Absent(Domain, Record, Type string) (*RecordChanges, error) {
	zone, err := s.client.ListPublicZones(&model.ListPublicZonesRequest{
		Name: &Domain,
	})
	if err != nil {
		return nil, err
	}
	if len(*zone.Zones) == 0 {
		return nil, errors.New("zone not found")
	}
	records, err := s.client.ListRecordSetsByZone(&model.ListRecordSetsByZoneRequest{
		ZoneId: *(*zone.Zones)[0].Id,
		Name:   &Record,
	})
	if err != nil {
		return nil, err
	}
	recordChanges := &RecordChanges{}
	for _, v := range *records.Recordsets {
		if strings.Compare(Record, strings.Trim(*v.Name, ".")) == 0 && strings.Compare(Type, *v.Type) == 0 {
			_, err := s.client.DeleteRecordSet(&model.DeleteRecordSetRequest{
				ZoneId:      *v.ZoneId,
				RecordsetId: *v.Id,
			})
			if err != nil {
				return recordChanges, err
			}
			if recordChanges.Delete == nil {
				recordChanges.Delete = make([]DNSRecord, 0)
			}
			recordChanges.Delete = append(recordChanges.Delete, DNSRecord{
				*v.Name, *v.Type, int(*v.Ttl), *v.Records,
			})
		}
	}
	return recordChanges, nil
}

func NewHuaweiProvider(info map[string]string) DNSProvider {
	provider := HuaweiProvider{}
	if v, ok := info["Endpoint"]; ok {
		provider.Region = v
	} else {
		provider.Region = "cn-north-4"
	}
	if v, ok := info["AK"]; ok {
		provider.AK = v
	} else {
		log.Fatal("Huawei: missing AK")
	}
	if v, ok := info["SK"]; ok {
		provider.SK = v
	} else {
		log.Fatal("Huawei: missing SK")
	}
	auth := basic.NewCredentialsBuilder().
		WithAk(provider.AK).WithSk(provider.SK).Build()
	client := dns.NewDnsClient(
		dns.DnsClientBuilder().
			WithRegion(region.ValueOf(provider.Region)).
			WithCredential(auth).Build())
	provider.client = client
	return &provider
}
