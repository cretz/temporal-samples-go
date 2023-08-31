package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

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

func (o *OnDemandWorkerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// We expect the body to be a proto JSON of capabilities if there is a body
	var capabilities workflowservice.GetSystemInfoResponse_Capabilities
	if body, err := io.ReadAll(r.Body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		// We just put the error on the output, but in a real server this is likely
		// logged and kept internal
		w.Write([]byte(fmt.Sprintf("Failed reading body: %v", err)))
		return
	} else if len(body) > 0 {
		// We don't care about the content type, we try JSON no matter what
		if err := unknownFieldSafeUnmarshaler.Unmarshal(bytes.NewReader(body), &capabilities); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Failed reading body as capabilities: %v", err)))
			return
		}
	}

	// Invoke the on-demand worker
	err := o.worker.ExecuteSingleRun(r.Context(), worker.OnDemandWorkflowWorkerExecuteOptions{
		ServerCapabilities: &capabilities,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Worker failed with error: %v", err)))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (o *OnDemandWorkerHandler) Close() { o.worker.Close() }
