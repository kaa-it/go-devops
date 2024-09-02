// Package agent contains implementation for metric agent.
//
// The agent collects counter and gauge metrics to storage and sends their to server with interval.
package agent

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"google.golang.org/grpc"
	"log"
	math "math/rand"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	pb "github.com/kaa-it/go-devops/internal/proto"
)

// Agent describes metric agent.
type Agent struct {
	storage    *Storage
	config     *Config
	client     *resty.Client
	grpcClient pb.MetricsClient
	grpcConn   *grpc.ClientConn
	publicKey  *rsa.PublicKey
}

// New creates new metric agent
//
// Takes client to connect to metric server and config with total configuration of agent.
func New(config *Config) (*Agent, error) {

	var publicKey *rsa.PublicKey

	if config.Agent.PublicKeyPath != "" {
		publicKeyPEM, err := os.ReadFile(config.Agent.PublicKeyPath)
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

	agent := &Agent{
		storage:   NewStorage(),
		config:    config,
		publicKey: publicKey,
	}

	if config.Server.Address != "" {
		if err := agent.initREST(); err != nil {
			return nil, err
		}
	}

	if config.Server.GRPCAddress != "" {
		if err := agent.initGRPC(); err != nil {
			return nil, err
		}
	}

	return agent, nil
}

// SetRESTClient allows to set custom REST client
func (a *Agent) SetRESTClient(client *resty.Client) {
	a.client = client
}

// Run runs metric agent and control its lifecycle.
func (a *Agent) Run() {
	log.Println("Agent started")

	if a.config.Server.GRPCAddress != "" {
		defer a.grpcConn.Close()
	}

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
			if a.config.Server.Address != "" {
				a.reportREST()
			}

			if a.config.Server.GRPCAddress != "" {
				a.reportGRPC()
			}
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
