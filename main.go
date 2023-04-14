package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"
)

type SubscribeConfig struct {
	Ps   string `json:"ps"`
	Port string `json:"port"`
	Id   string `json:"id"`
	Aid  int    `json:"aid"`
	Net  string `json:"net"`
	Type string `json:"type"`
	Tls  string `json:"tls"`
	Add  string `json:"add"`
}

type TemplateData struct {
	SocksPort string
	HttpPort  string
	*SubscribeConfig
}

var (
	subscribe    = flag.String("subscribe", "", "")
	httpPort     = flag.String("http-port", "1087", "")
	socksPort    = flag.String("socks-port", "1080", "")
	configOutput = flag.String("config-output", "./v2ray.json", "")
)

func main() {
	flag.Parse()

	if *subscribe == "" {
		log.Fatal("flag --subscribe is required")
	}

	res, err := http.Get(*subscribe)
	if err != nil {
		log.Fatalf("request subscribe url error: %s", err)
	}

	if res.StatusCode != http.StatusOK {
		log.Fatalf("subscribe returned bad status: %s", res.Status)
	}

	rawlinks := new(bytes.Buffer)
	_, err = rawlinks.ReadFrom(base64.NewDecoder(base64.StdEncoding, res.Body))

	if err != nil {
		log.Fatalf("decode base64 error: %s", err)
	}

	var subConfs []*SubscribeConfig

	rl := bufio.NewReader(rawlinks)

	for {
		line, _, err := rl.ReadLine()
		if err == io.EOF {
			break
		}

		if !bytes.HasPrefix(line, []byte("vmess://")) {
			continue
		}

		conf, err := base64.StdEncoding.DecodeString(string(line[8:]))
		if err != nil {
			log.Printf("decode vmess link error: %s, link=%s", err, line)
			continue
		}

		subConf := new(SubscribeConfig)

		if err = json.Unmarshal(conf, subConf); err != nil {
			log.Printf("unmarshal vmess conf error: %s, conf=%s", err, conf)
			continue
		} else {
			subConfs = append(subConfs, subConf)
		}
	}

	if len(subConfs) == 0 {
		log.Fatal("no available config")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	d := net.Dialer{}

	var fasterConf *SubscribeConfig

	now := time.Now()

	for _, conf := range subConfs {
		go func(conf *SubscribeConfig) {
			log.Printf("dial %s...", conf.Add)

			_, err := d.DialContext(ctx, "tcp", conf.Add+":"+conf.Port)
			if err != nil {
				if !strings.Contains(err.Error(), "canceled") {
					log.Printf("dial %s error: %s", conf.Add, err)
				}
				return
			}

			if fasterConf == nil {
				cancel()
				fasterConf = conf
				log.Printf("use server %s latency=%s", conf.Add, time.Since(now))
			}
		}(conf)
	}

	<-ctx.Done()

	if fasterConf == nil {
		fasterConf = subConfs[0]
	}

	data := &TemplateData{
		SocksPort:       *socksPort,
		HttpPort:        *httpPort,
		SubscribeConfig: fasterConf,
	}

	tpl, err := template.New("v2ray-config").Parse(confTpl)
	if err != nil {
		log.Fatal(err.Error())
	}

	confBuf := new(bytes.Buffer)
	if err := tpl.Execute(confBuf, data); err != nil {
		log.Fatal(err.Error())
	}

	err = os.WriteFile(*configOutput, confBuf.Bytes(), 0777)

	if err != nil {
		log.Fatalf("write config file error: %s", err)
	}

	log.Printf("v2ray config write to %s", *configOutput)
}
