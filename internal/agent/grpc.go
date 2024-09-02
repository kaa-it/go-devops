package agent

import (
	"context"
	"errors"
	pb "github.com/kaa-it/go-devops/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
	"log"
)

func (a *Agent) initGRPC() error {
	conn, err := grpc.NewClient(a.config.Server.GRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	a.grpcClient = pb.NewMetricsClient(conn)
	a.grpcConn = conn

	return nil
}

func (a *Agent) reportGRPC() {
	var metrics []*pb.Metric

	a.storage.ForEachGauge(func(key string, value float64) {
		metrics = a.applyGaugeGRPC(key, value, metrics)
	})

	a.storage.ForEachCounter(func(key string, value int64) {
		metrics = a.applyCounterGRPC(key, value, metrics)

		// Subtract sent value to take into account
		// possible counter updates after sending
		a.storage.UpdateCounter(key, -value)
	})

	if err := a.sendMetricsGRPC(metrics); err != nil {
		log.Println(err)
		return
	}

	log.Println("Report done")
}

func (a *Agent) applyGaugeGRPC(name string, value float64, metrics []*pb.Metric) []*pb.Metric {
	m := &pb.Metric{
		Type: &pb.Metric_Gauge_{
			Gauge: &pb.Metric_Gauge{
				Name:  name,
				Value: value,
			},
		},
	}

	return append(metrics, m)
}

func (a *Agent) applyCounterGRPC(name string, value int64, metrics []*pb.Metric) []*pb.Metric {
	m := &pb.Metric{
		Type: &pb.Metric_Counter_{
			Counter: &pb.Metric_Counter{
				Name:  name,
				Value: value,
			},
		},
	}

	return append(metrics, m)
}

func (a *Agent) sendMetricsGRPC(metrics []*pb.Metric) error {
	resp, err := a.grpcClient.Updates(context.Background(), &pb.UpdatesRequest{
		Metrics: metrics,
	}, grpc.UseCompressor(gzip.Name))

	if err != nil {
		return err
	}

	if resp.Error != "" {
		return errors.New(resp.Error)
	}

	return nil
}
