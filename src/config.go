package dnscli

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/miekg/dns"
)

var (
	tsigAlg = map[string]string{
		"hmac-sha1":   "hmac-sha1.",
		"hmac-sha224": "hmac-sha224.",
		"hmac-sha256": "hmac-sha256.",
		"hmac-sha384": "hmac-sha384.",
		"hmac-sha512": "hmac-sha512.",
	}
)

type Config struct {
	Providers map[string]map[string]string
	Domains   map[string]string
	Tsig      string
	Listen    string
}

func (s *Config) Load(path string) *Config {
	if path == "" {
		path = os.Getenv("DNSCLI_CONFIG")
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Load config error, %s", err)
	}
	s.Listen = "[::]:53"
	err = json.Unmarshal(data, s)
	if err != nil {
		log.Fatalf("Parse config error, %s", err)
	}
	return s
}

func (s *Config) parseTsig() (alg, name, secret string, err error) {
	t := strings.Split(s.Tsig, ":")
	if len(t) == 3 {
		if _, ok := tsigAlg[t[0]]; ok {
			alg = tsigAlg[t[0]]
		} else {
			return "", "", "", errors.New("tsig algorithm not found")
		}
		name = dns.Fqdn(t[1])
		secret = t[2]
		return
	} else if len(t) == 2 {
		alg = "hmac-sha1."
		name = dns.Fqdn(t[0])
		secret = t[1]
		return
	} else {
		return "", "", "", errors.New("tsig name not found")
	}
}
