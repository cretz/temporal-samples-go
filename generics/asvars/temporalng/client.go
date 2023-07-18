package temporalng

import (
	"context"
	"fmt"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/api/query/v1"
	workflowpb "go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
)

func (w Workflow[In, Out]) Start(
	ctx context.Context,
	cl client.Client,
	arg In,
	options client.StartWorkflowOptions,
) (*WorkflowHandle[Out], error) {
	run, err := cl.ExecuteWorkflow(ctx, options, w.Name, arg)
	if err != nil {
		return nil, err
	}
	return &WorkflowHandle[Out]{
		WorkflowRun: WorkflowRun{
			Client:              cl,
			ID:                  run.GetID(),
			FirstExecutionRunID: run.GetRunID(),
		},
		ResultRunID: run.GetRunID(),
	}, nil
}

func (w Workflow[In, Out]) Execute(
	ctx context.Context,
	cl client.Client,
	arg In,
	options client.StartWorkflowOptions,
) (Out, error) {
	var out Out
	h, err := w.Start(ctx, cl, arg, options)
	if err == nil {
		out, err = h.GetResult(ctx, WorkflowResultOptions{})
	}
	return out, err
}

type SignalWithStartOptions struct {
	StartOptions client.StartWorkflowOptions
	SignalName   string
	SignalArg    any
}

func (w Workflow[In, Out]) SignalWithStart(
	ctx context.Context,
	cl client.Client,
	arg In,
	options SignalWithStartOptions,
) (*WorkflowHandle[Out], error) {
	run, err := cl.SignalWithStartWorkflow(
		ctx, options.StartOptions.ID, options.SignalName, options.SignalArg, options.StartOptions, w.Name, arg)
	if err != nil {
		return nil, err
	}
	return &WorkflowHandle[Out]{
		WorkflowRun: WorkflowRun{
			Client:              cl,
			ID:                  run.GetID(),
			FirstExecutionRunID: run.GetRunID(),
		},
		ResultRunID: run.GetRunID(),
	}, nil
}

type WorkflowHandleOptions struct {
	ID                  string
	RunID               string
	FirstExecutionRunID string
}

func (w Workflow[In, Out]) GetHandle(cl client.Client, options WorkflowHandleOptions) *WorkflowHandle[Out] {
	return &WorkflowHandle[Out]{
		WorkflowRun: WorkflowRun{
			Client:              cl,
			ID:                  options.ID,
			RunID:               options.RunID,
			FirstExecutionRunID: options.FirstExecutionRunID,
		},
		ResultRunID: options.RunID,
	}
}

type WorkflowRun struct {
	Client              client.Client
	ID                  string
	RunID               string
	FirstExecutionRunID string
}

type WorkflowHandle[Out any] struct {
	WorkflowRun
	ResultRunID string
}

type WorkflowResultOptions struct {
	DisableFollowingRuns bool
}

func (w *WorkflowHandle[Out]) GetResult(ctx context.Context, options WorkflowResultOptions) (Out, error) {
	run := w.Client.GetWorkflow(ctx, w.ID, w.ResultRunID)
	var out Out
	err := run.GetWithOptions(ctx, &out, client.WorkflowRunGetOptions{DisableFollowingRuns: options.DisableFollowingRuns})
	return out, err
}

type workflowRun interface {
	GetClient() client.Client
	GetID() string
	GetRunID() string
	GetFirstExecutionRunID() string
}

var _ workflowRun = WorkflowRun{}

func (w WorkflowRun) GetClient() client.Client {
	return w.Client
}

func (w WorkflowRun) GetID() string {
	return w.ID
}

func (w WorkflowRun) GetRunID() string {
	return w.RunID
}

func (w WorkflowRun) GetFirstExecutionRunID() string {
	return w.RunID
}

type WorkflowCancelOptions struct {
}

func (w WorkflowRun) Cancel(ctx context.Context, options WorkflowCancelOptions) error {
	// TODO(cretz): Support first execution run ID if/when Go SDK does
	return w.Client.CancelWorkflow(ctx, w.ID, w.RunID)
}

type WorkflowTerminateOptions struct {
	Reason  string
	Details []any
}

func (w WorkflowRun) Terminate(ctx context.Context, options WorkflowTerminateOptions) error {
	// TODO(cretz): Support first execution run ID if/when Go SDK does
	return w.Client.TerminateWorkflow(ctx, w.ID, w.RunID, options.Reason, options.Details...)
}

type WorkflowHistoryOptions struct {
}

func (w WorkflowRun) FetchHistory(ctx context.Context, options WorkflowHistoryOptions) (*history.History, error) {
	iter := w.Client.GetWorkflowHistory(ctx, w.ID, w.RunID, false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	var hist history.History
	for iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			return nil, err
		}
		hist.Events = append(hist.Events, event)
	}
	return &hist, nil
}

type WorkflowExecution struct {
	Raw *workflowpb.WorkflowExecutionInfo
	// TODO(cretz): Add common fields here
}

type WorkflowExecutionDescription struct {
	WorkflowExecution
	Raw *workflowservice.DescribeWorkflowExecutionResponse
	// TODO(cretz): Add common fields here
}

type WorkflowDescribeOptions struct {
}

func (w WorkflowRun) Describe(
	ctx context.Context,
	options WorkflowDescribeOptions,
) (*WorkflowExecutionDescription, error) {
	desc, err := w.Client.DescribeWorkflowExecution(ctx, w.ID, w.RunID)
	if err != nil {
		return nil, err
	}
	return &WorkflowExecutionDescription{
		WorkflowExecution: WorkflowExecution{Raw: desc.WorkflowExecutionInfo},
		Raw:               desc,
	}, nil
}

type WorkflowSignalOptions struct {
}

func (w WorkflowSignal[In]) Send(ctx context.Context, workflow workflowRun, arg In, options WorkflowSignalOptions) error {
	return workflow.GetClient().SignalWorkflow(ctx, workflow.GetID(), workflow.GetRunID(), w.Name, arg)
}

type WorkflowQueryOptions struct {
	RejectCondition enums.QueryRejectCondition
}

type ErrWorkflowQueryRejected struct{ query.QueryRejected }

func (e *ErrWorkflowQueryRejected) Error() string {
	return fmt.Sprintf("query rejected, workflow status: %v", e.Status)
}

func (w WorkflowQuery[In, Out]) Execute(
	ctx context.Context,
	workflow workflowRun,
	arg In,
	options WorkflowQueryOptions,
) (Out, error) {
	var out Out
	resp, err := workflow.GetClient().QueryWorkflowWithOptions(ctx, &client.QueryWorkflowWithOptionsRequest{
		WorkflowID:           workflow.GetID(),
		RunID:                workflow.GetRunID(),
		QueryType:            w.Name,
		Args:                 []any{arg},
		QueryRejectCondition: options.RejectCondition,
	})
	if err == nil {
		if resp.QueryRejected != nil {
			err = &ErrWorkflowQueryRejected{QueryRejected: *resp.QueryRejected}
		} else {
			err = resp.QueryResult.Get(&out)
		}
	}
	return out, err
}

type WorkflowUpdateOptions struct {
	ID string
}

type WorkflowUpdateHandle[Out any] struct {
	ID            string
	WorkflowID    string
	WorkflowRunID string

	underlying client.WorkflowUpdateHandle
}

func (w WorkflowUpdate[In, Out]) Start(
	ctx context.Context,
	workflow workflowRun,
	arg In,
	options WorkflowUpdateOptions,
) (*WorkflowUpdateHandle[Out], error) {
	h, err := workflow.GetClient().UpdateWorkflowWithOptions(ctx, &client.UpdateWorkflowWithOptionsRequest{
		UpdateID:            options.ID,
		WorkflowID:          workflow.GetID(),
		RunID:               workflow.GetRunID(),
		UpdateName:          w.Name,
		Args:                []any{arg},
		FirstExecutionRunID: workflow.GetFirstExecutionRunID(),
	})
	if err != nil {
		return nil, err
	}
	return &WorkflowUpdateHandle[Out]{
		ID:            h.UpdateID(),
		WorkflowID:    h.WorkflowID(),
		WorkflowRunID: h.RunID(),
		underlying:    h,
	}, nil
}

func (w WorkflowUpdate[In, Out]) Execute(
	ctx context.Context,
	workflow workflowRun,
	arg In,
	options WorkflowUpdateOptions,
) (Out, error) {
	var out Out
	h, err := w.Start(ctx, workflow, arg, options)
	if err == nil {
		out, err = h.GetResult(ctx, WorkflowUpdateResultOptions{})
	}
	return out, err
}

type WorkflowUpdateResultOptions struct {
}

func (w *WorkflowUpdateHandle[Out]) GetResult(ctx context.Context, options WorkflowUpdateResultOptions) (Out, error) {
	var out Out
	var err error
	if w.underlying == nil {
		// TODO(cretz): Remove this requirement when we support async update result fetching
		err = fmt.Errorf("must have obtained handle from update start")
	} else {
		err = w.underlying.Get(ctx, &out)
	}
	return out, err
}

type WorkflowListOptions struct {
	Query string
}

type WorkflowExecutionIterator struct {
	ctx       context.Context
	client    client.Client
	req       *workflowservice.ListWorkflowExecutionsRequest
	lastResp  *workflowservice.ListWorkflowExecutionsResponse
	currIndex int
	lastErr   error
}

func ListWorkflows(
	ctx context.Context,
	cl client.Client,
	options WorkflowListOptions,
) *WorkflowExecutionIterator {
	return &WorkflowExecutionIterator{
		ctx:    ctx,
		client: cl,
		req:    &workflowservice.ListWorkflowExecutionsRequest{Query: options.Query},
	}
}

func (w *WorkflowExecutionIterator) HasNext() bool {
	// Fetch page (in a loop because token may come back with empty list)
	for {
		// If there's an error or more in the execution set, return true. If there
		// is a response but no next page token and didn't match above, return
		// false.
		if w.lastErr != nil || w.currIndex < len(w.lastResp.GetExecutions()) {
			return true
		} else if w.lastResp != nil && len(w.lastResp.NextPageToken) == 0 {
			return false
		}
		// Fetch next page
		w.currIndex = 0
		w.req.NextPageToken = w.lastResp.GetNextPageToken()
		w.lastResp, w.lastErr = w.client.ListWorkflow(w.ctx, w.req)
	}
}

func (w *WorkflowExecutionIterator) Next() (*WorkflowExecution, error) {
	if !w.HasNext() {
		panic("Next called without checking HasNext")
	} else if w.lastErr != nil {
		return nil, w.lastErr
	}
	w.currIndex++
	return &WorkflowExecution{Raw: w.lastResp.Executions[w.currIndex-1]}, nil
}
