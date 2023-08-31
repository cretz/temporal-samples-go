package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/temporalio/samples-go/serverlessworker/awslambda"
	"go.temporal.io/sdk/client"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Accept interval and duration flags
	var interval, duration time.Duration
	flag.DurationVar(&interval, "interval", 0, "Interval between workflow starts")
	flag.DurationVar(&duration, "duration", 0, "Total duration to run")
	flag.Parse()
	if interval == 0 || duration == 0 {
		return fmt.Errorf("-interval and -duration required")
	}

	// Create client
	c, err := client.Dial(client.Options{})
	if err != nil {
		return fmt.Errorf("failed creating client: %w", err)
	}
	defer c.Close()

	// Count iterations to log at the end
	var iterations int
	defer func(start time.Time) { log.Printf("Started %v workflows in %v", iterations, time.Since(start)) }(time.Now())

	// Start workflow every interval
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	errCh := make(chan error, 1)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	doneCh := time.After(duration)
	for {
		select {
		case <-sigCh:
			log.Print("Signal received, exiting")
			return nil
		case err := <-errCh:
			return err
		case <-doneCh:
			// For this simple example, we're ok not waiting and just cancelling
			// outstanding work knowing that iterations can technically be off
			return nil
		case <-ticker.C:
			iterations++
			if iterations%50 == 0 {
				log.Printf("Started %v workflows", iterations)
			}
			go func(iteration int) {
				_, err := c.ExecuteWorkflow(
					ctx,
					client.StartWorkflowOptions{
						ID:        "serverlessworker-awslambda_workflowID_" + strconv.Itoa(iteration),
						TaskQueue: "serverlessworker-awslambda",
					},
					awslambda.SayHelloWorkflow, "Temporal",
				)
				if err != nil {
					// Non-blocking send attempt
					select {
					case errCh <- err:
					default:
					}
				}
			}(iterations)
		}
	}
}
