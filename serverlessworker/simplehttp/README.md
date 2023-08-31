# HTTP-based serverless worker

This example shows serverless worker powered by HTTP.

## Manually stepping

First, start a Temporal server with 3s task poll expiration. For example using Temporal CLI:

```
temporal server start-dev --task-poll-expiration 3s
```

Now, start a workflow (this will not wait for workflow to complete):

```
go run ./starter
```

When visiting the UI, the workflow has started but not progressed. Now, run the serverless worker (which is an HTTP
server):

```
go run ./worker
```

Now that worker is running, we can issue single-run requests by hitting http://127.0.0.1:9233. For example, with curl
run:

```
curl http://127.0.0.1:9233
```

This will have progressed the workflow one stage. Note, this HTTP call takes 3s to complete because that's the poll
expiration we set. The UI shows the 5s timer but the serverless worker request didn't run long enough (3s) to get timer
completion (5s). If the timer were shorter or the poll expiration longer, the same serverless worker request would have
processed timer completion too.

To complete the workflow, do another single run:

```
curl http://127.0.0.1:9233
```

Now viewing the UI will show the workflow as completed.

## Automatic server HTTP notification

For this use case, the server will be invoking the HTTP handler, so we need to start the worker first:

```
go run ./worker
```

Now start the server with the URL to our worker HTTP server:

```
temporal server start-dev --task-poll-expiration 3s --task-notify-url http://127.0.0.1:9233
```

With that running, in another terminal, start the workflow:

```
go run ./starter
```

This will run the workflow to completion. It is clear from the logs that this starts two workers because the poll
expiration is less than the sleep. Change the expiration to 10s:

```
temporal server start-dev --task-poll-expiration 10s --task-notify-url http://127.0.0.1:9233
```

With that running, in another terminal, start the workflow:

```
go run ./starter
```

Now it will be clear from the logs that only one worker started/stopped.

## Simple benchmark

This is the same as the automatic server step above, but in this case we will be using a starter to run several
workflows.

Start the worker first:

```
go run ./worker
```

Now start the server with the URL to our worker HTTP server:

```
temporal server start-dev --task-poll-expiration 3s --task-notify-url http://127.0.0.1:9233
```

With that running, in another terminal, start a workflow approximately every 20ms for 10s:

```
go run ./multistarter -interval 20ms -duration 10s
```

All the workflows will get started and complete regularly after their sleeps.