package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/gogo/protobuf/jsonpb"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/worker"
)

type OnDemandWorkerHandler struct {
	worker *worker.OnDemandWorkflowWorker
}

func NewOnDemandWorkerHandler(options worker.OnDemandWorkflowWorkerOptions) (*OnDemandWorkerHandler, error) {
	worker, err := worker.NewOnDemandWorkflowWorker(options)
	if err != nil {
		return nil, err
	}
	return &OnDemandWorkerHandler{worker}, nil
}

var unknownFieldSafeUnmarshaler = jsonpb.Unmarshaler{AllowUnknownFields: true}

func (o *OnDemandWorkerHandler) HandleRequest(ctx context.Context, req *events.LambdaFunctionURLRequest) error {
	// We expect the body to be a proto JSON of capabilities if there is a body
	var options worker.OnDemandWorkflowWorkerExecuteOptions
	if req.Body != "" {
		options.ServerCapabilities = &workflowservice.GetSystemInfoResponse_Capabilities{}
		if err := unknownFieldSafeUnmarshaler.Unmarshal(strings.NewReader(req.Body), options.ServerCapabilities); err != nil {
			return fmt.Errorf("failed reading body as capabilities: %w", err)
		}
	}

	// Invoke the on-demand worker
	return o.worker.ExecuteSingleRun(ctx, options)
}

func (o *OnDemandWorkerHandler) Close() { o.worker.Close() }
