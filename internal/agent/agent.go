package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/kaa-it/go-devops/internal/api"

	"github.com/go-resty/resty/v2"
)

const (
	_requestTimeout = 5 * time.Second
)

type Agent struct {
	storage *Storage
	config  *Config
	client  *resty.Client
}

func New(client *resty.Client, config *Config) *Agent {
	client.SetTimeout(_requestTimeout)

	return &Agent{
		storage: NewStorage(),
		config:  config,
		client:  client,
	}
}

func (a *Agent) Run() {
	log.Println("Agent started")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	ctx, cancel := context.WithCancel(context.Background())

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go a.runPoller(ctx, wg)
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
	a.storage.UpdateGauge("RandomValue", rand.Float64())

	log.Println("Poll done")
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

	req.Header.Set("Content-Type", "application/json")
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

	req.SetBody(buf)

	resp, err := req.Send()
	if err != nil {
		return fmt.Errorf("failed to send request for %s: %w", url, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("received status code %d for %s", resp.StatusCode(), url)
	}

	return nil
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
