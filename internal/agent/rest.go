package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/kaa-it/go-devops/internal/api"
	"log"
	"net"
	"net/http"
	"time"
)

const (
	_retryCount       = 3
	_retryWaitTime    = time.Second
	_retryMaxWaitTime = 5 * time.Second
	_retryDelay       = 2 * time.Second
	_requestTimeout   = 5 * time.Second
)

func (a *Agent) initREST() error {
	client := resty.New()

	client.SetTimeout(_requestTimeout)
	client.SetRetryCount(_retryCount)
	client.SetRetryWaitTime(_retryWaitTime)
	client.SetRetryMaxWaitTime(_retryMaxWaitTime)
	client.SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
		return _retryDelay, nil
	})

	ip, err := resolveHostIP()
	if err != nil {
		return err
	}

	client.SetHeader("X-Real-IP", ip)

	a.client = client

	return nil
}

func resolveHostIP() (string, error) {

	netInterfaceAddresses, err := net.InterfaceAddrs()

	if err != nil {
		return "", err
	}

	for _, netInterfaceAddress := range netInterfaceAddresses {

		networkIP, ok := netInterfaceAddress.(*net.IPNet)

		if ok && !networkIP.IP.IsLoopback() && networkIP.IP.To4() != nil {

			ip := networkIP.IP.String()

			return ip, nil
		}
	}

	return "", errors.New("address not found")
}

func (a *Agent) reportREST() {
	var metrics []api.Metrics
	a.storage.ForEachGauge(func(key string, value float64) {
		metrics = a.applyGaugeREST(key, value, metrics)
	})

	a.storage.ForEachCounter(func(key string, value int64) {
		metrics = a.applyCounterREST(key, value, metrics)

		// Subtract sent value to take into account
		// possible counter updates after sending
		a.storage.UpdateCounter(key, -value)
	})

	if err := a.sendMetricsREST(metrics); err != nil {
		log.Println(err)
		return
	}

	log.Println("Report done")
}

func (a *Agent) sendMetricsREST(metrics []api.Metrics) error {
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

func (a *Agent) applyGaugeREST(name string, value float64, metrics []api.Metrics) []api.Metrics {
	m := api.Metrics{
		ID:    name,
		MType: api.GaugeType,
		Value: &value,
	}

	return append(metrics, m)
}

func (a *Agent) applyCounterREST(name string, value int64, metrics []api.Metrics) []api.Metrics {
	m := api.Metrics{
		ID:    name,
		MType: api.CounterType,
		Delta: &value,
	}

	return append(metrics, m)
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
