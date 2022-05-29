var assert  = require("assert");
var rest = require('./rest/rest.js');

//test GetSupportedCurrencies
testcurrencyservice = new rest.CurrencyService();
console.log(testcurrencyservice.getSupportedCurrencies());

//test Convert
var req = new rest.CurrencyConversionRequest(new rest.Money("EUR", 300, 0), "USD");
var resp = testcurrencyservice.convert(req);
var expect = new rest.Money("USD", 339, 150000000);
assert.deepEqual(resp, expect, "the response is wrong!")