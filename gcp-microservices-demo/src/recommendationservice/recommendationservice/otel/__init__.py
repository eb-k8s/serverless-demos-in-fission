from .otel import *
import os

from opentelemetry import trace
import opentelemetry.sdk.resources as resources
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.trace.sampling import ALWAYS_ON
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter

resource = resources.Resource.create({"service.name": "recommendationservice"})

endpoint = os.getenv("OTEL_COLLECTOR_ADDR")
with_otel = False

if endpoint == "":
    print("OTEL_COLLECTOR_ADDR not set, skipping Opentelemtry tracing")
    with_otel = False
    trace.set_tracer_provider(TracerProvider(resource=resource, sampler=ALWAYS_ON))
else:
    print("cartservice with opentelemetry collector: %s\n" % endpoint)
    with_otel = True
    bsp = BatchSpanProcessor(OTLPSpanExporter(endpoint=endpoint, insecure=True))
    trace.set_tracer_provider(TracerProvider(active_span_processor=bsp, resource=resource, sampler=ALWAYS_ON))

tracer = trace.get_tracer("")