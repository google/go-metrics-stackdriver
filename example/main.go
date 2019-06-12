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
package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"

	stackdriver "github.com/google/go-metrics-stackdriver"

	"cloud.google.com/go/compute/metadata"
	monitoring "cloud.google.com/go/monitoring/apiv3"
	metrics "github.com/armon/go-metrics"
)

func main() {
	// setup client
	ctx, cancel := context.WithCancel(context.Background())
	client, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	projectID, err := metadata.ProjectID()
	if err != nil {
		log.Fatal(err)
	}

	if os.Getenv("GOOGLE_CLOUD_PROJECT") != "" {
		projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}

	taskInfo := &stackdriver.TaskInfo{
		ProjectID: projectID,
		Location:  "us-east1-c",
		Job:       "example",
	}

	log.Printf("initializing sink: %v", taskInfo)

	// create sink
	ss := stackdriver.NewSink(60*time.Second, taskInfo, client)

	// capture ctrl+c
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		log.Printf("ctrl+c detected... shutting down")
		cancel()
	}()

	// generate data
	log.Printf("sending data")
	exercise(ctx, ss)
}

func exercise(ctx context.Context, m metrics.MetricSink) {
	for {
		select {
		case <-time.After(100 * time.Millisecond):
			m.SetGauge([]string{"foo"}, 42)
			m.IncrCounter([]string{"baz"}, 1)
			m.AddSample([]string{"method", "rand"}, 500*rand.Float32())
			m.AddSample([]string{"method", "const"}, 200)
			m.AddSample([]string{"method", "dist"}, 50)
			m.AddSample([]string{"method", "dist"}, 100)
			m.AddSample([]string{"method", "dist"}, 150)
		case <-ctx.Done():
			log.Printf("terminating")
			return
		}
	}
}
