var rest = require('./rest/rest.js');
const pino = require('pino');

const logger = pino({
    name: 'currencyservice',
    messageKey: 'message',
    levelKey: 'severity',
    useLevelLabels: true
});

const api = require('@opentelemetry/api')
const { defaultTextMapGetter, ROOT_CONTEXT, SpanKind } = require('@opentelemetry/api');
const tracer = require('./tracer')('currencyservice');

var currencyservice = new rest.CurrencyService();

module.exports = async function(context) {
    const parentCtx = api.propagation.extract(ROOT_CONTEXT, context.request.headers, defaultTextMapGetter)
    const span = tracer.startSpan(
        'handle request',
        {
            kind: SpanKind.SERVER,
        },
        parentCtx,
    )

    if (context.request.method == "GET") {  //GetSupportedCurrencies
        try {
            span.addEvent("invoke GetSupportedCurrencies");
            var resp = currencyservice.getSupportedCurrencies();
            span.addEvent("successfully handle request");
            span.end();
            return {
                status: 200,
                body: resp
            }
        } catch(err) {
            logger.error(err);
            span.addEvent("an error occurred in GetSupportedCurrencies");
            span.end();
            return {
                status: 400,
                body: new rest.GetSupportedCurrenciesResponse()
            }
        }
    } else if (context.request.method == "POST") {    //Convert
        try {
            span.addEvent("invoke Convert");
            var req = new rest.CurrencyConversionRequest(
                new rest.Money(
                    context.request.body.from.currency_code,
                    context.request.body.from.units,
                    context.request.body.from.nanos
                ),
                context.request.body.to_code
            );
            var resp = currencyservice.convert(req);
            span.addEvent("successfully handle request");
            span.end();
            return {
                status: 200,
                body: resp
            }
        } catch(err) {
            logger.error(err);
            span.addEvent("an error occurred in Convert");
            span.end();
            return {
                status: 400,
                body: new rest.Money()
            }
        }
    } else {
        logger.error("methods other than GET and POST are not supported");
        span.addEvent("methods other than GET and POST are not supported");
        span.end();
        return {
            status: 400
        }
    }
}