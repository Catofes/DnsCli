package dnscli

import (
	"flag"
	"log"
	"os"
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
	for k, v := range s.Config.Domains {
		domainName := k
		providerName := v
		if providerInfo, ok := s.Config.Providers[providerName]; ok {
			if providerType, ok := providerInfo["Type"]; ok {
				switch providerType {
				case "GoogleCloud":
					provider := NewGoogleProvider(providerInfo)
					s.dnsProviders[domainName] = provider
				}
			}
		}
	}
	return s
}

func Do() {
	versionFlag := flag.Bool("v", false, "Show version.")
	configPathFlag := flag.String("c","","Config path.")
	flag.Parse()
	if *versionFlag {
		log.Printf("Git commit: %s .", _version_)
		os.Exit(0)
	}
	cli := (&Cli{}).Init(*configPathFlag).Load()
	domain := flag.NewFlagSet("domain", flag.ExitOnError)
	if domain.Parsed() {
		log.Print(domain.Args())
		if domain.Arg(1) == "list" {
			log.Print("All Domains:")
			for k := range cli.dnsProviders {
				log.Print(k)
			}
		}
	}
}
