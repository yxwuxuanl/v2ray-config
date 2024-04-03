package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"strings"
)

type Subscribe interface {
	ToOutboundConfig() OutboundConfig
	Parse(string) error
	Address() string
	Name() string
}

type VmessSubscribe struct {
	Ps   string `json:"ps"`
	Port string `json:"port"`
	Id   string `json:"id"`
	Aid  int    `json:"aid"`
	Net  string `json:"net"`
	Type string `json:"type"`
	Tls  string `json:"tls"`
	Add  string `json:"add"`
}

type ShadowsocksSubscribe struct {
	*url.URL
}

func (v *VmessSubscribe) Address() string {
	return v.Add + ":" + v.Port
}

func (v *VmessSubscribe) Name() string {
	return "vmess://" + v.Ps
}

func (v *VmessSubscribe) ToOutboundConfig() OutboundConfig {
	return OutboundConfig{
		Protocol: "vmess",
		Tag:      "proxy",
		StreamSettings: StreamSettings{
			Security: "none",
			Network:  "tcp",
			TCPSettings: TCPSettings{
				Header: map[string]string{
					"type": "none",
				},
			},
		},
		Settings: OutboundSettings{
			VNext: []Vmess{
				{
					Address: v.Add,
					Port:    toPort(v.Port),
					Users: []VmessUser{
						{
							AlterId:  v.Aid,
							Security: "auto",
							Id:       v.Id,
						},
					},
				},
			},
		},
	}
}

func (v *VmessSubscribe) Parse(u string) error {
	jsonconf, err := base64.RawStdEncoding.DecodeString(u)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonconf, v)
}

func (s *ShadowsocksSubscribe) ToOutboundConfig() OutboundConfig {
	password, _ := s.User.Password()

	return OutboundConfig{
		Tag:      "proxy",
		Protocol: "shadowsocks",
		StreamSettings: StreamSettings{
			Security: "none",
			Network:  "tcp",
			TCPSettings: TCPSettings{
				Header: map[string]string{
					"type": "none",
				},
			},
		},
		Settings: OutboundSettings{
			Servers: []Shadowsocks{
				{
					Port:     toPort(s.Port()),
					Address:  s.Hostname(),
					Method:   s.User.Username(),
					Password: password,
				},
			},
		},
	}
}

func (s *ShadowsocksSubscribe) Name() string {
	return "ss://" + s.Host
}

func (s *ShadowsocksSubscribe) Parse(u string) error {
	if i := strings.Index(u, "#"); i > -1 {
		u = u[:i]
	}

	ss, err := base64.RawStdEncoding.DecodeString(u)
	if err != nil {
		return err
	}

	s.URL, err = url.Parse("ss://" + string(ss))
	return err
}

func (s *ShadowsocksSubscribe) Address() string {
	return s.Host
}

func ParseSubscribe(u string) (Subscribe, error) {
	var subscribeConfig Subscribe

	items := strings.Split(u, "://")
	switch items[0] {
	case "ss":
		subscribeConfig = new(ShadowsocksSubscribe)
	case "vmess":
		subscribeConfig = new(VmessSubscribe)
	default:
		return nil, errors.New("unknown subscribe server protocol: " + items[0])
	}

	return subscribeConfig, subscribeConfig.Parse(items[1])
}

func toPort(v string) int {
	port, _ := strconv.Atoi(v)
	return port
}

func buildV2rayConfig(subConf Subscribe) *V2rayConfig {
	return &V2rayConfig{
		Log: LogConfig{Loglevel: "info"},
		Inbounds: []InboundConfig{
			{
				Protocol: "http",
				Port:     strconv.Itoa(*httpPort),
				Listen:   "0.0.0.0",
				Settings: map[string]any{
					"timeout": 600,
				},
			},
			{
				Protocol: "socks",
				Port:     strconv.Itoa(*socksPort),
				Listen:   "0.0.0.0",
				Settings: map[string]any{
					"udp":  false,
					"auth": "noauth",
				},
			},
		},
		Outbounds: []OutboundConfig{
			subConf.ToOutboundConfig(),
		},
	}
}
