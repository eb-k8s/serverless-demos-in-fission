var rest = require('./rest/rest.js');
const pino = require('pino');

const logger = pino({
    name: 'paymentservice',
    messageKey: 'message',
    levelKey: 'severity',
    useLevelLabels: true
});

const api = require('@opentelemetry/api')
const { defaultTextMapGetter, ROOT_CONTEXT, SpanKind } = require('@opentelemetry/api');
const tracer = require('./tracer')('paymentservice');

var paymentservice = new rest.PaymentService();

module.exports = async function(context) {
    const parentCtx = api.propagation.extract(ROOT_CONTEXT, context.request.headers, defaultTextMapGetter)
    const span = tracer.startSpan(
        'handle request',
        {
            kind: SpanKind.SERVER,
        },
        parentCtx,
    )

    if (context.request.method == "POST") {  //Charge
        try {
            span.addEvent("invoke Charge");
            var req = new rest.ChargeRequest(
                new rest.Money(
                    context.request.body.amount.currency_code,
                    context.request.body.amount.units,
                    context.request.body.amount.nanos
                ),
                new rest.CreditCardInfo(
                    context.request.body.credit_card.credit_card_number,
                    context.request.body.credit_card.credit_card_cvv,
                    context.request.body.credit_card.credit_card_expiration_year,
                    context.request.body.credit_card.credit_card_expiration_month
                )
            );
            var resp = paymentservice.charge(req);
            span.addEvent("successfully handle request");
            span.end();
            return {
                status: 200,
                body: resp
            }
        } catch(err) {
            logger.error(err);
            span.addEvent("an error occurred in Charge");
            span.end();
            return {
                status: 400
            }
        }
    } else {
        logger.error("methods other than POST are not supported");
        span.addEvent("methods other than POST are not supported");
        span.end();
        return {
            status: 400
        }
    }
}