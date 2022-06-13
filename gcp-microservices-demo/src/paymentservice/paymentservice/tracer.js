const api = require('@opentelemetry/api');
const grpc = require('@grpc/grpc-js');
const { BasicTracerProvider, BatchSpanProcessor } = require('@opentelemetry/sdk-trace-base');
const { OTLPTraceExporter } = require('@opentelemetry/exporter-trace-otlp-grpc');
const { Resource } = require('@opentelemetry/resources');
const { SemanticResourceAttributes } = require('@opentelemetry/semantic-conventions');
const { W3CTraceContextPropagator } = require('@opentelemetry/core');
const pino = require('pino');

const logger = pino({
    name: 'paymentservice',
    messageKey: 'message',
    levelKey: 'severity',
    useLevelLabels: true
});

var endpoint = process.env.OTEL_COLLECTOR_ADDR

module.exports = (serviceName) => {
    const provider = new BasicTracerProvider({
        resource: new Resource({
            [SemanticResourceAttributes.SERVICE_NAME]: serviceName,
        })
    });
    if (endpoint == "") {
        logger.info("OTEL_COLLECTOR_ADDR not set, skipping Opentelemtry tracing")
        provider.register();
    } else {
        logger.info("cartservice with opentelemetry collector: %s", endpoint);
        const collectorOptions = {
            url: endpoint,
            credentials: grpc.credentials.createInsecure(),
        };
        const exporter = new OTLPTraceExporter(collectorOptions);
        provider.addSpanProcessor(new BatchSpanProcessor(exporter));
        provider.register();
    }
    api.propagation.setGlobalPropagator(new W3CTraceContextPropagator())
    return provider.getTracer('');
}