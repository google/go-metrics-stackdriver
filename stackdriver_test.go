// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package stackdriver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	metrics "github.com/armon/go-metrics"
	emptypb "github.com/golang/protobuf/ptypes/empty"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/option"
	distributionpb "google.golang.org/genproto/googleapis/api/distribution"
	"google.golang.org/genproto/googleapis/api/metric"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func benchmarkAddSample(concurrency int, b *testing.B) {
	ss := newTestSink(100*time.Millisecond, nil)
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			// Note: b.N is run for each goroutine, when
			// interpreting ns/ops results remember to normalize
			// for the concurrency parameter.
			for i := 0; i < b.N; i++ {
				ss.AddSample([]string{"foo", "bar"}, float32(i)*0.3)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	// do something with ss so that the compiler doesn't optimize it away
	b.Logf("%d", len(ss.histograms))
}

func BenchmarkAddSample1(b *testing.B)   { benchmarkAddSample(1, b) }
func BenchmarkAddSample2(b *testing.B)   { benchmarkAddSample(2, b) }
func BenchmarkAddSample10(b *testing.B)  { benchmarkAddSample(10, b) }
func BenchmarkAddSample50(b *testing.B)  { benchmarkAddSample(50, b) }
func BenchmarkAddSample100(b *testing.B) { benchmarkAddSample(100, b) }

// All metrics collection is paused while a copy of the current state is
// snapshotted. We isolate and benchmark this copy since all metrics collection
// functions will block until the copy completes.
func benchmarkCopy(samples, gauges, counters int, b *testing.B) {
	ss := newTestSink(0*time.Second, nil)
	for i := 0; i < samples; i++ {
		ss.AddSample([]string{fmt.Sprintf("%d", i)}, float32(i)*0.3)
	}
	for i := 0; i < gauges; i++ {
		ss.SetGauge([]string{fmt.Sprintf("%d", i)}, float32(i)*0.3)
	}
	for i := 0; i < counters; i++ {
		ss.IncrCounter([]string{fmt.Sprintf("%d", i)}, float32(i)*0.3)
	}

	var n int
	for i := 0; i < b.N; i++ {
		_, s, g, c := ss.deep()
		// do something with the copy so that the compiler doesn't optimize it away
		n = len(s) + len(g) + len(c)
	}
	b.Logf("%d", n)
}

func BenchmarkReport1(b *testing.B)   { benchmarkCopy(1, 1, 1, b) }
func BenchmarkReport10(b *testing.B)  { benchmarkCopy(10, 10, 10, b) }
func BenchmarkReport50(b *testing.B)  { benchmarkCopy(50, 50, 50, b) }
func BenchmarkReport100(b *testing.B) { benchmarkCopy(100, 100, 100, b) }

func TestSample(t *testing.T) {
	ss := newTestSink(0*time.Second, nil)

	tests := []struct {
		name     string
		collect  func()
		createFn func(*testing.T) func(context.Context, *monitoringpb.CreateTimeSeriesRequest) (*emptypb.Empty, error)
	}{
		{
			name: "histogram",
			collect: func() {
				ss.AddSample([]string{"foo", "bar"}, 5.0)
			},
			createFn: func(t *testing.T) func(context.Context, *monitoringpb.CreateTimeSeriesRequest) (*emptypb.Empty, error) {
				return func(_ context.Context, req *monitoringpb.CreateTimeSeriesRequest) (*emptypb.Empty, error) {
					if req.TimeSeries[0].Points[0].Value.GetDistributionValue().BucketCounts[0] == 1 {
						return &emptypb.Empty{}, nil
					}
					t.Errorf("unexpected CreateTimeSeriesRequest\nwant: %s\ngot: %v", "bucket 0 count 1", req)
					return nil, errors.New("unexpected CreateTimeSeriesRequest")
				}
			},
		},
		{
			name: "hisogram with samples in multiple buckets",
			collect: func() {
				ss.AddSample([]string{"foo", "bar"}, 5.0)
				ss.AddSample([]string{"foo", "bar"}, 100.0)
				ss.AddSample([]string{"foo", "bar"}, 500.0)
			},
			createFn: func(t *testing.T) func(context.Context, *monitoringpb.CreateTimeSeriesRequest) (*emptypb.Empty, error) {
				return func(_ context.Context, req *monitoringpb.CreateTimeSeriesRequest) (*emptypb.Empty, error) {
					want := &monitoringpb.CreateTimeSeriesRequest{
						Name: "projects/foo",
						TimeSeries: []*monitoringpb.TimeSeries{
							&monitoringpb.TimeSeries{
								Metric: &metricpb.Metric{
									Type: "custom.googleapis.com/go-metrics/foo_bar",
								},
								MetricKind: metric.MetricDescriptor_CUMULATIVE,
								Points: []*monitoringpb.Point{
									&monitoringpb.Point{
										Value: &monitoringpb.TypedValue{
											Value: &monitoringpb.TypedValue_DistributionValue{
												DistributionValue: &distributionpb.Distribution{
													BucketCounts: []int64{1, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
													Count:        3,
												},
											},
										},
									},
								},
							},
						},
					}
					if diff := diffCreateMsg(want, req); diff != "" {
						t.Errorf("unexpected CreateTimeSeriesRequest (-want +got):\n%s", diff)
					}
					return &emptypb.Empty{}, nil
				}
			},
		},
		{
			name: "hisogram with multiple samples in one bucket",
			collect: func() {
				ss.AddSample([]string{"foo", "bar"}, 5.0)
				ss.AddSample([]string{"foo", "bar"}, 5.0)
				ss.AddSample([]string{"foo", "bar"}, 5.0)
			},
			createFn: func(t *testing.T) func(context.Context, *monitoringpb.CreateTimeSeriesRequest) (*emptypb.Empty, error) {
				return func(_ context.Context, req *monitoringpb.CreateTimeSeriesRequest) (*emptypb.Empty, error) {
					if req.TimeSeries[0].Points[0].Value.GetDistributionValue().BucketCounts[0] == 3 {
						return &emptypb.Empty{}, nil
					}
					t.Errorf("unexpected CreateTimeSeriesRequest\nwant: %s\ngot: %v", "bucket 0 count 3", req)
					return nil, errors.New("unexpected CreateTimeSeriesRequest")
				}
			},
		},
		{
			name: "counter",
			collect: func() {
				ss.IncrCounter([]string{"foo", "bar"}, 1.0)
			},
			createFn: func(t *testing.T) func(context.Context, *monitoringpb.CreateTimeSeriesRequest) (*emptypb.Empty, error) {
				return func(_ context.Context, req *monitoringpb.CreateTimeSeriesRequest) (*emptypb.Empty, error) {
					if req.TimeSeries[0].Points[0].Value.GetDoubleValue() == 1.0 {
						return &emptypb.Empty{}, nil
					}
					t.Errorf("unexpected CreateTimeSeriesRequest\nwant: %s\ngot: %v", "value 1.0", req)
					return nil, errors.New("unexpected CreateTimeSeriesRequest")
				}
			},
		},
		{
			name: "multiple counts",
			collect: func() {
				ss.IncrCounter([]string{"foo", "bar"}, 1.0)
				ss.IncrCounter([]string{"foo", "bar"}, 1.0)
				ss.IncrCounter([]string{"foo", "bar"}, 1.0)
			},
			createFn: func(t *testing.T) func(context.Context, *monitoringpb.CreateTimeSeriesRequest) (*emptypb.Empty, error) {
				return func(_ context.Context, req *monitoringpb.CreateTimeSeriesRequest) (*emptypb.Empty, error) {
					if req.TimeSeries[0].Points[0].Value.GetDoubleValue() == 3.0 {
						return &emptypb.Empty{}, nil
					}
					t.Errorf("unexpected CreateTimeSeriesRequest\nwant: %s\ngot: %v", "value 3.0", req)
					return nil, errors.New("unexpected CreateTimeSeriesRequest")
				}
			},
		},
		{
			name: "gauge",
			collect: func() {
				ss.SetGauge([]string{"foo", "bar"}, 50.0)
			},
			createFn: func(t *testing.T) func(context.Context, *monitoringpb.CreateTimeSeriesRequest) (*emptypb.Empty, error) {
				return func(_ context.Context, req *monitoringpb.CreateTimeSeriesRequest) (*emptypb.Empty, error) {
					if req.TimeSeries[0].Points[0].Value.GetDoubleValue() == 50.0 {
						return &emptypb.Empty{}, nil
					}
					t.Errorf("unexpected CreateTimeSeriesRequest\nwant: %s\ngot: %v", "value 50.0", req)
					return nil, errors.New("unexpected CreateTimeSeriesRequest")
				}
			},
		},
		{
			name: "repeated gauge",
			collect: func() {
				ss.SetGauge([]string{"foo", "bar"}, 50.0)
				ss.SetGauge([]string{"foo", "bar"}, 50.0)
				ss.SetGauge([]string{"foo", "bar"}, 50.0)
			},
			createFn: func(t *testing.T) func(context.Context, *monitoringpb.CreateTimeSeriesRequest) (*emptypb.Empty, error) {
				return func(_ context.Context, req *monitoringpb.CreateTimeSeriesRequest) (*emptypb.Empty, error) {
					if req.TimeSeries[0].Points[0].Value.GetDoubleValue() == 50.0 {
						return &emptypb.Empty{}, nil
					}
					t.Errorf("unexpected CreateTimeSeriesRequest\nwant: %s\ngot: %v", "value 50.0", req)
					return nil, errors.New("unexpected CreateTimeSeriesRequest")
				}
			},
		},
		{
			name: "changing gauge",
			collect: func() {
				ss.SetGauge([]string{"foo", "bar"}, 50.0)
				ss.SetGauge([]string{"foo", "bar"}, 51.0)
				ss.SetGauge([]string{"foo", "bar"}, 52.0)
			},
			createFn: func(t *testing.T) func(context.Context, *monitoringpb.CreateTimeSeriesRequest) (*emptypb.Empty, error) {
				return func(_ context.Context, req *monitoringpb.CreateTimeSeriesRequest) (*emptypb.Empty, error) {
					if req.TimeSeries[0].Points[0].Value.GetDoubleValue() == 52.0 {
						return &emptypb.Empty{}, nil
					}
					t.Errorf("unexpected CreateTimeSeriesRequest\nwant: %s\ngot: %v", "value 52.0", req)
					return nil, errors.New("unexpected CreateTimeSeriesRequest")
				}
			},
		},
		{
			name: "batching",
			collect: func() {
				for i := 0; i < 300; i++ {
					ss.SetGauge([]string{"foo", fmt.Sprintf("%d", i)}, 50.0)
				}
			},
			createFn: func(t *testing.T) func(context.Context, *monitoringpb.CreateTimeSeriesRequest) (*emptypb.Empty, error) {
				return func(_ context.Context, req *monitoringpb.CreateTimeSeriesRequest) (*emptypb.Empty, error) {
					// 300 TimeSeries were created, we expect 2 RPCs (first with 200 and second with 100 TS)
					if len(req.TimeSeries) == 200 || len(req.TimeSeries) == 100 {
						return &emptypb.Empty{}, nil
					}
					t.Errorf("unexpected CreateTimeSeriesRequest\ngot(# of TimeSeries): %v", len(req.TimeSeries))
					return nil, errors.New("unexpected CreateTimeSeriesRequest")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			lis := bufconn.Listen(1024 * 1024)
			serv := grpc.NewServer()
			monitoringpb.RegisterMetricServiceServer(serv, &mockMetricServer{
				createFn: tc.createFn(t),
			})

			go func() {
				if err := serv.Serve(lis); err != nil {
					t.Fatalf("server error: %v", err)
				}
			}()

			conn, err := grpc.Dial(lis.Addr().String(), grpc.WithDialer(func(string, time.Duration) (net.Conn, error) { return lis.Dial() }), grpc.WithInsecure())
			if err != nil {
				t.Fatalf("failed to dial: %v", err)
			}
			defer conn.Close()
			client, err := monitoring.NewMetricClient(ctx, option.WithGRPCConn(conn))
			if err != nil {
				t.Fatalf("failed to create MetricClient: %v", err)
			}

			ss.reset()
			ss.client = client
			tc.collect()
			ss.report(ctx)
		})
	}
}

func TestExtract(t *testing.T) {
	ss := newTestSink(0*time.Second, nil)
	ss.extractor = func(key []string, kind string) ([]string, []metrics.Label, error) {
		return key[:1], []metrics.Label{
			{
				Name:  "method",
				Value: key[1],
			},
		}, nil
	}

	tests := []struct {
		name     string
		collect  func()
		createFn func(*testing.T) func(context.Context, *monitoringpb.CreateTimeSeriesRequest) (*emptypb.Empty, error)
	}{
		{
			name: "histogram",
			collect: func() {
				ss.AddSample([]string{"foo", "bar"}, 5.0)
			},
			createFn: func(t *testing.T) func(context.Context, *monitoringpb.CreateTimeSeriesRequest) (*emptypb.Empty, error) {
				return func(_ context.Context, req *monitoringpb.CreateTimeSeriesRequest) (*emptypb.Empty, error) {
					metric := req.TimeSeries[0].GetMetric()
					if metric.GetType() == "custom.googleapis.com/go-metrics/foo" && metric.GetLabels()["method"] == "bar" && req.TimeSeries[0].Points[0].Value.GetDistributionValue().BucketCounts[0] == 1 {
						return &emptypb.Empty{}, nil
					}
					t.Errorf("unexpected CreateTimeSeriesRequest\nwant: %s\ngot: %v", "bucket 0 count 1", req)
					return nil, errors.New("unexpected CreateTimeSeriesRequest")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			lis := bufconn.Listen(1024 * 1024)
			serv := grpc.NewServer()
			monitoringpb.RegisterMetricServiceServer(serv, &mockMetricServer{
				createFn: tc.createFn(t),
			})

			go func() {
				if err := serv.Serve(lis); err != nil {
					t.Fatalf("server error: %v", err)
				}
			}()

			conn, err := grpc.Dial(lis.Addr().String(), grpc.WithDialer(func(string, time.Duration) (net.Conn, error) { return lis.Dial() }), grpc.WithInsecure())
			if err != nil {
				t.Fatalf("failed to dial: %v", err)
			}
			defer conn.Close()
			client, err := monitoring.NewMetricClient(ctx, option.WithGRPCConn(conn))
			if err != nil {
				t.Fatalf("failed to create MetricClient: %v", err)
			}

			ss.reset()
			ss.client = client
			tc.collect()
			ss.report(ctx)
		})
	}
}

type mockMetricServer struct {
	monitoringpb.MetricServiceServer

	createFn func(context.Context, *monitoringpb.CreateTimeSeriesRequest) (*emptypb.Empty, error)
}

func (s *mockMetricServer) CreateTimeSeries(ctx context.Context, req *monitoringpb.CreateTimeSeriesRequest) (*emptypb.Empty, error) {
	if s.createFn != nil {
		return s.createFn(ctx, req)
	}
	return nil, errors.New("unimplemented")
}

// Skips defaults that are not appropriate for tests.
func newTestSink(interval time.Duration, client *monitoring.MetricClient) *Sink {
	s := &Sink{}
	s.taskInfo = &taskInfo{
		ProjectID: "foo",
	}
	s.interval = interval
	s.bucketer = DefaultBucketer
	s.extractor = DefaultLabelExtractor
	s.reset()
	go s.flushMetrics(context.Background())
	return s
}

func diffCreateMsg(want, got *monitoringpb.CreateTimeSeriesRequest) string {
	out := ""
	if want.GetName() != "" && (want.GetName() != got.GetName()) {
		out += fmt.Sprintf("Unexpected Name, got: %s, want:%s\n", got.GetName(), want.GetName())
	}

	for i := range want.GetTimeSeries() {
		w := want.GetTimeSeries()[i]
		g := got.GetTimeSeries()[i]

		if w.GetMetricKind() != g.GetMetricKind() {
			out += fmt.Sprintf("Unexpected MetricKind, got: %s, want:%s\n", g.GetMetricKind(), w.GetMetricKind())
		}

		if w.GetMetric().GetType() != g.GetMetric().GetType() {
			out += fmt.Sprintf("Unexpected Metric Type, got: %s, want:%s\n", g.GetMetric().GetType(), w.GetMetric().GetType())
		}

		if len(w.GetMetric().GetLabels()) != 0 {
			d := cmp.Diff(g.GetMetric().GetLabels(), w.GetMetric().GetLabels())
			if d != "" {
				out += fmt.Sprintf("Unexpected metric labels diff:%s \n", d)
			}
		}

		for j := range w.GetPoints() {
			wp := w.GetPoints()[j]
			gp := g.GetPoints()[j]

			// TODO: support diffing the start/end times

			// gauge/count
			if wp.GetValue().GetDoubleValue() != gp.GetValue().GetDoubleValue() {
				out += fmt.Sprintf("Unexpected value (@point %d), got: %v, want:%v\n", j, gp.GetValue().GetDoubleValue(), wp.GetValue().GetDoubleValue())
			}

			// distribution
			if wd := wp.GetValue().GetDistributionValue(); wd != nil {
				gd := gp.GetValue().GetDistributionValue()
				// TODO: support diffing custom buckets
				d := cmp.Diff(gd.GetBucketCounts(), wd.GetBucketCounts())
				if d != "" {
					out += fmt.Sprintf("Unexpected bucket counts diff (@point %d):%s \n", j, d)
				}
				if gd.GetCount() != wd.GetCount() {
					out += fmt.Sprintf("Unexpected count (@point %d), got: %v, want: %v\n", j, gd.GetCount(), wd.GetCount())
				}
			}
		}
	}
	return out
}
