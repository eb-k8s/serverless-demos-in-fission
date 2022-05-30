var assert  = require("assert");
var rest = require('./rest/rest.js');

testpaymentservice = new rest.PaymentService();

//test Charge
//expect charge successfully
var req = new rest.ChargeRequest(
    new rest.Money("USD", 100, 0),
    new rest.CreditCardInfo("4432-8015-6152-0454", 123, 2030, 12)
);
assert.notEqual(testpaymentservice.charge(req), null, "the charge is failed!");

//expect throw InvalidCreditCard error when charge
req = new rest.ChargeRequest(
    new rest.Money("USD", 100, 0),
    new rest.CreditCardInfo("4432-8015-6152-0450", 123, 2030, 12)
);
assert.throws(
    () => {testpaymentservice.charge(req)},
    rest.InvalidCreditCard,
    "should throw InvalidCreditCard!"
);

//expect throw UnacceptedCreditCard error when charge
req = new rest.ChargeRequest(
    new rest.Money("USD", 100, 0),
    new rest.CreditCardInfo("378282246310005", 123, 2030, 12)   //this card is amex
);
assert.throws(
    () => {testpaymentservice.charge(req)},
    rest.UnacceptedCreditCard,
    "should throw UnacceptedCreditCard!"
);

//expect throw ExpiredCreditCard error when charge
var req = new rest.ChargeRequest(
    new rest.Money("USD", 100, 0),
    new rest.CreditCardInfo("4432-8015-6152-0454", 123, 2020, 12)
);
assert.throws(
    () => {testpaymentservice.charge(req)},
    rest.ExpiredCreditCard,
    "should throw ExpiredCreditCard!"
);