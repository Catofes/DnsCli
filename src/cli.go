package dnscli

import (
	"fmt"
	"net"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

var _version_ string

type Cli struct {
	Config
	dnsProviders map[string]DNSProvider
}

func (s *Cli) Init(path string) *Cli {
	s.Config = *(s.Config.Load(path))
	s.dnsProviders = make(map[string]DNSProvider)
	return s
}

func (s *Cli) Load() *Cli {
	tmp := make(map[string]DNSProvider)
	for k, v := range s.Config.Providers {
		if providerType, ok := v["Type"]; ok {
			switch providerType {
			case "GoogleCloud":
				provider := NewGoogleProvider(v)
				tmp[k] = provider
			case "Cloudflare":
				provider := NewCloudflareProvider(v)
				tmp[k] = provider
			}
		}
	}
	for k, v := range s.Config.Domains {
		domainName := k
		providerName := v
		if v, ok := tmp[providerName]; ok {
			s.dnsProviders[domainName] = v
		}
	}
	return s
}

func (s *Cli) PrintDomains() {
	fmt.Println("List All Domains:")
	domains := make([]string, 0)
	for k := range s.dnsProviders {
		domains = append(domains, k)
	}
	sort.Strings(domains)
	for _, v := range domains {
		fmt.Println(v)
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
		value := strings.Join(v.Datas, " ")
		if len(value) > 48 {
			value = value[:48] + string("...")
		}
		table.Append([]string{v.Name, value, v.Type, strconv.Itoa(v.TTL)})
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

func (s *Cli) ListDomain(args []string) {
	if len(args) <= 0 {
		fmt.Println("Empty domain.")
		os.Exit(1)
	}
	domain := args[0]
	typeFilters := []string{"TXT", "CNAME", "A", "AAAA"}
	if len(args) >= 2 {
		typeFilters = args[1:]
	}
	if v, ok := s.dnsProviders[domain]; ok {
		records, err := v.List(domain)
		if err != nil {
			fmt.Printf("List domain err, %s.\n", err.Error())
			os.Exit(1)
		}
		records = choose(records, func(i int) bool {
			for _, t := range typeFilters {
				if records[i].Type == t {
					return true
				}
			}
			return false
		}).([]DNSRecord)
		printRecords(records, domain)
		os.Exit(0)
	} else {
		fmt.Println("Unknown domain.")
		os.Exit(1)
	}
}

func (s *Cli) findDomain(record string) string {
	tmpDomain := make([]string, 0)
	for k := range s.dnsProviders {
		if strings.Contains(record, k) {
			tmpDomain = append(tmpDomain, k)
		}
	}
	if len(tmpDomain) <= 0 {
		return ""
	}
	sort.Slice(tmpDomain, func(i, j int) bool { return len(tmpDomain[i]) > len(tmpDomain[j]) })
	return tmpDomain[0]
}

func (s *Cli) ShowRecord(args []string) {
	if len(args) <= 0 {
		fmt.Println("Empty record.")
		os.Exit(1)
	}
	record := args[0]
	domain := s.findDomain(record)
	if domain == "" {
		fmt.Println("Domain not found")
		os.Exit(1)
	}
	records, err := s.dnsProviders[domain].List(domain)
	if err != nil {
		fmt.Printf("List domain err, %s.\n", err.Error())
		os.Exit(1)
	}
	typeFilter := ""
	if len(args) >= 2 {
		typeFilter = args[1]
	}
	if typeFilter != "" {
		records = choose(records, func(i int) bool { return records[i].Type == typeFilter }).([]DNSRecord)
	}
	records = choose(records, func(i int) bool { return strings.Contains(records[i].Name, record) }).([]DNSRecord)
	printRecords(records, domain)
}

func (s *Cli) SetRecord(args []string) {
	if len(args) <= 1 {
		fmt.Println("Please input record value [type] [ttl].")
		os.Exit(1)
	}
	record := args[0]
	domain := s.findDomain(record)
	if domain == "" {
		fmt.Println("Domain not found")
		os.Exit(1)
	}
	provider := s.dnsProviders[domain]

	recordValue := args[1]
	recordType := ""
	recordTTL := 300
	if len(args) >= 3 {
		recordType = strings.ToUpper(args[2])
	}
	if len(args) >= 4 {
		recordTTL, _ = strconv.Atoi(args[3])
	}
	if recordType == "" {
		ip := net.ParseIP(recordValue)
		if ip != nil {
			if ip.To4() != nil {
				recordType = "A"
			} else {
				recordType = "AAAA"
			}
		} else {
			isDomain, _ := regexp.Match("^(([a-zA-Z]{1})|([a-zA-Z]{1}[a-zA-Z]{1})|([a-zA-Z]{1}[0-9]{1})|([0-9]{1}[a-zA-Z]{1})|([a-zA-Z0-9][a-zA-Z0-9-_]{1,61}[a-zA-Z0-9]))\\.([a-zA-Z]{2,6}|[a-zA-Z0-9-]{2,30}\\.[a-zA-Z]{2,3})$",
				[]byte(recordValue))
			if isDomain {
				recordType = "CNAME"
			} else {
				recordType = "TXT"
			}
		}
	}
	changes, err := provider.Present(domain, record, recordType, recordValue, recordTTL)
	if err != nil {
		fmt.Printf("Set record error, %s.\n", err.Error())
		os.Exit(1)
	} else {
		fmt.Printf("Set success.\n")
		printChanges(*changes)
	}
}

func (s *Cli) DeleteRecord(args []string) {
	if len(args) <= 1 {
		fmt.Println("Please input record type.")
		os.Exit(1)
	}
	record := args[0]
	recordType := args[1]
	domain := s.findDomain(record)
	if domain == "" {
		fmt.Println("Domain not found")
		os.Exit(1)
	}
	provider := s.dnsProviders[domain]
	changes, err := provider.Absent(domain, record, recordType)
	if err != nil {
		fmt.Printf("Delete record error, %s.\n", err.Error())
		os.Exit(1)
	} else {
		fmt.Printf("Delete success.\n")
		printChanges(*changes)
	}
}

func parseOperation(args []string) []string {
	for len(args) > 0 {
		if args[0][0] == '-' {
			args = args[2:]
		} else {
			break
		}
	}
	return args
}

func Do(configPath string) {
	args := os.Args[1:]
	args = parseOperation(args)
	cli := (&Cli{}).Init(configPath).Load()

	if len(args) > 0 {
		switch args[0] {
		case "domain":
			cli.PrintDomains()
		case "d":
			cli.PrintDomains()
		case "list":
			cli.ListDomain(args[1:])
		case "l":
			cli.ListDomain(args[1:])
		case "get":
			cli.ShowRecord(args[1:])
		case "g":
			cli.ShowRecord(args[1:])
		case "set":
			cli.SetRecord(args[1:])
		case "s":
			cli.SetRecord(args[1:])
		case "add":
			cli.SetRecord(args[1:])
		case "a":
			cli.SetRecord(args[1:])
		case "delete":
			cli.DeleteRecord(args[1:])
		case "del":
			cli.DeleteRecord(args[1:])
		default:
			fmt.Printf("Command not found. \n Please input domain, list, get, set or delete.")
		}
	}
}
