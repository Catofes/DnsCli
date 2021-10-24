package dnscli

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/miekg/dns"
	"github.com/olekukonko/tablewriter"
)

func fqdn(input string) string {
	return dns.Fqdn(input)
}

func defqdn(input string) string {
	if input[len(input)-1] == '.' {
		return input[:len(input)-1]
	} else {
		return input
	}
}

func compareRecord(l, r DNSRecord) bool {
	reserveStringSplice := func(s []string) []string {
		for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
			s[i], s[j] = s[j], s[i]
		}
		return s
	}
	reserveDomain := func(s string) string {
		t := strings.Split(s, ".")
		reserveStringSplice(t)
		return strings.Join(t, ".")
	}
	return strings.Compare(reserveDomain(l.Name), reserveDomain(r.Name)) <= 0
}

func sortRecord(records []DNSRecord) {
	sort.Slice(records, func(i, j int) bool {
		return compareRecord(records[i], records[j])
	})
}

func printRecords(records []DNSRecord, domain string) {
	sortRecord(records)
	fmt.Printf("Records in %s\n", domain)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Value", "Type", "TTL"})
	table.SetColumnAlignment([]int{tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_LEFT, tablewriter.ALIGN_CENTER})
	table.SetAutoWrapText(false)
	for _, v := range records {
		//value := strings.Join(v.Datas, " ")
		value := v.Datas[0]
		if len(value) > 48 {
			value = value[:48] + string("...")
		}
		table.Append([]string{v.Name, value, v.Type, strconv.Itoa(v.TTL)})
		for _, v := range v.Datas[1:] {
			table.Append([]string{"", v, "", ""})
		}
	}
	table.Render()
}

func printChanges(change RecordChanges) {
	sortRecord(change.Add)
	sortRecord(change.Delete)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Operate", "Name", "Value", "Type", "TTL"})
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_LEFT, tablewriter.ALIGN_CENTER})
	table.SetAutoWrapText(false)
	for _, v := range change.Add {
		value := strings.Join(v.Datas, " ")
		if len(value) > 48 {
			value = value[:48] + string("...")
		}
		table.Append([]string{"ADD", v.Name, value, v.Type, strconv.Itoa(v.TTL)})
	}
	for _, v := range change.Delete {
		value := strings.Join(v.Datas, " ")
		if len(value) > 48 {
			value = value[:48] + string("...")
		}
		table.Append([]string{"DEL", v.Name, value, v.Type, strconv.Itoa(v.TTL)})
	}
	table.Render()
}

func choose(slice interface{}, filter func(i int) bool) interface{} {
	if t := reflect.TypeOf(slice); t.Kind() != reflect.Slice && t.Kind() != reflect.Array {
		return nil
	}
	sliceType := reflect.TypeOf(slice)
	result := reflect.MakeSlice(sliceType, 0, reflect.ValueOf(slice).Len())
	for i := 0; i < reflect.ValueOf(slice).Len(); i++ {
		if filter(i) {
			result = reflect.Append(result, reflect.ValueOf(slice).Index(i))
		}
	}
	ptr := reflect.New(sliceType)
	ptr.Elem().Set(result)
	return ptr.Elem().Interface()
}

func RR2DNSRecord(rr []dns.RR) []DNSRecord {
	result := make([]DNSRecord, 0)
	for _, a := range rr {
		switch v := a.(type) {
		case *dns.A:
			result = append(result, DNSRecord{
				Name:  v.Hdr.Name,
				TTL:   int(v.Hdr.Ttl),
				Type:  "A",
				Datas: []string{v.A.String()},
			})
		case *dns.AAAA:
			result = append(result, DNSRecord{
				Name:  v.Hdr.Name,
				TTL:   int(v.Hdr.Ttl),
				Type:  "AAAA",
				Datas: []string{v.AAAA.String()},
			})
		case *dns.CNAME:
			result = append(result, DNSRecord{
				Name:  v.Hdr.Name,
				TTL:   int(v.Hdr.Ttl),
				Type:  "CNAME",
				Datas: []string{v.Target},
			})
		case *dns.TXT:
			result = append(result, DNSRecord{
				Name:  v.Hdr.Name,
				TTL:   int(v.Hdr.Ttl),
				Type:  "TXT",
				Datas: v.Txt,
			})
		case *dns.NS:
			result = append(result, DNSRecord{
				Name:  v.Hdr.Name,
				TTL:   int(v.Hdr.Ttl),
				Type:  "NS",
				Datas: []string{v.Ns},
			})
		case *dns.PTR:
			result = append(result, DNSRecord{
				Name:  v.Hdr.Name,
				TTL:   int(v.Hdr.Ttl),
				Type:  "PTR",
				Datas: []string{v.Ptr},
			})
		case *dns.MX:
			result = append(result, DNSRecord{
				Name:  v.Hdr.Name,
				TTL:   int(v.Hdr.Ttl),
				Type:  "MX",
				Datas: []string{fmt.Sprintf("%d %s", v.Preference, v.Mx)},
			})
		case *dns.SRV:
			result = append(result, DNSRecord{
				Name:  v.Hdr.Name,
				TTL:   int(v.Hdr.Ttl),
				Type:  "SRV",
				Datas: []string{fmt.Sprintf("%d %d %s:%d", v.Priority, v.Weight, v.Target, v.Port)},
			})
		case *dns.SOA:
			result = append(result, DNSRecord{
				Name:  v.Hdr.Name,
				TTL:   int(v.Hdr.Ttl),
				Type:  "SOA",
				Datas: []string{fmt.Sprintf("%s %s %d %d %d %d %d", v.Ns, v.Mbox, v.Serial, v.Refresh, v.Retry, v.Expire, v.Minttl)},
			})
		}
	}
	return result
}
