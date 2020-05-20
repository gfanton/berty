package bertytypes

import (
	"context"

	"berty.tech/berty/v2/go/internal/tracer"

	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/api/propagation"
	"go.opentelemetry.io/otel/api/trace"
)

type metadataSupplier struct {
	header *MessageHeaders
}

func (m *metadataSupplier) Get(key string) string {
	if meta := m.header.GetMetadata(); meta != nil {
		if v, ok := meta[key]; ok {
			return v
		}
	}

	return ""
}

func (m *metadataSupplier) Set(key string, val string) {
	if m.header.Metadata == nil {
		m.header.Metadata = make(map[string]string)
	}

	m.header.Metadata[key] = val
}

func (h *MessageHeaders) InjectSpanContext(ctx context.Context) {
	propagation.InjectHTTP(ctx, tracer.Propagators(), &metadataSupplier{h})
}

func (h *MessageHeaders) ExtractSpanContext(ctx context.Context) context.Context {
	return propagation.ExtractHTTP(ctx, tracer.Propagators(), &metadataSupplier{h})
}

func (h *MessageHeaders) SpanFromContext(ctx context.Context, name string, attrs ...kv.KeyValue) (context.Context, trace.Span) {
	hctx := h.ExtractSpanContext(context.Background())
	sctx := trace.RemoteSpanContextFromContext(hctx)
	return tracer.From(ctx).Start(hctx, name, trace.LinkedTo(sctx))
}
