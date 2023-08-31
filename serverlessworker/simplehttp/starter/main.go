package main

import (
	"context"
	"fmt"
	"log"

	"github.com/temporalio/samples-go/serverlessworker/simplehttp"
	"go.temporal.io/sdk/client"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	c, err := client.Dial(client.Options{})
	if err != nil {
		return fmt.Errorf("failed creating client: %w", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "serverlessworker-simplehttp_workflowID",
		TaskQueue: "serverlessworker-simplehttp",
	}

	// Just start it then exit
	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, simplehttp.SayHelloWorkflow, "Temporal")
	if err != nil {
		return fmt.Errorf("failed starting workflow: %w", err)
	}
	log.Printf("Started workflow, ID: %v, run ID: %v", we.GetID(), we.GetRunID())
	return nil
}