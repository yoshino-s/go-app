package otelresty

import (
	"net/http/httptest"

	"resty.dev/v3"
)

func ExampleTraceClient() {
	cli := resty.New()
	opts := []Option{
		WithTracerName("my-tracer"),
		WithHideURL(true),
	}
	server := httptest.NewServer(testHandler())

	// this hook is executed before the hook added by `TraceClient`
	cli.AddRequestMiddleware(func(c *resty.Client, r *resty.Request) error {
		r.Header.Add("x-custom-header", "value")
		return nil
	})

	TraceClient(cli, opts...)

	cli.R().Get(server.URL)
}
