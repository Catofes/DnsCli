## DnsCli
![](https://travis-ci.org/Catofes/CertDistribution.svg?branch=master)

A little cli to change dns in `Google Cloud DNS` or `Cloudflare`. Powered by golang.

#### Install

From source
```
go get -u github.com/Catofes/DnsCli/cmd/dns
```

#### Build

You need to install `go dep` before build.

```
git clone github.com/Catofes/DnsCli
cd DnsCli
dep ensure
make
```

#### Config
You can use environments `DNSCLI_CONFIG` or parameters to set config path.

Example config json file:
```
{
  "Providers": {
    "GoogleCloud": {
      "Type": "GoogleCloud",
      "Project": "PROJECTNAME",
      "SaFile": "OAUTH JSON FILE"
    },
    "Cloudflare": {
      "Type": "Cloudflare",
      "Email": "AAA@example.com",
      "Key": "abc2123"
    }
  },
  "Domains": {
    "example.com": "GoogleCloud",
    "test.moe": "GoogleCloud",
    "good.wf": "GoogleCloud",
    "big.app": "Cloudflare",
    "ssss.xyz": "Cloudflare",
    "le.com": "Cloudflare"
  }
}
```

#### Usage

```
dns domain
dns d
dns list example.com
dns l example.com A
dns get test.example.com
dns g sub.example.com AAAA
dns set test.test.moe 127.0.0.1
dns s test.test.moe 2001::da8
dns s test.big.app j.test.com
dns delete test.test.moe A
dns del test.test.moe AAAA
```