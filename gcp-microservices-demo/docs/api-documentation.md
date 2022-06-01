# API Documentation
Microservices in this demo use **http/json** to communicate with each other.

## Service
| Path | Method | Request | Response | API_name | Function | 
| ---- | ------ | ------- | -------- | -------- | -------- |
| /cart | POST | AddItemRequest | \<empty\> | AddItem | cartservice |
| /cart | GET | GetCartRequest | Cart | GetCart | cartservice |
| /cart | DELETE | EmptyCartRequest | \<empty\> | EmptyCart | cartservice |
| /recommendation | GET | ListRecommendationsRequest | ListRecommendationsResponse | ListRecommendations | recommendationservice |
| /product | GET | \<empty\> | ListProductsResponse | ListProducts | productcatalogservice |
| /product | GET | GetProductRequest | Product | GetProduct | productcatalogservice |
| /product | GET | SearchProductsRequest | SearchProductsResponse | SearchProducts | productcatalogservice |
| /shipping | POST | GetQuoteRequest | GetQuoteResponse | GetQuote | shippingService |
| /shipping | PUT | ShipOrderRequest | ShipOrderResponse | ShipOrder | shippingService |
| /currency | GET | \<empty\> | GetSupportedCurrenciesResponse | GetSupportedCurrencies | currencyservice |
| /currency | POST | CurrencyConversionRequest | Money | Convert | currencyservice |
| /payment | POST | ChargeRequest | ChargeResponse | Charge | paymentservice |
| /email | POST | SendOrderConfirmationRequest | \<empty\> | SendOrderConfirmation | emailservice |
| /checkout | POST | PlaceOrderRequest | PlaceOrderResponse | PlaceOrder | checkoutservice |
| /ad | GET | AdRequest | AdResponse | GetAds | adservice |

## Message
<table>
    <tr>
        <th> message_name </th>
        <th> field_name </th>
        <th> field_type </th>
    </tr>
    <tr>
        <td rowspan="2"> AddItemRequest </td>
        <td> user_id </td>
        <td> String </td>
    </tr>
    <tr>
        <td> item </td>
        <td> CartItem </td>
    </tr>
    <tr>
        <td rowspan="2"> CartItem </td>
        <td> product_id </td>
        <td> String </td>
    </tr>
    <tr>
        <td> quantity </td>
        <td> Integer </td>
    </tr>
    <tr>
        <td> GetCartRequest </td>
        <td> user_id </td>
        <td> String </td>
    </tr>
    <tr>
        <td rowspan="2"> Cart </td>
        <td> user_id </td>
        <td> String </td>
    </tr>
    <tr>
        <td> items </td>
        <td> CartItem[] </td>
    </tr>
    <tr>
        <td> EmptyCartRequest </td>
        <td> user_id </td>
        <td> String </td>
    </tr>
    <tr>
        <td rowspan="2"> ListRecommendationsRequest </td>
        <td> user_id </td>
        <td> String </td>
    </tr>
    <tr>
        <td> product_ids </td>
        <td> String[] (separated by ',' when method is GET) </td>
    </tr>
    <tr>
        <td> ListRecommendationsResponse </td>
        <td> product_ids </td>
        <td> String[] </td>
    </tr>
    <tr>
        <td> ListProductsResponse </td>
        <td> products </td>
        <td> Product[] </td>
    </tr>
    <tr>
        <td rowspan="6"> Product </td>
        <td> id </td>
        <td> String </td>
    </tr>
    <tr>
        <td> name </td>
        <td> String </td>
    </tr>
    <tr>
        <td> description </td>
        <td> String </td>
    </tr>
    <tr>
        <td> picture </td>
        <td> String </td>
    </tr>
    <tr>
        <td> price_usd </td>
        <td> Money </td>
    </tr>
    <tr>
        <td> categories </td>
        <td> String[] </td>
    </tr>
    <tr>
        <td rowspan="3"> Money </td>
        <td> currency_code </td>
        <td> String </td>
    </tr>
    <tr>
        <td> units </td>
        <td> Integer </td>
    </tr>
    <tr>
        <td> nanos </td>
        <td> Integer </td>
    </tr>
    <tr>
        <td> GetProductRequest </td>
        <td> id </td>
        <td> String </td>
    </tr>
    <tr>
        <td> SearchProductsRequest </td>
        <td> query </td>
        <td> String </td>
    </tr>
    <tr>
        <td> SearchProductsResponse </td>
        <td> results </td>
        <td> Product[] </td>
    </tr>
    <tr>
        <td rowspan="2"> GetQuoteRequest </td>
        <td> address </td>
        <td> Address </td>
    </tr>
    <tr>
        <td> items </td>
        <td> CartItem[] </td>
    </tr>
    <tr>
        <td rowspan="5"> Address </td>
        <td> street_address </td>
        <td> String </td>
    </tr>
    <tr>
        <td> city </td>
        <td> String </td>
    </tr>
    <tr>
        <td> state </td>
        <td> String </td>
    </tr>
    <tr>
        <td> country </td>
        <td> String </td>
    </tr>
    <tr>
        <td> zip_code </td>
        <td> Integer </td>
    </tr>
    <tr>
        <td> GetQuoteResponse </td>
        <td> cost_usd </td>
        <td> Money </td>
    </tr>
    <tr>
        <td rowspan="2"> ShipOrderRequest </td>
        <td> address </td>
        <td> Address </td>
    </tr>
    <tr>
        <td> items </td>
        <td> CartItem[] </td>
    </tr>
    <tr>
        <td> ShipOrderResponse </td>
        <td> tracking_id </td>
        <td> String </td>
    </tr>
    <tr>
        <td> GetSupportedCurrenciesResponse </td>
        <td> currency_codes </td>
        <td> String[] </td>
    </tr>
    <tr>
        <td rowspan="2"> CurrencyConversionRequest </td>
        <td> from </td>
        <td> Money </td>
    </tr>
    <tr>
        <td> to_code </td>
        <td> String </td>
    </tr>
    <tr>
        <td rowspan="2"> ChargeRequest </td>
        <td> amount </td>
        <td> Money </td>
    </tr>
    <tr>
        <td> credit_card </td>
        <td> CreditCardInfo </td>
    </tr>
    <tr>
        <td rowspan="4"> CreditCardInfo </td>
        <td> credit_card_number </td>
        <td> String </td>
    </tr>
    <tr>
        <td> credit_card_cvv </td>
        <td> Integer </td>
    </tr>
    <tr>
        <td> credit_card_expiration_year </td>
        <td> Integer </td>
    </tr>
    <tr>
        <td> credit_card_expiration_month </td>
        <td> Integer </td>
    </tr>
    <tr>
        <td> ChargeResponse </td>
        <td> transaction_id </td>
        <td> String </td>
    </tr>
    <tr>
        <td rowspan="2"> SendOrderConfirmationRequest </td>
        <td> email </td>
        <td> String </td>
    </tr>
    <tr>
        <td> order </td>
        <td> OrderResult </td>
    </tr>
    <tr>
        <td rowspan="5"> OrderResult </td>
        <td> order_id </td>
        <td> String </td>
    </tr>
    <tr>
        <td> shipping_tracking_id </td>
        <td> String </td>
    </tr>
    <tr>
        <td> shipping_cost </td>
        <td> Money </td>
    </tr>
    <tr>
        <td> shipping_address </td>
        <td> Address </td>
    </tr>
    <tr>
        <td> items </td>
        <td> OrderItem[] </td>
    </tr>
    <tr>
        <td rowspan="2"> OrderItem </td>
        <td> item </td>
        <td> CartItem </td>
    </tr>
    <tr>
        <td> cost </td>
        <td> Money </td>
    </tr>
    <tr>
        <td rowspan="5"> PlaceOrderRequest </td>
        <td> user_id </td>
        <td> String </td>
    </tr>
    <tr>
        <td> user_currency </td>
        <td> String </td>
    </tr>
    <tr>
        <td> address </td>
        <td> Address </td>
    </tr>
    <tr>
        <td> email </td>
        <td> String </td>
    </tr>
    <tr>
        <td> credit_card </td>
        <td> CreditCardInfo </td>
    </tr>
    <tr>
        <td> PlaceOrderResponse </td>
        <td> order </td>
        <td> OrderResult </td>
    </tr>
    <tr>
        <td> AdRequest </td>
        <td> context_keys </td>
        <td> String[] (separated by ',' when method is GET) </td>
    </tr>
    <tr>
        <td> AdResponse </td>
        <td> ads </td>
        <td> Ad[] </td>
    </tr>
    <tr>
        <td rowspan="2"> Ad </td>
        <td> redirect_url </td>
        <td> String </td>
    </tr>
    <tr>
        <td> text </td>
        <td> String </td>
    </tr>
</table>