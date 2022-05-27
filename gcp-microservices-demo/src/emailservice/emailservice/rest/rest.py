import json

class CartItem:
    def __init__(self, product_id, quantity):
        self.product_id = product_id
        self.quantity = quantity

    def toDict(self):
        return {"product_id": self.product_id, "quantity": self.quantity}

def dict2CartItem(dic):
    product_id, quantity = None, None
    if "product_id" in dic:
        product_id = dic["product_id"]
    if "quantity" in dic:
        quantity = dic["quantity"]
    return CartItem(product_id, quantity)

class Money:
    def __init__(self, currency_code, units, nanos):
        self.currency_code = currency_code
        self.units = units
        self.nanos = nanos
    
    def toDict(self):
        return {"currency_code": self.currency_code, "units": self.units, "nanos": self.nanos}

def dict2Money(dic):
    currency_code, units, nanos = None, None, None
    if "currency_code" in dic:
        currency_code = dic["currency_code"]
    if "units" in dic:
        units = dic["units"]
    if "nanos" in dic:
        nanos = dic["nanos"]
    return Money(currency_code, units, nanos)

class OrderItem:
    def __init__(self, item, cost):
        self.item = item
        self.cost = cost
    
    def toDict(self):
        return {"item": self.item.toDict(), "cost": self.cost.toDict()}

def dict2OrderItem(dic):
    item, cost = None, None
    if "item" in dic:
        item = dict2CartItem(dic["item"])
    if "cost" in dic:
        cost = dict2Money(dic["cost"])
    return OrderItem(item, cost)

class Address:
    def __init__(self, street_address, city, state, country, zip_code):
        self.street_address = street_address
        self.city = city
        self.state = state
        self.country = country
        self.zip_code = zip_code
    
    def toDict(self):
        return {"street_address": self.street_address, "city": self.city, 
        "state": self.state, "country": self.country, "zip_code": self.zip_code}

def dict2Address(dic):
    street_address, city, state, country, zip_code = None, None, None, None, None
    if "street_address" in dic:
        street_address = dic["street_address"]
    if "city" in dic:
        city = dic["city"]
    if "state" in dic:
        state = dic["state"]
    if "country" in dic:
        country = dic["country"]
    if "zip_code" in dic:
        zip_code = dic["zip_code"]  
    return Address(street_address, city, state, country, zip_code)

class OrderResult:
    def __init__(self, order_id, shipping_tracking_id, shipping_cost, shipping_address, items):
        self.order_id = order_id
        self.shipping_tracking_id = shipping_tracking_id
        self.shipping_cost = shipping_cost
        self.shipping_address = shipping_address
        self.items = items
    
    def toDict(self):
        itemsArr = []
        for item in self.items:
            itemsArr.append(item.toDict())
        return {
            "order_id": self.order_id,
            "shipping_tracking_id": self.shipping_tracking_id,
            "shipping_cost": self.shipping_cost.toDict(),
            "shipping_address": self.shipping_address.toDict(),
            "items": itemsArr
        }

def dict2OrderResult(dic):
    order_id, shipping_tracking_id, shipping_cost, shipping_address, items = None, None, None, None, []
    if "order_id" in dic:
        order_id = dic["order_id"]
    if "shipping_tracking_id" in dic:
        shipping_tracking_id = dic["shipping_tracking_id"]
    if "shipping_cost" in dic:
        shipping_cost = dict2Money(dic["shipping_cost"])
    if "shipping_address" in dic:
        shipping_address = dict2Address(dic["shipping_address"])
    if "items" in dic:
        itemsArr = dic["items"]
        for item in itemsArr:
            items.append(dict2OrderItem(item))
    return OrderResult(order_id, shipping_tracking_id, shipping_cost, shipping_address, items)

class SendOrderConfirmationRequest:
    def __init__(self, email, order):
        self.email = email
        self.order = order
    
    def toDict(self):
        return {"email": self.email, "order": self.order.toDict()}

def dict2SendOrderConfirmationRequest(dic):
    email, order = None, None
    if "email" in dic:
        email = dic["email"]
    if "order" in dic:
        order = dict2OrderResult(dic["order"])
    return SendOrderConfirmationRequest(email, order)

class DummyEmailService:
    def __init__(self):
        return
    
    def sendOrderConfirmation(self, sendOrderConfirmationRequest):
        print('A request to send order confirmation email to {} has been received.'.format(sendOrderConfirmationRequest.email))
        return