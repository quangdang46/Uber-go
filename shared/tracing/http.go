package tracing

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func WrapHandlerFunc(handler http.HandlerFunc, operation string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		otelhttp.NewHandler(handler, operation).ServeHTTP(w, r)
	}
}
