package temporalng

import (
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func (w Workflow[In, Out]) Register(worker worker.Worker, workflowFunc func(Context, In) (Out, error)) {
	if w.Name == "" {
		panic("no workflow name")
	}
	worker.RegisterWorkflowWithOptions(workflowFunc, workflow.RegisterOptions{Name: w.Name})
}
