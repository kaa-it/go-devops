package agent

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

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
	a.storage.ForEachGauge(func(key string, value float64) {
		if err := a.sendGauge(key, value); err != nil {
			log.Println(err)
		}
	})

	a.storage.ForEachCounter(func(key string, value int64) {
		if err := a.sendCounter(key, value); err != nil {
			log.Println(err)
		}
	})

	log.Println("Report done")
}

func (a *Agent) sendMetric(url string) error {
	req := a.client.R()
	req.Method = http.MethodPost
	req.URL = url

	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Content-Length", "0")

	resp, err := req.Send()
	if err != nil {
		return fmt.Errorf("failed to send request for %s: %w", url, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("received status code %d for %s", resp.StatusCode(), url)
	}

	return nil
}

func (a *Agent) sendGauge(name string, value float64) error {
	strVal := strconv.FormatFloat(value, 'f', 4, 64)

	url := fmt.Sprintf(
		"http://%s/update/%s/%s/%s",
		a.config.Server.Address,
		"gauge",
		name,
		strVal,
	)

	return a.sendMetric(url)
}

func (a *Agent) sendCounter(name string, value int64) error {
	strVal := strconv.FormatInt(value, 10)

	url := fmt.Sprintf(
		"http://%s/update/%s/%s/%s",
		a.config.Server.Address,
		"counter",
		name,
		strVal,
	)

	return a.sendMetric(url)
}
