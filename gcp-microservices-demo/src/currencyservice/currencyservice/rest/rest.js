const pino = require('pino');

const logger = pino({
    name: 'currencyservice',
    messageKey: 'message',
    levelKey: 'severity',
    useLevelLabels: true
});

class GetSupportedCurrenciesResponse {
    constructor(currency_codes) {
        this.currency_codes = currency_codes
    }
}

class CurrencyConversionRequest {
    constructor(from, to_code) {
        this.from = from
        this.to_code = to_code
    }
}

class Money {
    constructor(currency_code, units, nanos) {
        this.currency_code = currency_code
        this.units = units
        this.nanos = nanos
    }
}

/**
 * Helper function that gets currency data from a stored JSON file
 * Uses public data from European Central Bank
 */
function _getCurrencyData () {
    const data = require('./currency_conversion.json');
    return data;
}

/**
 * Helper function that handles decimal/fractional carrying
 */
function _carry (amount) {
    const fractionSize = Math.pow(10, 9);
    amount.nanos += (amount.units % 1) * fractionSize;
    amount.units = Math.floor(amount.units) + Math.floor(amount.nanos / fractionSize);
    amount.nanos = amount.nanos % fractionSize;
    return amount;
}

function _moneyToString (m) {
    return `${m.units}.${m.nanos.toString().padStart(9,'0')} ${m.currency_code}`;
}

class CurrencyService {
    /**
    * Lists the supported currencies
    */
    getSupportedCurrencies () {
        logger.info('Getting supported currencies...');
        try {
            var data = _getCurrencyData();
            var response = new GetSupportedCurrenciesResponse(Object.keys(data));
            logger.info(`Currency codes: ${response.currency_codes}`);
            return response;
        } catch (err) {
            logger.error(`Error in GetSupportedCurrencies: ${err}`);
            return new GetSupportedCurrenciesResponse();
        }
    }

    /**
    * Converts between currencies
    */
    convert (currencyConversionRequest) {
        try {
            logger.info('convert...');
            var data = _getCurrencyData();
            
            // Convert: from_currency --> EUR
            const from = currencyConversionRequest.from;
            const euros = _carry({
                units: from.units / data[from.currency_code],
                nanos: from.nanos / data[from.currency_code]
            });
  
            euros.nanos = Math.round(euros.nanos);
  
            // Convert: EUR --> to_currency
            const result = _carry({
                units: euros.units * data[currencyConversionRequest.to_code],
                nanos: euros.nanos * data[currencyConversionRequest.to_code]
            });

            result.units = Math.floor(result.units);
            result.nanos = Math.floor(result.nanos);
            var resp = new Money(currencyConversionRequest.to_code, result.units, result.nanos);
  
            logger.info(`conversion request successful`);
            logger.info(`Convert: ${_moneyToString(currencyConversionRequest.from)} to ${_moneyToString(resp)}`);
            return resp;
        } catch (err) {
            logger.error(`Error in convert: ${err}`);
            return new Money();
        }
  }
}

module.exports = {
    GetSupportedCurrenciesResponse: GetSupportedCurrenciesResponse,
    CurrencyConversionRequest: CurrencyConversionRequest,
    Money: Money,
    CurrencyService: CurrencyService
}