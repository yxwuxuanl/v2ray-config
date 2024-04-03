package main

type InboundConfig struct {
	Listen   string `json:"listen"`
	Protocol string `json:"protocol"`
	Settings any    `json:"settings"`
	Port     string `json:"port"`
}

type TCPSettings struct {
	Header map[string]string `json:"header"`
}

type StreamSettings struct {
	TCPSettings TCPSettings `json:"tcpSettings"`
	Security    string      `json:"security"`
	Network     string      `json:"network"`
}

type Shadowsocks struct {
	Email    string `json:"email"`
	Port     int    `json:"port"`
	Level    int    `json:"level"`
	Method   string `json:"method"`
	Ota      bool   `json:"ota"`
	Password string `json:"password"`
	Address  string `json:"address"`
}

type VmessUser struct {
	AlterId  int    `json:"alterId"`
	Level    int    `json:"level"`
	Security string `json:"security"`
	Id       string `json:"id"`
}

type Vmess struct {
	Address string      `json:"address"`
	Port    int         `json:"port"`
	Users   []VmessUser `json:"users"`
}

type OutboundSettings struct {
	VNext   []Vmess       `json:"vnext,omitempty"`
	Servers []Shadowsocks `json:"servers,omitempty"`
}

type OutboundConfig struct {
	Protocol       string           `json:"protocol"`
	Tag            string           `json:"tag"`
	StreamSettings StreamSettings   `json:"streamSettings"`
	Settings       OutboundSettings `json:"settings"`
}

type LogConfig struct {
	Loglevel string `json:"loglevel"`
}

type V2rayConfig struct {
	Log       LogConfig        `json:"log"`
	Inbounds  []InboundConfig  `json:"inbounds"`
	Outbounds []OutboundConfig `json:"outbounds"`
}
