package dnscli

type DNSRecord struct {
	Name  string
	Type  string
	TTL   int64
	Datas []string
}

type DNSProvider interface {
	List(domain string) ([]DNSRecord, error)
	SetA(Record, Value string, TLL int64) error
	SetAAAA(Record, Value string, TLL int64) error
	SetCNAME(Record, Value string, TLL int64) error
	SetTXT(Record, Value string, TLL int64) error
}
