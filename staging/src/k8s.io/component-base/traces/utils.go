/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package traces

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"k8s.io/klog/v2"
)

// NewProvider initializes tracing in the component, and enforces recommended tracing behavior.
func NewProvider(ctx context.Context, baseSampler sdktrace.Sampler, resourceOpts []resource.Option, opts ...otlpgrpc.Option) trace.TracerProvider {
	opts = append(opts, otlpgrpc.WithInsecure())
	driver := otlpgrpc.NewDriver(opts...)
	exporter, err := otlp.NewExporter(ctx, driver)
	if err != nil {
		klog.Fatalf("Failed to create OTLP exporter: %v", err)
	}

	res, err := resource.New(ctx, resourceOpts...)
	if err != nil {
		klog.Fatalf("Failed to create resource: %v", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter)

	return sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.ParentBased(baseSampler)),
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithResource(res),
	)
}

// Propagators returns the recommended set of propagators.
func Propagators() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
}
