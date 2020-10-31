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
	Present(Domain, record, recordType, recordValue string, recordTTL int) (*RecordChanges, error)
	Absent(Domain, record, recordType string) (*RecordChanges, error)
}
