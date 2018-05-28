package dnscli

type DNSRecord struct {
	Name  string
	Type  string
	TTL   int
	Datas []string
}

type RecordChanges struct {
	Add    []DNSRecord
	Delete []DNSRecord
}

type DNSProvider interface {
	List(Domain string) ([]DNSRecord, error)
	Present(Domain, Record, Type, Value string, TTL int) (*RecordChanges,error)
	Absent(Domain, Record, Type string) (*RecordChanges,error)
}
