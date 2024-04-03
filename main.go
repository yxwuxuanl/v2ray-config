package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"sync"
	"time"
)

var (
	subscribe      = flag.String("subscribe", "", "")
	httpPort       = flag.Int("http-port", 1087, "")
	socksPort      = flag.Int("socks-port", 1080, "")
	reloadDuration = flag.Duration("reload-duration", time.Hour, "")
	probeUrl       = flag.String("probe-url", "https://google.com", "")
)

var (
	probeClient  *http.Client
	v2rayProcess *os.Process
)

func main() {
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	proxyUrl, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", *httpPort))
	probeClient = &http.Client{
		Timeout: time.Second * 5,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		},
	}

	http.HandleFunc("/probe", func(rw http.ResponseWriter, r *http.Request) {
		if err := doProbe(1); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	})

	go http.ListenAndServe(":3000", nil)

	t := time.NewTicker(*reloadDuration)

	for {
		if err := run(); err != nil {
			if v2rayProcess == nil {
				log.Fatalln(err.Error())
			}

			slog.Error("run v2ray error", "err", err.Error())
		}

		<-t.C
	}
}

func run() error {
	subscribes, err := fetchSubscribe()
	if err != nil {
		return err
	}

	slog.Info("fetchSubscribe", "subscribes", len(subscribes))

	ch := make(chan Subscribe, len(subscribes))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	d := net.Dialer{}

	var wg sync.WaitGroup

	for _, subscribe := range subscribes {
		wg.Add(1)

		go func() {
			defer wg.Done()
			start := time.Now()

			conn, err := d.DialContext(ctx, "tcp", subscribe.Address())
			if err != nil {
				slog.Error(
					"dial server error",
					"err", err.Error(),
					"server", subscribe.Name(),
				)
				return
			}

			defer conn.Close()
			ch <- subscribe

			slog.Info(
				"dial server",
				"duration", time.Since(start).String(),
				"server", subscribe.Name(),
			)
		}()
	}

	wg.Wait()
	close(ch)

	if len(ch) == 0 {
		return errors.New("no available subscribe")
	}

	for subscribe := range ch {
		if err := runV2ray(buildV2rayConfig(subscribe)); err != nil {
			return err
		}

		start := time.Now()
		if err := doProbe(3); err != nil {
			slog.Error(
				"doProbe error",
				"err", err,
				"server", subscribe.Name(),
			)
			continue
		}

		slog.Info(
			"doProbe success",
			"duration", time.Since(start).String(),
			"server", subscribe.Name(),
		)
		break
	}

	return nil
}

func runV2ray(config *V2rayConfig) error {
	cmd := exec.Command("v2ray", "run", "-format", "json")

	cmd.Env = []string{"PATH=" + os.Getenv("PATH")}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := json.NewEncoder(stdinPipe).Encode(config); err != nil {
		return err
	}

	stdinPipe.Close()

	if v2rayProcess != nil {
		slog.Info("stop v2ray process", "pid", v2rayProcess.Pid)
		v2rayProcess.Kill()
		v2rayProcess.Wait()
		v2rayProcess = nil
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	v2rayProcess = cmd.Process
	slog.Info("start v2ray process", "pid", cmd.Process.Pid)

	return nil
}

func fetchSubscribe() ([]Subscribe, error) {
	res, err := http.Get(*subscribe)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad http status: %s", res.Status)
	}

	buf := new(bytes.Buffer)

	_, err = buf.ReadFrom(base64.NewDecoder(base64.StdEncoding, res.Body))
	if err != nil {
		return nil, err
	}

	res.Body.Close()
	scan := bufio.NewScanner(buf)

	var subscribes []Subscribe

	for scan.Scan() {
		sub, err := ParseSubscribe(scan.Text())

		if err != nil {
			slog.Error(
				"ParseSubscribe error",
				"err", err.Error(),
				"url", scan.Text(),
			)
			continue
		}

		subscribes = append(subscribes, sub)
	}

	return subscribes, nil
}

func doProbe(retry int) error {
	var err error

	for range retry {
		if _, err = probeClient.Get(*probeUrl); err != nil {
			time.Sleep(time.Second)
			continue
		}

		return nil
	}

	return err
}
