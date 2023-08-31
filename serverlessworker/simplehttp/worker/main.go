package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/temporalio/samples-go/serverlessworker/simplehttp"
	"github.com/uber-go/tally/v4"
	"github.com/uber-go/tally/v4/prometheus"
	"go.temporal.io/sdk/client"
	sdktally "go.temporal.io/sdk/contrib/tally"
	"go.temporal.io/sdk/worker"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Listen on port
	l, err := net.Listen("tcp", "127.0.0.1:9233")
	if err != nil {
		return fmt.Errorf("failed listening: %w", err)
	}
	defer l.Close()

	// Create handler
	h, err := NewOnDemandWorkerHandler(worker.OnDemandWorkflowWorkerOptions{
		TaskQueue: "serverlessworker-simplehttp",
		Workflows: []worker.RegisteredWorkflow{{Workflow: simplehttp.SayHelloWorkflow}},
		ClientOptions: client.Options{
			MetricsHandler: sdktally.NewMetricsHandler(newPrometheusScope(prometheus.Configuration{
				ListenAddress: "127.0.0.1:9090",
				TimerType:     "histogram",
			})),
		},
	})
	if err != nil {
		return fmt.Errorf("failed creating handler: %w", err)
	}
	defer h.Close()

	// Disable workflow cache, we want to be stateless
	worker.SetStickyWorkflowCacheSize(0)

	// Serve
	errCh := make(chan error, 1)
	s := &http.Server{Handler: h}
	go func() { errCh <- s.Serve(l) }()
	log.Print("Serving at http://127.0.0.1:9233 (metrics at http://127.0.0.1:9090/metrics)")

	// Stop gracefully on interrupt
	interruptCh := make(chan os.Signal, 1)
	select {
	case err := <-errCh:
		return fmt.Errorf("failed serving: %w", err)
	case <-interruptCh:
		// Shutdown w/ graceful timeout of 30s
		log.Print("Shutting down")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := s.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed shutting down server: %w", err)
		}
	}
	return nil
}

func newPrometheusScope(c prometheus.Configuration) tally.Scope {
	reporter, err := c.NewReporter(
		prometheus.ConfigurationOptions{
			Registry: prom.NewRegistry(),
			OnError: func(err error) {
				log.Println("error in prometheus reporter", err)
			},
		},
	)
	if err != nil {
		log.Fatalln("error creating prometheus reporter", err)
	}
	scopeOpts := tally.ScopeOptions{
		CachedReporter:  reporter,
		Separator:       prometheus.DefaultSeparator,
		SanitizeOptions: &sdktally.PrometheusSanitizeOptions,
		Prefix:          "temporal_samples",
	}
	scope, _ := tally.NewRootScope(scopeOpts, time.Second)
	scope = sdktally.NewPrometheusNamingScope(scope)

	log.Print("Prometheus metrics scope created")
	return scope
}
