package dnscli

import (
	"flag"
	"log"
)

var version string

type Cli struct {
	Config
	dnsProviders map[string]DNSProvider
}

func (s *Cli) Init() *Cli {
	s.Config = *(s.Config.Load())
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
				case "Google":
					provider := NewGoogleProvider(providerInfo)
					s.dnsProviders[domainName] = provider
				}
			}
		}
	}
}

func Do() {
	versionFlag := flag.Bool("v", false, "Show version.")
	flag.Parse()
	if *versionFlag {
		log.Printf("Git commit: %s .", version)
	}

	domain := flag.NewFlagSet("domain", flag.ExitOnError)
	if domain.Parsed() {
		cli := (&Cli{}).Init().Load()
		if domain.Arg(1) == "list" {
			log.Print("All Domains:")
			for k := range cli.dnsProviders {
				log.Print(k)
			}
		}
	}
}
