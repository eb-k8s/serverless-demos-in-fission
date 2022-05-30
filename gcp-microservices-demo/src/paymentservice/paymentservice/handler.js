var rest = require('./rest/rest.js');
const pino = require('pino');

const logger = pino({
    name: 'paymentservice',
    messageKey: 'message',
    levelKey: 'severity',
    useLevelLabels: true
});

var paymentservice = new rest.PaymentService();

module.exports = async function(context) {
    if (context.request.method == "POST") {  //Charge
        try {
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
            return {
                status: 200,
                body: resp
            }
        } catch(err) {
            logger.error(err);
            return {
                status: 400
            }
        }
    } else {
        logger.error("methods other than POST are not supported");
        return {
            status: 400
        }
    }
}