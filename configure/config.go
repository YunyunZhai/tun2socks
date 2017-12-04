package configure

import (
	"errors"
	"gopkg.in/gcfg.v1"
	"log"
	"net/url"
)

const (
	DnsDefaultPort         = 53
	DnsDefaultTtl          = 600
	DnsDefaultPacketSize   = 4096
	DnsDefaultReadTimeout  = 5
	DnsDefaultWriteTimeout = 5
	DnsIPPoolMaxSpace      = 0x3ffff // 4*65535
)

type GeneralConfig struct {
	Network      string // tun network
	NetstackAddr string `gcfg:"netstack-addr"`
	NetstackPort uint16 `gcfg:"netstack-port"`
	Mtu          uint32
}

type DnsConfig struct {
	DnsMode         string `gcfg:"dns-mode"`
	Proxy           string
	DnsPort         uint16 `gcfg:"dns-port"`
	DnsTtl          uint   `gcfg:"dns-ttl"`
	DnsPacketSize   uint16 `gcfg:"dns-packet-size"`
	DnsReadTimeout  uint   `gcfg:"dns-read-timeout"`
	DnsWriteTimeout uint   `gcfg:"dns-write-timeout"`
	Nameserver      []string // backend dns
}

type RouteConfig struct {
	V []string
}

type PatternConfig struct {
	Proxy  string
	Scheme string
	V      []string
}

type RuleConfig struct {
	Pattern []string
	Final   string
}

type ProxyConfig struct {
	Url     string
	Default bool
}

type AppConfig struct {
	General GeneralConfig
	Dns     DnsConfig
	Route   RouteConfig
	Proxy   map[string]*ProxyConfig
	Pattern map[string]*PatternConfig
	Rule    RuleConfig
}

func (cfg *AppConfig) check() error {
	// TODO
	return nil
}

func Parse(filename string) (*AppConfig, error) {
	cfg := new(AppConfig)

	// set default value
	cfg.General.Network = "10.192.0.1/16"
	cfg.General.NetstackAddr = "10.192.0.2"
	cfg.General.NetstackPort = 7777
	cfg.General.Mtu = 1500

	cfg.Dns.DnsMode = "fake"
	cfg.Dns.Proxy = ""
	cfg.Dns.DnsPort = DnsDefaultPort
	cfg.Dns.DnsTtl = DnsDefaultTtl
	cfg.Dns.DnsPacketSize = DnsDefaultPacketSize
	cfg.Dns.DnsReadTimeout = DnsDefaultReadTimeout
	cfg.Dns.DnsWriteTimeout = DnsDefaultWriteTimeout

	// decode config value
	err := gcfg.ReadFileInto(cfg, filename)
	if err != nil {
		return nil, err
	}

	// set backend dns default value
	if len(cfg.Dns.Nameserver) == 0 {
		cfg.Dns.Nameserver = append(cfg.Dns.Nameserver, "114.114.114.114:53")
		cfg.Dns.Nameserver = append(cfg.Dns.Nameserver, "223.5.5.5:53")
	}

	err = cfg.check()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// Get default proxy, eg: socks5://127.0.0.1:1080, return 127.0.0.1:1080
func (cfg *AppConfig) DefaultPorxy() (string, error) {
	for _, proxyConfig := range cfg.Proxy {
		if proxyConfig.Default {
			url, err := url.Parse(proxyConfig.Url)
			if err != nil {
				log.Println("Parse url failed", err)
				break
			}
			return url.Host, nil
		}
	}

	return "", errors.New("404")
}

func (cfg *AppConfig) UdpProxy() (string, error) {
	proxyConfig := cfg.Proxy[cfg.Dns.Proxy]
	if proxyConfig == nil {
		for _, pc := range cfg.Proxy {
			if pc.Default {
				proxyConfig = pc
			}
		}
	}
	if proxyConfig != nil {
		url, err := url.Parse(proxyConfig.Url)
		if err != nil {
			log.Println("Parse url failed", err)
			return "", err
		}
		return url.Host, nil
	}

	return "", errors.New("404")
}