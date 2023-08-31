package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/temporalio/samples-go/serverlessworker/awslambda"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	// Disable workflow cache, we want to be stateless
	worker.SetStickyWorkflowCacheSize(0)

	// We want one shared handler because we get client connect reuse benefits if
	// the worker can be reused (there is nothing else shared)
	var handler *OnDemandWorkerHandler

	lambda.StartWithOptions(
		func(ctx context.Context, req *events.LambdaFunctionURLRequest) error {
			// Create the handler only if not already created
			if handler == nil {
				// We expect the client and namespace environment variables to always be
				// present. Real code could have many more settings here.
				address := os.Getenv("TEMPORAL_ADDRESS")
				if address == "" {
					return fmt.Errorf("TEMPORAL_ADDRESS environment variable must be set")
				}
				namespace := os.Getenv("TEMPORAL_NAMESPACE")
				if namespace == "" {
					return fmt.Errorf("TEMPORAL_NAMESPACE environment variable must be set")
				}

				// We want the binary checksum to be the ARN
				lambdaCtx, _ := lambdacontext.FromContext(ctx)
				worker.SetBinaryChecksum(lambdaCtx.InvokedFunctionArn)

				// Create the handler
				var err error
				handler, err = NewOnDemandWorkerHandler(worker.OnDemandWorkflowWorkerOptions{
					TaskQueue:     "serverlessworker-awslambda",
					Workflows:     []worker.RegisteredWorkflow{{Workflow: awslambda.SayHelloWorkflow}},
					ClientOptions: client.Options{HostPort: address, Namespace: namespace},
				})
				if err != nil {
					return fmt.Errorf("failed creating worker handler: %w", err)
				}
			}

			// Run the handler
			return handler.HandleRequest(ctx, req)
		},
		// Close the handler on SIGTERM
		lambda.WithEnableSIGTERM(func() {
			if handler != nil {
				handler.Close()
			}
		}),
	)
}
