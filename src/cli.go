package dnscli

import (
	"os"
	"fmt"
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
			fmt.Print("List All Domains:")
			for k := range cli.dnsProviders {
				fmt.Print(k)
			}
		}
	}
}
