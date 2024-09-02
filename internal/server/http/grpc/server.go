// Package grpc contains GRPC API implementation for server
package grpc

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/kaa-it/go-devops/internal/api"
	pb "github.com/kaa-it/go-devops/internal/proto"
	"github.com/kaa-it/go-devops/internal/server/updating"
)

type Logger interface {
	Error(args ...interface{})
}

type metricsServer struct {
	pb.UnimplementedMetricsServer
	a updating.Service
	l Logger
}

type Server struct {
	grpcServer *grpc.Server
}

func NewServer(l Logger, a updating.Service) *Server {
	server := grpc.NewServer()
	pb.RegisterMetricsServer(server, newMetricsServer(l, a))

	return &Server{grpcServer: server}
}

func (s *Server) Run(address string) error {
	listen, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	return s.grpcServer.Serve(listen)
}

func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
}

func newMetricsServer(l Logger, a updating.Service) *metricsServer {
	return &metricsServer{
		l: l,
		a: a,
	}
}

func (ms *metricsServer) Ping(context.Context, *emptypb.Empty) (*pb.Response, error) {
	var response pb.Response

	return &response, nil
}

func (ms *metricsServer) Update(ctx context.Context, in *pb.UpdateRequest) (*pb.Response, error) {
	var response pb.Response

	switch x := in.Metric.GetType().(type) {
	case *pb.Metric_Counter_:
		if err := ms.a.UpdateCounter(ctx, x.Counter.Name, x.Counter.Value); err != nil {
			ms.l.Error(err.Error())
			response.Error = err.Error()
		}
	case *pb.Metric_Gauge_:
		if err := ms.a.UpdateGauge(ctx, x.Gauge.Name, x.Gauge.Value); err != nil {
			ms.l.Error(err.Error())
			response.Error = err.Error()
		}
	default:
		response.Error = "wrong metric type"
	}

	return &response, nil
}

func (ms *metricsServer) Updates(ctx context.Context, in *pb.UpdatesRequest) (*pb.Response, error) {
	var response pb.Response

	metrics := make([]api.Metrics, 0, len(in.Metrics))

	for _, m := range in.Metrics {
		switch x := m.GetType().(type) {
		case *pb.Metric_Counter_:
			metrics = append(metrics, api.Metrics{
				ID:    x.Counter.Name,
				MType: api.CounterType,
				Delta: &x.Counter.Value,
			})
		case *pb.Metric_Gauge_:
			metrics = append(metrics, api.Metrics{
				ID:    x.Gauge.Name,
				MType: api.GaugeType,
				Value: &x.Gauge.Value,
			})
		}
	}

	if err := ms.a.Updates(ctx, metrics); err != nil {
		ms.l.Error(fmt.Sprintf("batch update failed: %v", err.Error()))
		response.Error = err.Error()
	}

	return &response, nil
}
