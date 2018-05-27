package dnscli

import (
	"os"
	"io/ioutil"
	"log"
	"encoding/json"
)

type Config struct {
	Providers map[string]map[string]string
	Domains   map[string]string
}

func (s *Config) Load(path string) *Config {
	if path == "" {
		path = os.Getenv("DnsCliConfig")
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Load config error, %s", err)
	}
	err = json.Unmarshal(data, s)
	if err != nil {
		log.Fatalf("Parse config error, %s", err)
	}
	return s
}
