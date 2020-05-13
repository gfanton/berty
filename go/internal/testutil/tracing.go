package testutil

import (
	"fmt"
	"os"
	"testing"

	"berty.tech/berty/v2/go/internal/tracer"
	"go.opentelemetry.io/otel/api/trace"
)

func Tracer(t *testing.T, service string) trace.Tracer {
	t.Helper()

	tr_cfg := tracer.TracerConfig{}
	tr_cfg.RuntimeProvider = true

	tracingFlag := os.Getenv("TRACER")
	switch tracingFlag {
	case "":
		tr_cfg.Exporter = tracer.Exporter_None
	case "stdout":
		tr_cfg.Exporter = tracer.Exporter_Stdout
	default:
		tr_cfg.Exporter = tracer.Exporter_Jaeger
		tr_cfg.JaegerHost = tracingFlag
		tr_cfg.JaegerService = "test/" + service
	}

	fmt.Println(tracingFlag)
	pr, cl := tracer.InitTracer(tr_cfg)
	t.Cleanup(cl)
	return pr.Tracer(service)
}
