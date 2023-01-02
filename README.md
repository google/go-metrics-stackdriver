# go-metrics-stackdriver
[![godoc](https://godoc.org/github.com/google/go-metrics-stackdriver?status.svg)](http://godoc.org/github.com/google/go-metrics-stackdriver)

This library provides a stackdriver sink for applications instrumented with the
[go-metrics](https://github.com/armon/go-metrics) library.

## 🚨 Warning

__This is not an officially supported Google product.__

In general the author of this package would recommend instrumenting custom metrics for new code by following the [official GCP documentation](https://cloud.google.com/monitoring/custom-metrics), especially for new applications.

This package is intended as a way to publish metrics for applications that are _already_ instrumented with `go-metrics` without having to use a sidecar process like [stackdriver-prometheus-sidecar](https://github.com/Stackdriver/stackdriver-prometheus-sidecar).

## 🚨 Upgrading

Between v0.5.0 and v0.6.0, the behavior of the `IncrCounter()` method changed: previously it would create a `GAUGE` [metric kind](https://cloud.google.com/monitoring/api/v3/kinds-and-types), but from v0.6.0 forward it will create a `CUMULATIVE` metric kind.  (See https://github.com/google/go-metrics-stackdriver/issues/18 for a discussion.)

However, once a [MetricDescriptor](https://cloud.google.com/logging/docs/reference/v2/rest/v2/projects.metrics#MetricDescriptor) has been created in Google Cloud Monitoring, its `metricKind` field cannot be changed.  So if you have any _existing_ `GAUGE` metrics that were created by `IncrCounter()`, you will see errors in your logs when the v0.6.0 client attempts to update them and fails.  Your options for handling this are:

1. Change the name of the metric you are passing to `IncrCounter` (creating a new metricDescriptor with a different name), or:
2. Delete the existing metricDescriptor using the [delete API](https://cloud.google.com/monitoring/api/ref_v3/rest/v3/projects.metricDescriptors/delete) and let go-metrics re-create it as a `CUMULATIVE` metric

Additionally, v0.6.0 adds `ResetCounter()` and `ResetCounterWithLabels()` methods: calling these methods resets the counter value to zero.

## Details

[stackdriver.NewSink](https://godoc.org/github.com/google/go-metrics-stackdriver#NewSink)'s return value satisfies the go-metrics library's [MetricSink](https://godoc.org/github.com/armon/go-metrics#MetricSink) interface. When providing a `stackdriver.Sink` to libraries and applications instrumented against `MetricSink`, the metrics will be aggregated within this library and written to stackdriver as [Generic Task](https://cloud.google.com/monitoring/api/resources#tag_generic_task) timeseries metrics.

## Example

```go
import "github.com/google/go-metrics-stackdriver"
...
client, _ := monitoring.NewMetricClient(context.Background())
ss := stackdriver.NewSink(client, &stackdriver.Config{
  ProjectID: projectID,
})
...
ss.SetGauge([]string{"foo"}, 42)
ss.IncrCounter([]string{"baz"}, 1)
ss.AddSample([]string{"method", "const"}, 200)
```

The [full example](example/main.go) can be run from a cloud shell console to test how metrics are collected and displayed.

You can also try out the example using Cloud Run!

[![Run on Google Cloud](https://storage.googleapis.com/cloudrun/button.svg)](https://console.cloud.google.com/cloudshell/editor?shellonly=true&cloudshell_image=gcr.io/cloudrun/button&cloudshell_git_repo=https://github.com/google/go-metrics-stackdriver.git&cloudshell_working_dir=example)
