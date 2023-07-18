package temporalng

import (
	"context"

	"go.temporal.io/sdk/worker"
)

type Workflow[In, Out any] struct {
	Name string
}

type Activity[In, Out any] struct {
	Name string
}

func (*Activity[In, Out]) Register(worker.Worker, func(context.Context, In) (Out, error)) {
	panic("TODO")
}

type ActivityExecuteOptions struct {
	// ...
}

func (*Activity[In, Out]) Execute(Context, In, ActivityExecuteOptions) *Future[Out] {
	panic("TODO")
}

type WorkflowSignal[In any] struct {
	Name string
}

func (WorkflowSignal[In]) GetChannel(Context) *ReceiveChannel[In] {
	panic("TODO")
}

type WorkflowQuery[In, Out any] struct {
	Name string
}

func (WorkflowQuery[In, Out]) SetHandler(Context, func(In) (Out, error)) {
	panic("TODO")
}

type WorkflowUpdate[In, Out any] struct {
	Name string
}
