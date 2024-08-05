// Package agent contains implementation for metric agent.
//
// The agent collects counter and gauge metrics to storage and sends their to server with interval.
package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	math "math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/kaa-it/go-devops/internal/api"
)

const (
	_requestTimeout = 5 * time.Second
)

// Agent describes metric agent.
type Agent struct {
	storage   *Storage
	config    *Config
	client    *resty.Client
	publicKey *rsa.PublicKey
}

// New creates new metric agent
//
// Takes client to connect to metric server and config with total configuration of agent.
func New(client *resty.Client, config *Config) (*Agent, error) {
	client.SetTimeout(_requestTimeout)

	var publicKey *rsa.PublicKey

	if config.Agent.PublicKeyPath != nil {
		publicKeyPEM, err := os.ReadFile(*config.Agent.PublicKeyPath)
		if err != nil {
			return nil, err
		}

		publicKeyBlock, _ := pem.Decode(publicKeyPEM)
		pub, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
		if err != nil {
			return nil, err
		}

		pubKey, ok := pub.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("wrong key format")
		}

		publicKey = pubKey
	}

	return &Agent{
		storage:   NewStorage(),
		config:    config,
		client:    client,
		publicKey: publicKey,
	}, nil
}

// Run runs metric agent and control its lifecycle.
func (a *Agent) Run() {
	log.Println("Agent started")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	ctx, cancel := context.WithCancel(context.Background())

	wg := new(sync.WaitGroup)
	wg.Add(3)

	go a.runPoller(ctx, wg)
	go a.runAdditionalPoller(ctx, wg)
	go a.runReporter(ctx, wg)

	<-c

	cancel()

	wg.Wait()

	log.Println("Agent terminated")
}

func (a *Agent) runPoller(ctx context.Context, wg *sync.WaitGroup) {
	pollTicker := time.NewTicker(a.config.Agent.PollInterval)
	defer pollTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Poller terminated")
			wg.Done()
			return
		case <-pollTicker.C:
			a.poll()
		}
	}
}

func (a *Agent) runAdditionalPoller(ctx context.Context, wg *sync.WaitGroup) {
	pollTicker := time.NewTicker(a.config.Agent.PollInterval)
	defer pollTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Additional poller terminated")
			wg.Done()
			return
		case <-pollTicker.C:
			a.additionalPoll()
		}
	}
}

func (a *Agent) runReporter(ctx context.Context, wg *sync.WaitGroup) {
	reportTicker := time.NewTicker(a.config.Agent.ReportInterval)
	defer reportTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Reporter terminated")
			wg.Done()
			return
		case <-reportTicker.C:
			a.report()
		}
	}
}

func (a *Agent) poll() {
	stats := &runtime.MemStats{}

	runtime.ReadMemStats(stats)

	a.storage.UpdateGauge("Alloc", float64(stats.Alloc))
	a.storage.UpdateGauge("BuckHashSys", float64(stats.BuckHashSys))
	a.storage.UpdateGauge("Frees", float64(stats.Frees))
	a.storage.UpdateGauge("GCCPUFraction", stats.GCCPUFraction)
	a.storage.UpdateGauge("GCSys", float64(stats.GCSys))
	a.storage.UpdateGauge("HeapAlloc", float64(stats.HeapAlloc))
	a.storage.UpdateGauge("HeapIdle", float64(stats.HeapIdle))
	a.storage.UpdateGauge("HeapInuse", float64(stats.HeapInuse))
	a.storage.UpdateGauge("HeapObjects", float64(stats.HeapObjects))
	a.storage.UpdateGauge("HeapReleased", float64(stats.HeapReleased))
	a.storage.UpdateGauge("HeapSys", float64(stats.HeapSys))
	a.storage.UpdateGauge("LastGC", float64(stats.LastGC))
	a.storage.UpdateGauge("Lookups", float64(stats.Lookups))
	a.storage.UpdateGauge("MCacheSys", float64(stats.MCacheSys))
	a.storage.UpdateGauge("MCacheInuse", float64(stats.MCacheInuse))
	a.storage.UpdateGauge("MSpanInuse", float64(stats.MSpanInuse))
	a.storage.UpdateGauge("MSpanSys", float64(stats.MSpanSys))
	a.storage.UpdateGauge("Mallocs", float64(stats.Mallocs))
	a.storage.UpdateGauge("NextGC", float64(stats.NextGC))
	a.storage.UpdateGauge("NumForcedGC", float64(stats.NumForcedGC))
	a.storage.UpdateGauge("NumGC", float64(stats.NumGC))
	a.storage.UpdateGauge("OtherSys", float64(stats.OtherSys))
	a.storage.UpdateGauge("PauseTotalNs", float64(stats.PauseTotalNs))
	a.storage.UpdateGauge("StackInuse", float64(stats.StackInuse))
	a.storage.UpdateGauge("StackSys", float64(stats.StackSys))
	a.storage.UpdateGauge("Sys", float64(stats.Sys))
	a.storage.UpdateGauge("TotalAlloc", float64(stats.TotalAlloc))

	a.storage.UpdateCounter("PollCount", 1)
	a.storage.UpdateGauge("RandomValue", math.Float64())

	log.Println("Poll done")
}

func (a *Agent) additionalPoll() {
	v, _ := mem.VirtualMemory()
	cpu, _ := cpu.Percent(0, false)

	a.storage.UpdateGauge("TotalMemory", float64(v.Total))
	a.storage.UpdateGauge("FreeMemory", float64(v.Free))
	a.storage.UpdateGauge("CPUutilization1", cpu[0])

	log.Println("Additional poll done")
}

func (a *Agent) report() {
	var metrics []api.Metrics
	a.storage.ForEachGauge(func(key string, value float64) {
		metrics = a.applyGauge(key, value, metrics)
	})

	a.storage.ForEachCounter(func(key string, value int64) {
		metrics = a.applyCounter(key, value, metrics)

		// Subtract sent value to take into account
		// possible counter updates after sending
		a.storage.UpdateCounter(key, -value)
	})

	if err := a.sendMetrics(metrics); err != nil {
		log.Println(err)
		return
	}

	log.Println("Report done")
}

func (a *Agent) sendMetrics(metrics []api.Metrics) error {
	req := a.client.R()
	req.Method = http.MethodPost

	url := fmt.Sprintf("http://%s/updates/", a.config.Server.Address)

	req.URL = url

	if a.publicKey != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("Content-Encoding", "gzip")

	buf := bytes.NewBuffer(nil)
	zw := gzip.NewWriter(buf)

	enc := json.NewEncoder(zw)
	if err := enc.Encode(metrics); err != nil {
		return fmt.Errorf("failed to encode metric for %s: %w", url, err)
	}

	if err := zw.Close(); err != nil {
		return fmt.Errorf("failed to close gzip writer: %w", err)
	}

	var body []byte

	if a.publicKey == nil {
		body = buf.Bytes()
	} else {
		var err error
		body, err = a.encrypt(buf)
		if err != nil {
			return fmt.Errorf("failed to encrypt message: %w", err)
		}
	}

	req.SetBody(body)

	if len(a.config.Agent.Key) > 0 {
		req.Header.Set("Hash", a.calculateHash(body))
	}

	resp, err := req.Send()
	if err != nil {
		return fmt.Errorf("failed to send request for %s: %w", url, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("received status code %d for %s", resp.StatusCode(), url)
	}

	return nil
}

func (a *Agent) encrypt(buf *bytes.Buffer) ([]byte, error) {
	encrypted := make([]byte, 0, buf.Len())

	step := a.publicKey.Size() - 11
	total := buf.Len()
	msg := buf.Bytes()

	for start := 0; start < total; start += step {
		finish := start + step
		if finish > total {
			finish = total
		}

		cipher, err := rsa.EncryptPKCS1v15(rand.Reader, a.publicKey, msg[start:finish])
		if err != nil {
			return nil, err
		}

		encrypted = append(encrypted, cipher...)
	}

	return encrypted, nil
}

func (a *Agent) calculateHash(msg []byte) string {
	h := hmac.New(sha256.New, []byte(a.config.Agent.Key))
	h.Write(msg)

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (a *Agent) applyGauge(name string, value float64, metrics []api.Metrics) []api.Metrics {
	m := api.Metrics{
		ID:    name,
		MType: api.GaugeType,
		Value: &value,
	}

	return append(metrics, m)
}

func (a *Agent) applyCounter(name string, value int64, metrics []api.Metrics) []api.Metrics {
	m := api.Metrics{
		ID:    name,
		MType: api.CounterType,
		Delta: &value,
	}

	return append(metrics, m)
}
