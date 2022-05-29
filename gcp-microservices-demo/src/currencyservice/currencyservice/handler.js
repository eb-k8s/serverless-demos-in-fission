var rest = require('./rest/rest.js');
const pino = require('pino');

const logger = pino({
    name: 'currencyservice',
    messageKey: 'message',
    levelKey: 'severity',
    useLevelLabels: true
});

var currencyservice = new rest.CurrencyService();

module.exports = async function(context) {
    if (context.request.method == "GET") {  //GetSupportedCurrencies
        try {
            var resp = currencyservice.getSupportedCurrencies();
            return {
                status: 200,
                body: resp
            }
        } catch(err) {
            logger.error(err);
            return {
                status: 400,
                body: new rest.GetSupportedCurrenciesResponse()
            }
        }
    } else if (context.request.method == "POST") {    //Convert
        try {
            var req = new rest.CurrencyConversionRequest(
                new rest.Money(
                    context.request.body.from.currency_code,
                    context.request.body.from.units,
                    context.request.body.from.nanos
                ),
                context.request.body.to_code
            );
            var resp = currencyservice.convert(req);
            return {
                status: 200,
                body: resp
            }
        } catch(err) {
            logger.error(err);
            return {
                status: 400,
                body: new rest.Money()
            }
        }
    } else {
        logger.error("methods other than GET and POST are not supported");
        return {
            status: 400
        }
    }
}