# AWS Lambda-based worker

This example shows a serverless worker powered by AWS Lambda. There are two approaches to test this: ngrok with local
server and Temporal cloud. The way these are done in the samples are intentionally insecure. These steps can be adapted
for more specific setups.

## Create AWS lambda function

Go to the AWS Lambda console and create a new Go 1.x `x86_64` function with function URL enabled in the advanced
settings and `NONE` as the auth type. This is intentionally insecure for this sample, so the function should be deleted
immediately after testing here.

Now follow the AWS Lambda Go zip
[package creation steps](https://docs.aws.amazon.com/lambda/latest/dg/golang-package.html) where `./worker` is the
`go build` target.

Then upload the zip to the function and change the function's runtime settings to set the handler to the name of the
executable that was zipped instead of the default of `hello`.

## Ngrok with local server

This approach is just for testing. [ngrok](https://ngrok.com/) is one of many services that give you a stable URL for a
local server. Signing up and setting up the `ngrok` binary is beyond the scope of this README, see the ngrok docs.

Now, with our function deployed, start a local Temporal server with the task notification URL set as the function URL,
for example:

```
temporal server start-dev --task-poll-expiration 3s --task-notify-url https://bunchofcharactershere.lambda-url.us-east-2.on.aws/
```

With that running, in another terminal, run ngrok to give a stable URL to the local server:

```
ngrok tcp 7233
```

This will give a URL like `tcp://2.tcp.ngrok.io:12345`. Now go to the AWS function and set the following two env vars:

* `TEMPORAL_ADDRESS` as the ngrok `host:port` (i.e. sans the `tcp://` prefix)
* `TEMPORAL_NAMESPACE` as `default`

Now, with the server, ngrok, and our Lambda function running, we can start a workflow in another terminal:

```
go run ./starter
```

If setup correctly, the workflow should run on the lambda instance.

### Simple Benchmark

Now that it's setup, run a lot of them (this will run about 500 workflows):

```
go run ./multistarter -interval 20ms -duration 10s
```

## With Temporal Cloud

TODO

## TODO

Things to do:

* Integrate logging (https://docs.aws.amazon.com/lambda/latest/dg/golang-logging.html)
* Integrate metrics (https://aws.amazon.com/blogs/compute/operating-lambda-logging-and-custom-metrics/)
* Integrate tracing (https://docs.aws.amazon.com/lambda/latest/dg/golang-tracing.html)
* Benchmark large workflow histories being sent signals over and over (causes task turnover, lots of full history
  downloading)