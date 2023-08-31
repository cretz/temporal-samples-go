package awslambda

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

func SayHelloWorkflow(ctx workflow.Context, name string) (string, error) {
	workflow.GetLogger(ctx).Info("Waiting 5 seconds, then returning")
	if err := workflow.Sleep(ctx, 5*time.Second); err != nil {
		return "", err
	}
	return fmt.Sprintf("Hello, %v!", name), nil
}
