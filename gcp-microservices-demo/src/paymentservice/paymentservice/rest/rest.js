const cardValidator = require('simple-card-validator');
const uuid = require('uuid/v4');
const pino = require('pino');

const logger = pino({
    name: 'paymentservice-charge',
    messageKey: 'message',
    levelKey: 'severity',
    useLevelLabels: true
});

class Money {
    constructor(currency_code, units, nanos) {
        this.currency_code = currency_code;
        this.units = units;
        this.nanos = nanos;
    }
}

class CreditCardInfo {
    constructor(credit_card_number, credit_card_cvv, credit_card_expiration_year, credit_card_expiration_month) {
        this.credit_card_number = credit_card_number;
        this.credit_card_cvv = credit_card_cvv;
        this.credit_card_expiration_year = credit_card_expiration_year;
        this.credit_card_expiration_month = credit_card_expiration_month;
    }
}

class ChargeRequest {
    constructor(amount, credit_card) {
        this.amount = amount;
        this.credit_card = credit_card;
    }
}

class ChargeResponse {
    constructor(transaction_id) {
        this.transaction_id = transaction_id;
    }
}

class CreditCardError extends Error {
    constructor (message) {
        super(message);
        this.code = 400; // Invalid argument error
    }
}
  
class InvalidCreditCard extends CreditCardError {
    constructor (cardType) {
        super(`Credit card info is invalid`);
    }
}
  
class UnacceptedCreditCard extends CreditCardError {
    constructor (cardType) {
        super(`Sorry, we cannot process ${cardType} credit cards. Only VISA or MasterCard is accepted.`);
    }
}
  
class ExpiredCreditCard extends CreditCardError {
    constructor (number, month, year) {
        super(`Your credit card (ending ${number.substr(-4)}) expired on ${month}/${year}`);
    }
}

class PaymentService {
    /**
    * Verifies the credit card number and (pretend) charges the card.
    */
    charge (chargeRequest) {
        logger.info("charge...");
        const { amount: amount, credit_card: creditCard } = chargeRequest;
        const cardNumber = creditCard.credit_card_number;
        const cardInfo = cardValidator(cardNumber);
        const {
            card_type: cardType,
            valid
        } = cardInfo.getCardDetails();
  
        if (!valid) { throw new InvalidCreditCard(); }
  
        // Only VISA and mastercard is accepted, other card types (AMEX, dinersclub) will
        // throw UnacceptedCreditCard error.
        if (!(cardType === 'visa' || cardType === 'mastercard')) { throw new UnacceptedCreditCard(cardType); }
  
        // Also validate expiration is > today.
        const currentMonth = new Date().getMonth() + 1;
        const currentYear = new Date().getFullYear();
        const { credit_card_expiration_year: year, credit_card_expiration_month: month } = creditCard;
        if ((currentYear * 12 + currentMonth) > (year * 12 + month)) { throw new ExpiredCreditCard(cardNumber.replace('-', ''), month, year); }
  
        logger.info(`Transaction processed: ${cardType} ending ${cardNumber.substr(-4)}\
        Amount: ${amount.currency_code}${amount.units}.${amount.nanos}`);
  
        return new ChargeResponse(uuid());
    };
}

module.exports = {
    Money: Money,
    CreditCardInfo: CreditCardInfo,
    ChargeRequest: ChargeRequest,
    ChargeResponse: ChargeResponse,
    PaymentService: PaymentService,
    InvalidCreditCard: InvalidCreditCard,
    UnacceptedCreditCard: UnacceptedCreditCard,
    ExpiredCreditCard: ExpiredCreditCard
}