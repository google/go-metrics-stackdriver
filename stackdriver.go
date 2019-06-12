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
	"bytes"
	"context"
	"fmt"
	"log"
	"path"
	"strings"
	"sync"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	metrics "github.com/armon/go-metrics"
	googlepb "github.com/golang/protobuf/ptypes/timestamp"
	distributionpb "google.golang.org/genproto/googleapis/api/distribution"
	"google.golang.org/genproto/googleapis/api/metric"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

// Sink conforms to the metrics.MetricSink interface and is used to transmit
// metrics information to stackdriver.
//
// Sink performs in-process aggregation of metrics to limit calls to
// stackdriver.
type Sink struct {
	client    *monitoring.MetricClient
	interval  time.Duration
	firstTime time.Time

	gauges     map[string]*gauge
	counters   map[string]*counter
	histograms map[string]*histogram

	bucketer BucketFn
	taskInfo *TaskInfo

	sync.Mutex
}

// TaskInfo must uniquely identify the process that the Sink metrics are
// collecting. These will be added as labels to the metric so that values can
// be filtered/aggregated correctly.
//
// https://cloud.google.com/monitoring/api/resources#tag_generic_task
type TaskInfo struct {
	ProjectID string
	Location  string
	Namespace string
	Job       string
	TaskID    string
}

// BucketFn returns the histogram bucket thresholds based on the given metric
// name.
type BucketFn func(string) []float64

// DefaultBucketer is the default BucketFn used to determing bucketing values
// for metrics.
func DefaultBucketer(name string) []float64 {
	return []float64{10.0, 25.0, 50.0, 100.0, 150.0, 250.0, 300.0, 500.0, 1000.0, 2000.0, 5000.0}
}

// NewSink creates a Sink to flush metrics to stackdriver every interval. The
// interval should be greater than 1 minute.
func NewSink(interval time.Duration, taskInfo *TaskInfo, client *monitoring.MetricClient) *Sink {
	s := &Sink{
		client:   client,
		interval: interval,
		bucketer: DefaultBucketer,
		taskInfo: taskInfo,
	}
	s.reset()

	// run cancelable goroutine that reports on interval
	go s.flushMetrics(context.Background())

	return s
}

func (s *Sink) flushMetrics(ctx context.Context) {
	if s.interval == 0*time.Second {
		return
	}

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("stopped flushing metrics")
			return
		case <-ticker.C:
			s.report(ctx)
		}
	}
}

func (s *Sink) reset() {
	s.Lock()
	defer s.Unlock()

	s.firstTime = time.Now()
	s.gauges = make(map[string]*gauge)
	s.counters = make(map[string]*counter)
	s.histograms = make(map[string]*histogram)
}

func (s *Sink) deep() (time.Time, map[string]*gauge, map[string]*counter, map[string]*histogram) {
	rGauges := make(map[string]*gauge, len(s.gauges))
	rCounters := make(map[string]*counter, len(s.counters))
	rHistograms := make(map[string]*histogram, len(s.histograms))

	s.Lock()
	end := time.Now()
	for k, v := range s.gauges {
		rGauges[k] = &gauge{
			name:  v.name,
			value: v.value,
		}
	}
	for k, v := range s.counters {
		rCounters[k] = &counter{
			name:  v.name,
			value: v.value,
		}
	}
	for k, v := range s.histograms {
		r := &histogram{
			name:    v.name,
			buckets: v.buckets,
			counts:  make([]int64, len(v.counts)),
		}
		copy(r.counts, v.counts)
		rHistograms[k] = r
	}
	s.Unlock()

	return end, rGauges, rCounters, rHistograms
}

func (s *Sink) report(ctx context.Context) {
	end, rGauges, rCounters, rHistograms := s.deep()

	if s.client == nil {
		return
	}

	// https://cloud.google.com/monitoring/api/resources
	resource := &monitoredrespb.MonitoredResource{
		Type: "generic_task",
		Labels: map[string]string{
			"project_id": s.taskInfo.ProjectID,
			"location":   s.taskInfo.Location,
			"namespace":  s.taskInfo.Namespace,
			"job":        s.taskInfo.Job,
			"task_id":    s.taskInfo.TaskID,
		},
	}

	for _, v := range rCounters {
		err := s.client.CreateTimeSeries(ctx, &monitoringpb.CreateTimeSeriesRequest{
			Name: fmt.Sprintf("projects/%s", s.taskInfo.ProjectID),
			TimeSeries: []*monitoringpb.TimeSeries{
				{
					Metric: &metricpb.Metric{
						Type:   path.Join("custom.googleapis.com", "go-metrics", v.name.name),
						Labels: v.name.labelMap(),
					},
					MetricKind: metric.MetricDescriptor_GAUGE,
					Resource:   resource,
					Points: []*monitoringpb.Point{
						&monitoringpb.Point{
							Interval: &monitoringpb.TimeInterval{
								EndTime: &googlepb.Timestamp{
									Seconds: end.Unix(),
								},
							},
							Value: &monitoringpb.TypedValue{
								Value: &monitoringpb.TypedValue_DoubleValue{
									DoubleValue: v.value,
								},
							},
						},
					},
				},
			},
		})

		if err != nil {
			log.Printf("Failed to write time series data: %v", err)
		}
	}

	for _, v := range rGauges {
		err := s.client.CreateTimeSeries(ctx, &monitoringpb.CreateTimeSeriesRequest{
			Name: fmt.Sprintf("projects/%s", s.taskInfo.ProjectID),
			TimeSeries: []*monitoringpb.TimeSeries{
				{
					Metric: &metricpb.Metric{
						Type:   path.Join("custom.googleapis.com", "go-metrics", v.name.name),
						Labels: v.name.labelMap(),
					},
					MetricKind: metric.MetricDescriptor_GAUGE,
					Resource:   resource,
					Points: []*monitoringpb.Point{
						&monitoringpb.Point{
							Interval: &monitoringpb.TimeInterval{
								EndTime: &googlepb.Timestamp{
									Seconds: end.Unix(),
								},
							},
							Value: &monitoringpb.TypedValue{
								Value: &monitoringpb.TypedValue_DoubleValue{
									DoubleValue: float64(v.value),
								},
							},
						},
					},
				},
			},
		})

		if err != nil {
			log.Printf("Failed to write time series data: %v", err)
		}
	}

	for _, v := range rHistograms {
		var count int64
		count = 0
		for _, i := range v.counts {
			count += int64(i)
		}

		err := s.client.CreateTimeSeries(ctx, &monitoringpb.CreateTimeSeriesRequest{
			Name: fmt.Sprintf("projects/%s", s.taskInfo.ProjectID),
			TimeSeries: []*monitoringpb.TimeSeries{
				{
					Metric: &metricpb.Metric{
						Type:   path.Join("custom.googleapis.com", "go-metrics", v.name.name),
						Labels: v.name.labelMap(),
					},
					MetricKind: metric.MetricDescriptor_CUMULATIVE,
					Resource:   resource,
					Points: []*monitoringpb.Point{
						&monitoringpb.Point{
							Interval: &monitoringpb.TimeInterval{
								StartTime: &googlepb.Timestamp{
									Seconds: s.firstTime.Unix(),
								},
								EndTime: &googlepb.Timestamp{
									Seconds: end.Unix(),
								},
							},
							Value: &monitoringpb.TypedValue{
								Value: &monitoringpb.TypedValue_DistributionValue{
									DistributionValue: &distributionpb.Distribution{
										BucketOptions: &distributionpb.Distribution_BucketOptions{
											Options: &distributionpb.Distribution_BucketOptions_ExplicitBuckets{
												ExplicitBuckets: &distributionpb.Distribution_BucketOptions_Explicit{
													Bounds: v.buckets,
												},
											},
										},
										BucketCounts: v.counts,
										Count:        count,
									},
								},
							},
						},
					},
				},
			},
		})

		if err != nil {
			log.Printf("Failed to write time series data: %v", err)
		}
	}
}

// A Gauge should retain the last value it is set to.
func (s *Sink) SetGauge(key []string, val float32) {
	s.SetGaugeWithLabels(key, val, nil)
}

// A Gauge should retain the last value it is set to.
func (s *Sink) SetGaugeWithLabels(key []string, val float32, labels []metrics.Label) {
	n := newSeries(key, labels)

	g := &gauge{
		name:  n,
		value: val,
	}

	s.Lock()
	defer s.Unlock()

	s.gauges[n.key] = g
}

// Should emit a Key/Value pair for each call.
func (s *Sink) EmitKey(key []string, val float32) {
	// EmitKey is not implemented for stackdriver
}

// Counters should accumulate values.
func (s *Sink) IncrCounter(key []string, val float32) {
	s.IncrCounterWithLabels(key, val, nil)
}

// Counters should accumulate values.
func (s *Sink) IncrCounterWithLabels(key []string, val float32, labels []metrics.Label) {
	n := newSeries(key, labels)

	s.Lock()
	defer s.Unlock()

	c, ok := s.counters[n.key]
	if ok {
		c.value += float64(val)
	} else {
		s.counters[n.key] = &counter{
			name:  n,
			value: float64(val),
		}
	}
}

// Samples are for timing information, where quantiles are used.
func (s *Sink) AddSample(key []string, val float32) {
	s.AddSampleWithLabels(key, val, nil)
}

// Samples are for timing information, where quantiles are used.
func (s *Sink) AddSampleWithLabels(key []string, val float32, labels []metrics.Label) {
	n := newSeries(key, labels)

	s.Lock()
	defer s.Unlock()

	h, ok := s.histograms[n.key]
	if ok {
		h.addSample(val)
	} else {
		h = newHistogram(n, s.bucketer)
		h.addSample(val)
		s.histograms[n.key] = h
	}
}

var _ metrics.MetricSink = (*Sink)(nil)

// Series holds the naming for a timeseries metric.
type series struct {
	key    string
	name   string
	labels []metrics.Label
}

func newSeries(key []string, labels []metrics.Label) *series {
	buf := &bytes.Buffer{}
	replacer := strings.NewReplacer(" ", "_")

	if len(key) > 0 {
		replacer.WriteString(buf, key[0])
	}
	for _, k := range key[1:] {
		replacer.WriteString(buf, ".")
		replacer.WriteString(buf, k)
	}

	name := buf.String()

	for _, label := range labels {
		replacer.WriteString(buf, fmt.Sprintf(";%s=%s", label.Name, label.Value))
	}

	return &series{
		key:    buf.String(),
		name:   name,
		labels: labels,
	}
}

func (s *series) labelMap() map[string]string {
	o := make(map[string]string, len(s.labels))
	for _, v := range s.labels {
		o[v.Name] = v.Value
	}
	return o
}

// https://cloud.google.com/monitoring/api/ref_v3/rest/v3/TimeSeries#point
type gauge struct {
	name  *series
	value float32
}

// https://cloud.google.com/monitoring/api/ref_v3/rest/v3/TimeSeries#point
type counter struct {
	name  *series
	value float64
}

// https://cloud.google.com/monitoring/api/ref_v3/rest/v3/TimeSeries#distribution
type histogram struct {
	name    *series
	buckets []float64
	counts  []int64
}

func newHistogram(name *series, bucketer BucketFn) *histogram {
	h := &histogram{
		name:    name,
		buckets: bucketer(name.name),
	}
	h.counts = make([]int64, len(h.buckets))
	return h
}

func (h *histogram) addSample(val float32) {
	for i := range h.buckets {
		if float64(val) <= h.buckets[i] {
			h.counts[i]++
			return
		}
	}

	h.counts[len(h.buckets)-1]++
}
