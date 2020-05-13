package tracer

import (
	"fmt"
	"log"
	"runtime"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/exporters/trace/stdout"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var EnableGlobalTracer = false

type ExporterType int

const (
	Exporter_None ExporterType = iota
	Exporter_Jaeger
	Exporter_Stdout
)

type TracerConfig struct {
	Exporter        ExporterType
	RuntimeProvider bool

	JaegerService string
	JaegerHost    string
}

type Cleanup func()

func InitTracer(cfg TracerConfig) (trace.Provider, Cleanup) {
	var err error
	var provider trace.Provider
	var cleanup Cleanup = func() {}

	switch cfg.Exporter {
	case Exporter_None:
		provider = &trace.NoopProvider{}
	case Exporter_Jaeger:
		provider, cleanup, err = initJaegerProvider(cfg.JaegerHost, cfg.JaegerService)
	case Exporter_Stdout:
		provider, err = initStdoutProvider()
	default:
		err = fmt.Errorf("invalid exporter")
	}

	if err != nil {
		log.Fatalf("failed to init tracer", err)
	}

	if cfg.RuntimeProvider {
		provider = NewRuntimeProvider(provider)
	}

	global.SetTraceProvider(provider)
	return provider, cleanup
}

func initStdoutProvider() (trace.Provider, error) {
	exporter, err := stdout.NewExporter(stdout.Options{PrettyPrint: true})
	if err != nil {
		return nil, err
	}
	return sdktrace.NewProvider(sdktrace.WithConfig(
		sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithSyncer(exporter),
	)
}

func initJaegerProvider(host, service string) (trace.Provider, Cleanup, error) {
	return jaeger.NewExportPipeline(
		jaeger.WithCollectorEndpoint(fmt.Sprintf("http://%s/api/traces", host)),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: service,
			Tags: []kv.KeyValue{
				kv.String("exporter", "jaeger"),
				kv.String("os", runtime.GOOS),
				kv.String("arch", runtime.GOARCH),
				kv.String("go", runtime.Version()),
			},
		}),
		jaeger.WithSDK(&sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
	)
}

func Tracer(name string) trace.Tracer {
	return global.Tracer(name)
}

func Provider() trace.Provider {
	return global.TraceProvider()
}
