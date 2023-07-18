package temporalng

import "go.temporal.io/sdk/workflow"

type Context struct{ ctx workflow.Context }

func ConvertFromWorkflowContext(ctx workflow.Context) Context {
	return Context{ctx}
}

func (c Context) Done() ReceiveChannel[struct{}] {
	return ConvertFromWorkflowReceiveChannel[struct{}](c.ctx.Done())
}

func (c Context) Err() error {
	return c.ctx.Err()
}

func (c Context) Value(key any) any {
	return c.ctx.Value(key)
}

func (c Context) ConvertToWorkflowContext() workflow.Context {
	return c.ctx
}

type Channel[T any] struct {
	SendChannel[T]
	ReceiveChannel[T]
}

func NewChannel[T any](buffered int) Channel[T] {
	panic("TODO")
}

func ConvertFromWorkflowChannel[T any](ch workflow.Channel) Channel[T] {
	return Channel[T]{
		SendChannel:    ConvertFromWorkflowSendChannel[T](ch),
		ReceiveChannel: ConvertFromWorkflowReceiveChannel[T](ch),
	}
}

func (c Channel[T]) ConvertToWorkflowChannel() workflow.Channel {
	return c.SendChannel.ch.(workflow.Channel)
}

type ReceiveChannel[T any] struct{ ch workflow.ReceiveChannel }

func ConvertFromWorkflowReceiveChannel[T any](ch workflow.ReceiveChannel) ReceiveChannel[T] {
	return ReceiveChannel[T]{ch}
}

func (r ReceiveChannel[T]) ConvertToWorkflowReceiveChannel() workflow.ReceiveChannel {
	return r.ch
}

func (r ReceiveChannel[T]) Receive(ctx Context) T {
	t, _ := r.TryReceive(ctx)
	return t
}

func (r ReceiveChannel[T]) TryReceive(ctx Context) (T, bool) {
	var t T
	ok := r.ch.Receive(ctx.ctx, &t)
	return t, ok
}

func (r ReceiveChannel[T]) Len() int {
	return r.ch.Len()
}

type SendChannel[T any] struct{ ch workflow.SendChannel }

func ConvertFromWorkflowSendChannel[T any](ch workflow.SendChannel) SendChannel[T] {
	return SendChannel[T]{ch}
}

func (s SendChannel[T]) ConvertToWorkflowSendChannel() workflow.SendChannel {
	return s.ch
}

func (s SendChannel[T]) Send(ctx Context, value T) {
	s.ch.Send(ctx.ctx, value)
}

func (s SendChannel[T]) Close() {
	s.ch.Close()
}

type Future[T any] struct{ fut workflow.Future }

func NewFuture[T any](ctx Context) (Future[T], FutureResolver[T]) {
	fut, set := workflow.NewFuture(ctx.ctx)
	return ConvertFromWorkflowFuture[T](fut), ConvertFromWorkflowSettable[T](set)
}

func ConvertFromWorkflowFuture[T any](fut workflow.Future) Future[T] {
	return Future[T]{fut}
}

func (f Future[T]) ConvertToWorkflowFuture() workflow.Future {
	return f.fut
}

func (f Future[T]) Get(ctx Context) (T, error) {
	var t T
	err := f.fut.Get(ctx.ctx, &t)
	return t, err
}

func (f Future[T]) IsReady() bool {
	return f.fut.IsReady()
}

type FutureResolver[T any] struct{ set workflow.Settable }

func ConvertFromWorkflowSettable[T any](set workflow.Settable) FutureResolver[T] {
	return FutureResolver[T]{set}
}

func (f FutureResolver[T]) ConvertToWorkflowSettable() workflow.Settable {
	return f.set
}

func (f FutureResolver[T]) Resolve(value T) {
	f.set.SetValue(value)
}

func (f FutureResolver[T]) ResolveError(err error) {
	f.set.SetError(err)
}

type SelectCase interface{ apply(*selector) }

type selector struct {
	workflow.Selector
	ctx Context
	err error
}

type selectCase func(*selector)

func (s selectCase) apply(sel *selector) { s(sel) }

func Select(ctx Context, cases ...SelectCase) error {
	sel := &selector{Selector: workflow.NewSelector(ctx.ctx), ctx: ctx}
	for _, c := range cases {
		c.apply(sel)
	}
	sel.Select(ctx.ctx)
	return sel.err
}

func SelectCaseReceive[T any](ch ReceiveChannel[T], callback func(T) error) SelectCase {
	return selectCase(func(sel *selector) {
		sel.AddReceive(ch.ConvertToWorkflowReceiveChannel(), func(workflow.ReceiveChannel, bool) {
			sel.err = callback(ch.Receive(sel.ctx))
		})
	})
}

func SelectCaseTryReceive[T any](ch ReceiveChannel[T], callback func(T, bool) error) SelectCase {
	return selectCase(func(sel *selector) {
		sel.AddReceive(ch.ConvertToWorkflowReceiveChannel(), func(workflow.ReceiveChannel, bool) {
			sel.err = callback(ch.TryReceive(sel.ctx))
		})
	})
}

func SelectCaseSend[T any](ch SendChannel[T], value T, callback func() error) SelectCase {
	return selectCase(func(sel *selector) {
		sel.AddSend(ch.ConvertToWorkflowSendChannel(), value, func() {
			sel.err = callback()
		})
	})
}

func SelectCaseFuture[T any](fut Future[T], callback func() error) SelectCase {
	return selectCase(func(sel *selector) {
		sel.AddFuture(fut.ConvertToWorkflowFuture(), func(workflow.Future) {
			sel.err = callback()
		})
	})
}

func SelectCaseDefault(callback func() error) SelectCase {
	return selectCase(func(sel *selector) {
		sel.AddDefault(func() {
			sel.err = callback()
		})
	})
}
